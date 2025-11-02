package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Quiz represents a quiz for a book
type Quiz struct {
	ID         uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	BookID     uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"book_id"` // Each book has one quiz
	Questions  datatypes.JSON `gorm:"type:jsonb;not null" json:"questions"`
	AIModel    string         `gorm:"default:'gpt-4o-mini'" json:"ai_model"`
	Status     string         `gorm:"default:'completed'" json:"status"` // "completed", "failed", "retrying"
	RetryCount int            `gorm:"default:0" json:"retry_count"`
	ErrorLog   string         `gorm:"type:text" json:"error_log,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	
	// Relationship
	Book       Book           `gorm:"foreignKey:BookID" json:"-"`
}

// TableName specifies the table name for GORM
func (Quiz) TableName() string {
	return "quizzes"
}

// QuizQuestion represents a single quiz question
type QuizQuestion struct {
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	Answer      string   `json:"answer"`
	Explanation string   `json:"explanation"`
}

// QuizData represents the structure of quiz questions in JSONB
type QuizData struct {
	Quiz []QuizQuestion `json:"quiz"`
}

// QuizResponse represents the API response for a quiz
type QuizResponse struct {
	ID        uuid.UUID      `json:"id"`
	BookID    uuid.UUID      `json:"book_id"`
	Questions []QuizQuestion `json:"questions"`
	AIModel   string         `json:"ai_model"`
	CreatedAt time.Time      `json:"created_at"`
}

