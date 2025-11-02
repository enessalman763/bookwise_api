package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bookwise/api/internal/database"
	"github.com/bookwise/api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// QuizHandler handles quiz-related endpoints
type QuizHandler struct{}

// NewQuizHandler creates a new quiz handler
func NewQuizHandler() *QuizHandler {
	return &QuizHandler{}
}

// GetQuiz handles get quiz by book ID
// GET /quiz/:bookId
func (h *QuizHandler) GetQuiz(c *gin.Context) {
	bookIDStr := c.Param("bookId")
	
	bookID, err := uuid.Parse(bookIDStr)
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

	// Check quiz status
	switch book.QuizStatus {
	case "pending":
		c.JSON(http.StatusAccepted, gin.H{
			"success": false,
			"status":  "pending",
			"message": "Quiz henüz oluşturulmadı. Lütfen daha sonra tekrar deneyin.",
		})
		return
	
	case "generating":
		c.JSON(http.StatusAccepted, gin.H{
			"success": false,
			"status":  "generating",
			"message": "Quiz şu anda oluşturuluyor. Lütfen birkaç saniye sonra tekrar deneyin.",
		})
		return
	
	case "failed":
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"status":  "failed",
			"error":   "Quiz oluşturulamadı. Lütfen destek ekibiyle iletişime geçin.",
		})
		return
	}

	// Get quiz
	var quiz models.Quiz
	if err := database.DB.Where("book_id = ?", bookID).First(&quiz).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Quiz bulunamadı",
		})
		return
	}

	// Parse questions - try direct array first, then nested format
	var questions []models.QuizQuestion
	if err := json.Unmarshal(quiz.Questions, &questions); err != nil {
		// Try nested format {"quiz": [...]}
		var quizData models.QuizData
		if err2 := json.Unmarshal(quiz.Questions, &quizData); err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Quiz verisi okunamadı",
			})
			return
		}
		questions = quizData.Quiz
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":         quiz.ID,
			"book_id":    quiz.BookID,
			"quiz":       questions,
			"ai_model":   quiz.AIModel,
			"created_at": quiz.CreatedAt,
		},
	})
}

// GetQuizByID handles get quiz by quiz ID
// GET /quiz/id/:id
func (h *QuizHandler) GetQuizByID(c *gin.Context) {
	quizIDStr := c.Param("id")
	
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Geçersiz quiz ID",
		})
		return
	}

	var quiz models.Quiz
	if err := database.DB.Where("id = ?", quizID).First(&quiz).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Quiz bulunamadı",
		})
		return
	}

	// Parse questions - try direct array first, then nested format
	var questions []models.QuizQuestion
	if err := json.Unmarshal(quiz.Questions, &questions); err != nil {
		// Try nested format {"quiz": [...]}
		var quizData models.QuizData
		if err2 := json.Unmarshal(quiz.Questions, &quizData); err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Quiz verisi okunamadı",
			})
			return
		}
		questions = quizData.Quiz
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":         quiz.ID,
			"book_id":    quiz.BookID,
			"quiz":       questions,
			"ai_model":   quiz.AIModel,
			"created_at": quiz.CreatedAt,
		},
	})
}

