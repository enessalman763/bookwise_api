package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/bookwise/api/config"
	"github.com/bookwise/api/internal/models"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// QuizGeneratorService handles AI quiz generation using Google Gemini
type QuizGeneratorService struct {
	client         *genai.Client
	model          *genai.GenerativeModel
	modelName      string
	questionsCount int
	retryLimit     int
}

// NewQuizGeneratorService creates a new quiz generator service
func NewQuizGeneratorService(cfg *config.Config) *QuizGeneratorService {
	ctx := context.Background()
	
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.Gemini.APIKey))
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	model := client.GenerativeModel(cfg.Gemini.Model)
	
	// Configure model for JSON output
	model.SetTemperature(0.7)
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.ResponseMIMEType = "application/json"
	
	return &QuizGeneratorService{
		client:         client,
		model:          model,
		modelName:      cfg.Gemini.Model,
		questionsCount: cfg.Quiz.QuestionsCount,
		retryLimit:     cfg.Quiz.RetryLimit,
	}
}

// GenerateQuiz generates a quiz for a given book with retry mechanism
func (s *QuizGeneratorService) GenerateQuiz(book *models.Book) (*models.Quiz, error) {
	var lastErr error
	
	for attempt := 1; attempt <= s.retryLimit; attempt++ {
		log.Printf("🤖 Generating quiz for book '%s' (attempt %d/%d)", book.Title, attempt, s.retryLimit)
		
		quiz, err := s.generateQuizAttempt(book)
		if err == nil {
			log.Printf("✅ Quiz generated successfully for '%s'", book.Title)
			return quiz, nil
		}
		
		lastErr = err
		log.Printf("⚠️ Attempt %d failed: %v", attempt, err)
		
		if attempt < s.retryLimit {
			// Wait before retry (exponential backoff)
			waitTime := time.Duration(attempt) * 2 * time.Second
			log.Printf("⏳ Waiting %v before retry...", waitTime)
			time.Sleep(waitTime)
		}
	}
	
	return nil, fmt.Errorf("failed to generate quiz after %d attempts: %w", s.retryLimit, lastErr)
}

// generateQuizAttempt performs a single attempt to generate a quiz
func (s *QuizGeneratorService) generateQuizAttempt(book *models.Book) (*models.Quiz, error) {
	// Create book info for the prompt
	bookInfo := map[string]interface{}{
		"title":          book.Title,
		"authors":        book.Authors,
		"description":    book.Description,
		"categories":     book.Categories,
		"publisher":      book.Publisher,
		"published_date": book.PublishedDate,
	}
	
	bookInfoJSON, _ := json.MarshalIndent(bookInfo, "", "  ")
	
	// Create the prompt
	prompt := fmt.Sprintf(`Kitabın bilgileri:
%s

Bu kitap hakkında %d adet çoktan seçmeli quiz sorusu oluştur.
Sorular kitabın içeriği, teması, yazarı ve önemli noktaları hakkında olmalı.
Her soru için 4 seçenek (A, B, C, D) sun ve doğru cevabı işaretle.
Ayrıca her soru için kısa bir açıklama ekle.

JSON formatında dön (başka bir şey yazma, sadece geçerli JSON):
{
  "quiz": [
    {
      "question": "soru metni",
      "options": ["A) seçenek1", "B) seçenek2", "C) seçenek3", "D) seçenek4"],
      "answer": "doğru cevap (ör: B) seçenek2)",
      "explanation": "açıklama"
    }
  ]
}

ÖNEMLİ: Sadece JSON döndür, başka açıklama ekleme.`, string(bookInfoJSON), s.questionsCount)

	// Call Gemini API
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini api call failed: %w", err)
	}
	
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response from gemini")
	}
	
	if len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini")
	}
	
	// Extract text from response
	content := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	
	// Validate and parse JSON
	var quizData models.QuizData
	if err := json.Unmarshal([]byte(content), &quizData); err != nil {
		return nil, fmt.Errorf("failed to parse quiz JSON: %w. Content: %s", err, content)
	}
	
	// Validate quiz structure
	if len(quizData.Quiz) == 0 {
		return nil, fmt.Errorf("quiz is empty")
	}
	
	for i, q := range quizData.Quiz {
		if q.Question == "" {
			return nil, fmt.Errorf("question %d is empty", i+1)
		}
		if len(q.Options) != 4 {
			return nil, fmt.Errorf("question %d must have exactly 4 options, got %d", i+1, len(q.Options))
		}
		if q.Answer == "" {
			return nil, fmt.Errorf("question %d has no answer", i+1)
		}
	}
	
	// Create quiz model
	questionsJSON, err := json.Marshal(quizData.Quiz)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal questions: %w", err)
	}
	
	quiz := &models.Quiz{
		BookID:     book.ID,
		Questions:  questionsJSON,
		AIModel:    s.modelName,
		Status:     "completed",
		RetryCount: 0,
	}
	
	return quiz, nil
}

// ValidateQuizJSON validates if the quiz JSON is properly formatted
func (s *QuizGeneratorService) ValidateQuizJSON(data []byte) error {
	var quizData models.QuizData
	if err := json.Unmarshal(data, &quizData); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}
	
	if len(quizData.Quiz) == 0 {
		return fmt.Errorf("quiz cannot be empty")
	}
	
	for i, q := range quizData.Quiz {
		if q.Question == "" {
			return fmt.Errorf("question %d: question text is required", i+1)
		}
		if len(q.Options) != 4 {
			return fmt.Errorf("question %d: must have exactly 4 options", i+1)
		}
		if q.Answer == "" {
			return fmt.Errorf("question %d: answer is required", i+1)
		}
		if q.Explanation == "" {
			return fmt.Errorf("question %d: explanation is required", i+1)
		}
	}
	
	return nil
}

// Close closes the Gemini client
func (s *QuizGeneratorService) Close() error {
	return s.client.Close()
}
