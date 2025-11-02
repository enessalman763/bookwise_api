package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bookwise/api/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// BookMergerService handles merging book data from multiple sources
type BookMergerService struct {
	googleBooks *GoogleBooksService
	openLibrary *OpenLibraryService
}

// NewBookMergerService creates a new book merger service
func NewBookMergerService(googleBooksAPIKey string) *BookMergerService {
	return &BookMergerService{
		googleBooks: NewGoogleBooksService(googleBooksAPIKey),
		openLibrary: NewOpenLibraryService(),
	}
}

// SearchBook searches for a book using hybrid sources
func (s *BookMergerService) SearchBook(query, searchType string) (*models.Book, error) {
	var googleData, openLibraryData *BookData
	var err error
	
	sources := []string{}

	// Try Google Books first
	log.Printf("🔍 Searching Google Books for: %s (type: %s)", query, searchType)
	switch searchType {
	case "isbn":
		googleData, err = s.googleBooks.SearchByISBN(query)
	case "title":
		googleData, err = s.googleBooks.SearchByTitle(query)
	case "author":
		googleData, err = s.googleBooks.SearchByAuthor(query)
	default:
		return nil, fmt.Errorf("invalid search type: %s", searchType)
	}

	if err == nil && googleData != nil {
		sources = append(sources, "google_books")
		log.Println("✅ Found in Google Books")
	} else {
		log.Printf("⚠️ Google Books: %v", err)
	}

	// Try Open Library
	log.Printf("🔍 Searching Open Library for: %s (type: %s)", query, searchType)
	switch searchType {
	case "isbn":
		openLibraryData, err = s.openLibrary.SearchByISBN(query)
	case "title":
		openLibraryData, err = s.openLibrary.SearchByTitle(query)
	case "author":
		openLibraryData, err = s.openLibrary.SearchByAuthor(query)
	}

	if err == nil && openLibraryData != nil {
		sources = append(sources, "open_library")
		log.Println("✅ Found in Open Library")
	} else {
		log.Printf("⚠️ Open Library: %v", err)
	}

	// If no data found from either source
	if googleData == nil && openLibraryData == nil {
		return nil, fmt.Errorf("book not found in any source")
	}

	// Merge the data
	log.Println("🔄 Merging book data from sources...")
	mergedBook := s.mergeBookData(googleData, openLibraryData, sources)
	
	return mergedBook, nil
}

// mergeBookData merges book data from multiple sources
// Priority: Google Books > Open Library
func (s *BookMergerService) mergeBookData(googleData, openLibData *BookData, sources []string) *models.Book {
	book := &models.Book{
		ID:          uuid.New(),
		QuizStatus:  "pending",
		DataSources: pq.StringArray(sources),
	}

	// Helper to prefer non-empty values
	preferNonEmpty := func(primary, secondary string) string {
		if primary != "" {
			return primary
		}
		return secondary
	}

	preferNonZero := func(primary, secondary int) int {
		if primary > 0 {
			return primary
		}
		return secondary
	}

	preferNonEmptySlice := func(primary, secondary []string) []string {
		if len(primary) > 0 {
			return primary
		}
		return secondary
	}

	// Merge fields (Google Books has priority)
	if googleData != nil {
		book.Title = googleData.Title
		book.Authors = pq.StringArray(googleData.Authors)
		book.ISBN = googleData.ISBN
		book.ISBN13 = googleData.ISBN13
		book.Description = googleData.Description
		book.Publisher = googleData.Publisher
		book.PublishedDate = googleData.PublishedDate
		book.PageCount = googleData.PageCount
		book.Categories = pq.StringArray(googleData.Categories)
		book.Language = googleData.Language
		book.CoverURL = googleData.CoverURL
		book.ThumbnailURL = googleData.ThumbnailURL
	}

	// Fill missing fields from Open Library
	if openLibData != nil {
		book.Title = preferNonEmpty(book.Title, openLibData.Title)
		
		if len(book.Authors) == 0 {
			book.Authors = pq.StringArray(openLibData.Authors)
		}
		
		book.ISBN = preferNonEmpty(book.ISBN, openLibData.ISBN)
		book.ISBN13 = preferNonEmpty(book.ISBN13, openLibData.ISBN13)
		book.Description = preferNonEmpty(book.Description, openLibData.Description)
		book.Publisher = preferNonEmpty(book.Publisher, openLibData.Publisher)
		book.PublishedDate = preferNonEmpty(book.PublishedDate, openLibData.PublishedDate)
		book.PageCount = preferNonZero(book.PageCount, openLibData.PageCount)
		
		currentCategories := []string(book.Categories)
		book.Categories = pq.StringArray(preferNonEmptySlice(currentCategories, openLibData.Categories))
		
		book.Language = preferNonEmpty(book.Language, openLibData.Language)
		book.CoverURL = preferNonEmpty(book.CoverURL, openLibData.CoverURL)
		book.ThumbnailURL = preferNonEmpty(book.ThumbnailURL, openLibData.ThumbnailURL)
	}

	// Store raw source data for debugging
	sourceData := map[string]interface{}{}
	if googleData != nil {
		sourceData["google_books"] = googleData.RawData
	}
	if openLibData != nil {
		sourceData["open_library"] = openLibData.RawData
	}

	sourceJSON, _ := json.Marshal(sourceData)
	book.SourceData = datatypes.JSON(sourceJSON)

	// Ensure ISBN is set (required field)
	if book.ISBN == "" {
		book.ISBN = book.ISBN13
	}

	log.Printf("✅ Book merged successfully: %s (ISBN: %s) from sources: %v", book.Title, book.ISBN, sources)
	
	return book
}

