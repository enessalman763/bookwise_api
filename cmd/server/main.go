package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bookwise/api/config"
	"github.com/bookwise/api/internal/database"
	"github.com/bookwise/api/internal/handlers"
	"github.com/bookwise/api/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize services
	bookMerger := services.NewBookMergerService(cfg.APIs.GoogleBooksAPIKey)
	quizWorker := services.NewQuizWorker(cfg, 3) // 3 concurrent workers

	// Start quiz worker
	quizWorker.Start()
	defer quizWorker.Stop()

	// Process any pending quizzes on startup
	go quizWorker.ProcessPendingQuizzes()

	// Start periodic retry for failed quizzes (every 1 hour)
	quizWorker.StartPeriodicRetry(1 * time.Hour)

	// Initialize handlers
	booksHandler := handlers.NewBooksHandler(bookMerger, quizWorker)
	quizHandler := handlers.NewQuizHandler()
	healthHandler := handlers.NewHealthHandler(quizWorker)

	// Create router
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Server.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check routes
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/health/detailed", healthHandler.DetailedHealth)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Books routes
		books := v1.Group("/books")
		{
			books.GET("/search", booksHandler.SearchBook)      // GET /api/v1/books/search?q=...&type=...
			books.GET("", booksHandler.ListBooks)              // GET /api/v1/books?page=1&limit=10
			books.GET("/:id", booksHandler.GetBookByID)        // GET /api/v1/books/:id
			books.GET("/isbn/:isbn", booksHandler.GetBookByISBN) // GET /api/v1/books/isbn/:isbn
		}

		// Quiz routes
		quiz := v1.Group("/quiz")
		{
			quiz.GET("/:bookId", quizHandler.GetQuiz)       // GET /api/v1/quiz/:bookId
			quiz.GET("/id/:id", quizHandler.GetQuizByID)    // GET /api/v1/quiz/id/:id
		}
	}

	// Print routes
	log.Println("\n📚 Bookwise API Routes:")
	log.Println("  GET  /health")
	log.Println("  GET  /health/detailed")
	log.Println("  GET  /api/v1/books/search?q={query}&type={isbn|title|author}")
	log.Println("  GET  /api/v1/books")
	log.Println("  GET  /api/v1/books/:id")
	log.Println("  GET  /api/v1/books/isbn/:isbn")
	log.Println("  GET  /api/v1/quiz/:bookId")
	log.Println("  GET  /api/v1/quiz/id/:id")

	// Print worker stats
	quizWorker.PrettyPrintStats()

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("\n🛑 Shutting down gracefully...")
		
		// Stop quiz worker
		quizWorker.Stop()
		
		// Close database connection
		database.CloseDatabase()
		
		log.Println("✅ Shutdown complete")
		os.Exit(0)
	}()

	// Start server
	addr := ":" + cfg.Server.Port
	log.Printf("\n🚀 Server starting on %s\n", addr)
	
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

