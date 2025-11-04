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

// SearchBook handles book search requests - returns list of books
// GET /books/search?q={query}&type={isbn|title|author}&limit={limit}
func (h *BooksHandler) SearchBook(c *gin.Context) {
	query := c.Query("q")
	searchType := c.Query("type")
	limitStr := c.Query("limit")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "query parameter 'q' is required",
		})
		return
	}

	if searchType == "" {
		searchType = "title" // default to title for search
	}

	// Validate search type
	if searchType != "isbn" && searchType != "title" && searchType != "author" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "type must be one of: isbn, title, author",
		})
		return
	}

	// Parse limit
	limit := 10
	if limitStr != "" {
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
			limit = 10
		}
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 40 {
		limit = 40
	}

	log.Printf("🔍 Book search request: query='%s', type='%s', limit=%d", query, searchType, limit)

	// Search from external sources
	books, err := h.bookMerger.SearchBooks(query, searchType, limit)
	if err != nil {
		log.Printf("❌ Book search failed: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Kitap bulunamadı",
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Found %d books", len(books))

	// Convert to response format
	type BookSearchResult struct {
		Title         string   `json:"title"`
		Authors       []string `json:"authors"`
		ISBN          string   `json:"isbn,omitempty"`
		ISBN13        string   `json:"isbn13,omitempty"`
		Description   string   `json:"description,omitempty"`
		Publisher     string   `json:"publisher,omitempty"`
		PublishedDate string   `json:"published_date,omitempty"`
		PageCount     int      `json:"page_count,omitempty"`
		Categories    []string `json:"categories,omitempty"`
		Language      string   `json:"language,omitempty"`
		CoverURL      string   `json:"cover_url,omitempty"`
		ThumbnailURL  string   `json:"thumbnail_url,omitempty"`
		Source        string   `json:"source"`
	}

	results := make([]BookSearchResult, len(books))
	for i, book := range books {
		results[i] = BookSearchResult{
			Title:         book.Title,
			Authors:       book.Authors,
			ISBN:          book.ISBN,
			ISBN13:        book.ISBN13,
			Description:   book.Description,
			Publisher:     book.Publisher,
			PublishedDate: book.PublishedDate,
			PageCount:     book.PageCount,
			Categories:    book.Categories,
			Language:      book.Language,
			CoverURL:      book.CoverURL,
			ThumbnailURL:  book.ThumbnailURL,
			Source:        book.Source,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"count":   len(results),
		"message": fmt.Sprintf("%d kitap bulundu", len(results)),
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

// SaveBook saves a book to the database
// POST /books
// Body: { "isbn": "...", "generate_quiz": true/false }
func (h *BooksHandler) SaveBook(c *gin.Context) {
	type SaveBookRequest struct {
		ISBN         string `json:"isbn" binding:"required"`
		GenerateQuiz bool   `json:"generate_quiz"`
	}

	var req SaveBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ISBN is required",
			"details": err.Error(),
		})
		return
	}

	log.Printf("💾 Save book request: ISBN='%s', generate_quiz=%v", req.ISBN, req.GenerateQuiz)

	// First check if book already exists in database
	var existingBook models.Book
	err := database.DB.Where("isbn = ? OR isbn13 = ?", req.ISBN, req.ISBN).First(&existingBook).Error
	if err == nil {
		log.Printf("ℹ️ Book already exists in database: %s (ISBN: %s)", existingBook.Title, existingBook.ISBN)
		
		// If quiz generation requested and not already generated
		if req.GenerateQuiz && existingBook.QuizStatus == "pending" {
			h.quizWorker.Enqueue(existingBook.ID)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    existingBook.ToResponse(),
				"message": "Kitap zaten kayıtlı. Quiz oluşturuluyor...",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    existingBook.ToResponse(),
			"message": "Kitap zaten kayıtlı",
		})
		return
	}

	// Fetch book details from external sources
	book, err := h.bookMerger.SearchBook(req.ISBN, "isbn")
	if err != nil {
		log.Printf("❌ Book not found: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Kitap bulunamadı",
			"details": err.Error(),
		})
		return
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

	// Trigger quiz generation if requested
	message := "Kitap başarıyla kaydedildi"
	if req.GenerateQuiz {
		h.quizWorker.Enqueue(book.ID)
		message = "Kitap başarıyla kaydedildi. Quiz oluşturuluyor..."
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    book.ToResponse(),
		"message": message,
	})
}

// GenerateQuiz generates quiz for a specific book
// POST /books/:id/generate-quiz
func (h *BooksHandler) GenerateQuiz(c *gin.Context) {
	idStr := c.Param("id")
	
	bookID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Geçersiz kitap ID",
		})
		return
	}

	// Check if book exists
	var book models.Book
	if err := database.DB.Where("id = ?", bookID).First(&book).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Kitap bulunamadı",
		})
		return
	}

	log.Printf("🎯 Generate quiz request for book: %s (ID: %s, Status: %s)", book.Title, book.ID, book.QuizStatus)

	// Check quiz status
	if book.QuizStatus == "completed" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Quiz zaten oluşturulmuş. Yeni quiz oluşturulsun mu?",
			"status":  "completed",
		})
		return
	}

	if book.QuizStatus == "generating" {
		c.JSON(http.StatusAccepted, gin.H{
			"success": false,
			"message": "Quiz şu anda oluşturuluyor. Lütfen bekleyin.",
			"status":  "generating",
		})
		return
	}

	// Trigger quiz generation
	h.quizWorker.Enqueue(book.ID)

	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"message": "Quiz oluşturma işlemi başlatıldı. Lütfen birkaç saniye sonra kontrol edin.",
		"status":  "generating",
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

