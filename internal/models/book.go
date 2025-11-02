package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// Book represents a book in the system
type Book struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Title         string         `gorm:"not null" json:"title"`
	Authors       pq.StringArray `gorm:"type:text[]" json:"authors"`
	ISBN          string         `gorm:"uniqueIndex;not null" json:"isbn"`
	ISBN13        string         `json:"isbn13,omitempty"`
	Description   string         `gorm:"type:text" json:"description,omitempty"`
	Publisher     string         `json:"publisher,omitempty"`
	PublishedDate string         `json:"published_date,omitempty"`
	PageCount     int            `json:"page_count,omitempty"`
	Categories    pq.StringArray `gorm:"type:text[]" json:"categories,omitempty"`
	Language      string         `json:"language,omitempty"`
	CoverURL      string         `json:"cover_url,omitempty"`
	ThumbnailURL  string         `json:"thumbnail_url,omitempty"`
	SourceData    datatypes.JSON `gorm:"type:jsonb" json:"source_data,omitempty"`      // Raw data for debugging
	DataSources   pq.StringArray `gorm:"type:text[]" json:"data_sources,omitempty"`    // ["google_books", "open_library"]
	QuizID        *uuid.UUID     `gorm:"type:uuid" json:"quiz_id,omitempty"`
	QuizStatus    string         `gorm:"default:'pending'" json:"quiz_status"`         // "pending", "generating", "completed", "failed"
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Book) TableName() string {
	return "books"
}

// BookResponse represents the API response for a book
type BookResponse struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Authors       []string  `json:"authors"`
	ISBN          string    `json:"isbn"`
	ISBN13        string    `json:"isbn13,omitempty"`
	Description   string    `json:"description,omitempty"`
	Publisher     string    `json:"publisher,omitempty"`
	PublishedDate string    `json:"published_date,omitempty"`
	PageCount     int       `json:"page_count,omitempty"`
	Categories    []string  `json:"categories,omitempty"`
	Language      string    `json:"language,omitempty"`
	CoverURL      string    `json:"cover_url,omitempty"`
	ThumbnailURL  string    `json:"thumbnail_url,omitempty"`
	DataSources   []string  `json:"data_sources,omitempty"`
	QuizStatus    string    `json:"quiz_status"`
	QuizID        *uuid.UUID `json:"quiz_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToResponse converts Book model to BookResponse
func (b *Book) ToResponse() *BookResponse {
	return &BookResponse{
		ID:            b.ID,
		Title:         b.Title,
		Authors:       []string(b.Authors),
		ISBN:          b.ISBN,
		ISBN13:        b.ISBN13,
		Description:   b.Description,
		Publisher:     b.Publisher,
		PublishedDate: b.PublishedDate,
		PageCount:     b.PageCount,
		Categories:    []string(b.Categories),
		Language:      b.Language,
		CoverURL:      b.CoverURL,
		ThumbnailURL:  b.ThumbnailURL,
		DataSources:   []string(b.DataSources),
		QuizStatus:    b.QuizStatus,
		QuizID:        b.QuizID,
		CreatedAt:     b.CreatedAt,
	}
}

