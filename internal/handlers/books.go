package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bookwise/api/internal/database"
	"github.com/bookwise/api/internal/models"
	"github.com/bookwise/api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BooksHandler handles book-related endpoints
type BooksHandler struct {
	bookMerger *services.BookMergerService
	quizWorker *services.QuizWorker
}

// NewBooksHandler creates a new books handler
func NewBooksHandler(bookMerger *services.BookMergerService, quizWorker *services.QuizWorker) *BooksHandler {
	return &BooksHandler{
		bookMerger: bookMerger,
		quizWorker: quizWorker,
	}
}

// SearchBook handles book search requests
// GET /books/search?q={query}&type={isbn|title|author}
func (h *BooksHandler) SearchBook(c *gin.Context) {
	query := c.Query("q")
	searchType := c.Query("type")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "query parameter 'q' is required",
		})
		return
	}

	if searchType == "" {
		searchType = "isbn" // default to ISBN
	}

	// Validate search type
	if searchType != "isbn" && searchType != "title" && searchType != "author" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "type must be one of: isbn, title, author",
		})
		return
	}

	log.Printf("🔍 Book search request: query='%s', type='%s'", query, searchType)

	// First, check if book exists in cache (database)
	var existingBook models.Book
	var cacheHit bool

	if searchType == "isbn" {
		// For ISBN, we can do exact match
		err := database.DB.Where("isbn = ? OR isbn13 = ?", query, query).First(&existingBook).Error
		if err == nil {
			cacheHit = true
			log.Printf("✅ Cache hit for ISBN: %s", query)
			
			c.JSON(http.StatusOK, gin.H{
				"success":   true,
				"data":      existingBook.ToResponse(),
				"cache_hit": true,
				"message":   "Kitap önbellekten getirildi",
			})
			return
		}
	}

	// Cache miss, fetch from external sources
	log.Println("⚠️ Cache miss, fetching from external sources...")
	
	book, err := h.bookMerger.SearchBook(query, searchType)
	if err != nil {
		log.Printf("❌ Book search failed: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Kitap bulunamadı",
			"details": err.Error(),
		})
		return
	}

	// Check if book already exists by ISBN (to prevent duplicate key error)
	var existingBookByISBN models.Book
	if book.ISBN != "" {
		err := database.DB.Where("isbn = ? OR isbn13 = ?", book.ISBN, book.ISBN).First(&existingBookByISBN).Error
		if err == nil {
			// Book already exists, return it
			log.Printf("ℹ️ Book already exists in database: %s (ISBN: %s)", existingBookByISBN.Title, existingBookByISBN.ISBN)
			c.JSON(http.StatusOK, gin.H{
				"success":   true,
				"data":      existingBookByISBN.ToResponse(),
				"cache_hit": true,
				"message":   "Kitap veritabanında zaten mevcut",
			})
			return
		}
	}

	// Save to database
	if err := database.DB.Create(book).Error; err != nil {
		log.Printf("❌ Failed to save book to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Kitap kaydedilemedi",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Book saved to database: %s (ID: %s)", book.Title, book.ID)

	// Trigger quiz generation asynchronously
	h.quizWorker.Enqueue(book.ID)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"data":      book.ToResponse(),
		"cache_hit": cacheHit,
		"message":   "Kitap başarıyla getirildi. Quiz oluşturuluyor...",
	})
}

// GetBookByID handles get book by UUID
// GET /books/:id
func (h *BooksHandler) GetBookByID(c *gin.Context) {
	idStr := c.Param("id")
	
	bookID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Geçersiz kitap ID",
		})
		return
	}

	var book models.Book
	if err := database.DB.Where("id = ?", bookID).First(&book).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Kitap bulunamadı",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    book.ToResponse(),
	})
}

// GetBookByISBN handles get book by ISBN
// GET /books/isbn/:isbn
func (h *BooksHandler) GetBookByISBN(c *gin.Context) {
	isbn := c.Param("isbn")

	var book models.Book
	if err := database.DB.Where("isbn = ? OR isbn13 = ?", isbn, isbn).First(&book).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Kitap bulunamadı",
			"message": "Bu ISBN ile kayıtlı kitap bulunamadı. /books/search?q=" + isbn + "&type=isbn ile arama yapabilirsiniz.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    book.ToResponse(),
	})
}

// ListBooks handles listing all books with pagination
// GET /books?page=1&limit=10
func (h *BooksHandler) ListBooks(c *gin.Context) {
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil {
			page = 1
		}
	}

	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 10
		}
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	var books []models.Book
	var total int64

	database.DB.Model(&models.Book{}).Count(&total)
	database.DB.Offset(offset).Limit(limit).Order("created_at DESC").Find(&books)

	responses := make([]*models.BookResponse, len(books))
	for i, book := range books {
		responses[i] = book.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

