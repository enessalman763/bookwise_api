package services

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/bookwise/api/config"
	"github.com/bookwise/api/internal/database"
	"github.com/bookwise/api/internal/models"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// QuizWorker handles background quiz generation
type QuizWorker struct {
	generator  *QuizGeneratorService
	queue      chan uuid.UUID
	wg         sync.WaitGroup
	workerCount int
	running    bool
	mu         sync.Mutex
}

// NewQuizWorker creates a new quiz worker
func NewQuizWorker(cfg *config.Config, workerCount int) *QuizWorker {
	return &QuizWorker{
		generator:   NewQuizGeneratorService(cfg),
		queue:       make(chan uuid.UUID, 100),
		workerCount: workerCount,
		running:     false,
	}
}

// Start starts the worker pool
func (w *QuizWorker) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.mu.Unlock()

	log.Printf("🚀 Starting quiz worker pool with %d workers", w.workerCount)

	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.worker(i + 1)
	}
}

// Stop stops the worker pool
func (w *QuizWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	log.Println("🛑 Stopping quiz worker pool...")
	close(w.queue)
	w.wg.Wait()
	w.running = false
	log.Println("✅ Quiz worker pool stopped")
}

// Enqueue adds a book ID to the quiz generation queue
func (w *QuizWorker) Enqueue(bookID uuid.UUID) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		log.Println("⚠️ Worker not running, cannot enqueue")
		return
	}

	select {
	case w.queue <- bookID:
		log.Printf("📝 Book %s added to quiz generation queue", bookID)
	default:
		log.Printf("⚠️ Quiz queue is full, skipping book %s", bookID)
	}
}

// worker processes quiz generation jobs
func (w *QuizWorker) worker(id int) {
	defer w.wg.Done()

	log.Printf("👷 Worker #%d started", id)

	for bookID := range w.queue {
		log.Printf("👷 Worker #%d processing book %s", id, bookID)
		w.processQuizGeneration(bookID)
	}

	log.Printf("👷 Worker #%d stopped", id)
}

// processQuizGeneration generates a quiz for a book
func (w *QuizWorker) processQuizGeneration(bookID uuid.UUID) {
	// Get book from database
	var book models.Book
	if err := database.DB.Where("id = ?", bookID).First(&book).Error; err != nil {
		log.Printf("❌ Failed to get book %s: %v", bookID, err)
		return
	}

	// Check if quiz already exists and is completed
	var existingQuiz models.Quiz
	if err := database.DB.Where("book_id = ? AND status = ?", bookID, "completed").First(&existingQuiz).Error; err == nil {
		log.Printf("ℹ️ Quiz already exists for book '%s', skipping", book.Title)
		
		// Update book quiz status
		database.DB.Model(&book).Updates(map[string]interface{}{
			"quiz_id":     existingQuiz.ID,
			"quiz_status": "completed",
		})
		return
	}
	
	// Delete failed quiz if exists
	database.DB.Where("book_id = ? AND status = ?", bookID, "failed").Delete(&models.Quiz{})

	// Update book status to "generating"
	database.DB.Model(&book).Update("quiz_status", "generating")

	// Generate quiz
	quiz, err := w.generator.GenerateQuiz(&book)
	if err != nil {
		log.Printf("❌ Failed to generate quiz for book '%s': %v", book.Title, err)
		
		// Update book status to "failed"
		database.DB.Model(&book).Updates(map[string]interface{}{
			"quiz_status": "failed",
		})
		
	// Create failed quiz record for tracking
	failedQuiz := &models.Quiz{
		BookID:     bookID,
		Questions:  datatypes.JSON([]byte(`{"quiz":[]}`)),
		AIModel:    w.generator.modelName,
		Status:     "failed",
		RetryCount: w.generator.retryLimit,
		ErrorLog:   err.Error(),
	}
		database.DB.Create(failedQuiz)
		return
	}

	// Save quiz to database
	if err := database.DB.Create(quiz).Error; err != nil {
		log.Printf("❌ Failed to save quiz to database: %v", err)
		
		database.DB.Model(&book).Update("quiz_status", "failed")
		return
	}

	// Update book with quiz ID and status
	database.DB.Model(&book).Updates(map[string]interface{}{
		"quiz_id":     quiz.ID,
		"quiz_status": "completed",
	})

	log.Printf("✅ Quiz generated and saved for book '%s' (quiz_id: %s)", book.Title, quiz.ID)
}

// ProcessPendingQuizzes processes all books with pending quiz status
func (w *QuizWorker) ProcessPendingQuizzes() {
	var books []models.Book
	if err := database.DB.Where("quiz_status IN ?", []string{"pending", "failed"}).Find(&books).Error; err != nil {
		log.Printf("❌ Failed to get pending books: %v", err)
		return
	}

	if len(books) == 0 {
		log.Println("ℹ️ No pending quizzes to process")
		return
	}

	log.Printf("📚 Found %d books with pending quizzes", len(books))

	for _, book := range books {
		w.Enqueue(book.ID)
	}
}

// RetryFailedQuizzes retries quiz generation for failed books
func (w *QuizWorker) RetryFailedQuizzes() {
	var books []models.Book
	if err := database.DB.Where("quiz_status = ?", "failed").Find(&books).Error; err != nil {
		log.Printf("❌ Failed to get failed books: %v", err)
		return
	}

	if len(books) == 0 {
		log.Println("ℹ️ No failed quizzes to retry")
		return
	}

	log.Printf("🔄 Retrying %d failed quizzes", len(books))

	for _, book := range books {
		// Reset status to pending
		database.DB.Model(&book).Update("quiz_status", "pending")
		w.Enqueue(book.ID)
	}
}

// StartPeriodicRetry starts a periodic retry mechanism for failed quizzes
func (w *QuizWorker) StartPeriodicRetry(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("⏰ Periodic retry check triggered")
			w.RetryFailedQuizzes()
		}
	}()

	log.Printf("⏰ Periodic retry started (interval: %v)", interval)
}

// GetQueueSize returns the current queue size
func (w *QuizWorker) GetQueueSize() int {
	return len(w.queue)
}

// GetStats returns worker statistics
func (w *QuizWorker) GetStats() map[string]interface{} {
	var total, pending, generating, completed, failed int64

	database.DB.Model(&models.Book{}).Count(&total)
	database.DB.Model(&models.Book{}).Where("quiz_status = ?", "pending").Count(&pending)
	database.DB.Model(&models.Book{}).Where("quiz_status = ?", "generating").Count(&generating)
	database.DB.Model(&models.Book{}).Where("quiz_status = ?", "completed").Count(&completed)
	database.DB.Model(&models.Book{}).Where("quiz_status = ?", "failed").Count(&failed)

	return map[string]interface{}{
		"total_books":    total,
		"pending":        pending,
		"generating":     generating,
		"completed":      completed,
		"failed":         failed,
		"queue_size":     w.GetQueueSize(),
		"worker_count":   w.workerCount,
		"worker_running": w.running,
	}
}

// PrettyPrintStats prints worker statistics in a formatted way
func (w *QuizWorker) PrettyPrintStats() {
	stats := w.GetStats()
	statsJSON, _ := json.MarshalIndent(stats, "", "  ")
	log.Printf("📊 Quiz Worker Stats:\n%s", string(statsJSON))
}

