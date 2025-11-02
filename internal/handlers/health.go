package handlers

import (
	"net/http"
	"time"

	"github.com/bookwise/api/internal/database"
	"github.com/bookwise/api/internal/services"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	quizWorker *services.QuizWorker
	startTime  time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(quizWorker *services.QuizWorker) *HealthHandler {
	return &HealthHandler{
		quizWorker: quizWorker,
		startTime:  time.Now(),
	}
}

// HealthCheck handles basic health check
// GET /health
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "bookwise-api",
		"time":    time.Now(),
	})
}

// DetailedHealth handles detailed health check with system status
// GET /health/detailed
func (h *HealthHandler) DetailedHealth(c *gin.Context) {
	// Check database connection
	sqlDB, err := database.DB.DB()
	dbStatus := "healthy"
	if err != nil {
		dbStatus = "unhealthy"
	} else {
		if err := sqlDB.Ping(); err != nil {
			dbStatus = "unhealthy"
		}
	}

	// Get quiz worker stats
	workerStats := h.quizWorker.GetStats()

	uptime := time.Since(h.startTime)

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "bookwise-api",
		"uptime":  uptime.String(),
		"components": gin.H{
			"database":    dbStatus,
			"quiz_worker": workerStats,
		},
		"timestamp": time.Now(),
	})
}

