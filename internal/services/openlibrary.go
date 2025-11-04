package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OpenLibraryService handles Open Library API integration
type OpenLibraryService struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewOpenLibraryService creates a new Open Library service
func NewOpenLibraryService() *OpenLibraryService {
	return &OpenLibraryService{
		BaseURL: "https://openlibrary.org",
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// OpenLibraryBookResponse represents the response from Open Library
type OpenLibraryBookResponse struct {
	Key               string                    `json:"key"`
	Title             string                    `json:"title"`
	Authors           []OpenLibraryAuthorRef    `json:"authors"`
	Publishers        []string                  `json:"publishers"`
	PublishDate       string                    `json:"publish_date"`
	NumberOfPages     int                       `json:"number_of_pages"`
	ISBN10            []string                  `json:"isbn_10"`
	ISBN13            []string                  `json:"isbn_13"`
	Subjects          []string                  `json:"subjects"`
	Languages         []OpenLibraryLanguageRef  `json:"languages"`
	Description       interface{}               `json:"description"` // Can be string or object
	Covers            []int                     `json:"covers"`
}

// OpenLibraryAuthorRef represents an author reference
type OpenLibraryAuthorRef struct {
	Key string `json:"key"`
}

// OpenLibraryLanguageRef represents a language reference
type OpenLibraryLanguageRef struct {
	Key string `json:"key"`
}

// OpenLibrarySearchResponse represents search results
type OpenLibrarySearchResponse struct {
	NumFound int                         `json:"numFound"`
	Docs     []OpenLibrarySearchDoc      `json:"docs"`
}

// OpenLibrarySearchDoc represents a search result document
type OpenLibrarySearchDoc struct {
	Key              string   `json:"key"`
	Title            string   `json:"title"`
	AuthorName       []string `json:"author_name"`
	ISBN             []string `json:"isbn"`
	Publisher        []string `json:"publisher"`
	PublishYear      []int    `json:"publish_year"`
	NumberOfPagesMedian int   `json:"number_of_pages_median"`
	Subject          []string `json:"subject"`
	Language         []string `json:"language"`
	CoverI           int      `json:"cover_i"`
}

// SearchByISBN searches for a book by ISBN (returns single result)
func (s *OpenLibraryService) SearchByISBN(isbn string) (*BookData, error) {
	// Clean ISBN (remove hyphens)
	cleanISBN := strings.ReplaceAll(isbn, "-", "")
	
	// Try ISBN API first
	reqURL := fmt.Sprintf("%s/isbn/%s.json", s.BaseURL, cleanISBN)
	
	resp, err := s.HTTPClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("openlibrary api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result OpenLibraryBookResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode openlibrary response: %w", err)
		}
		return s.toBookData(&result), nil
	}

	// Fallback to search API
	results, err := s.searchMultipleByQuery(fmt.Sprintf("isbn:%s", cleanISBN), 1)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no books found")
	}
	return results[0], nil
}

// SearchByTitle searches for a book by title (returns single result)
func (s *OpenLibraryService) SearchByTitle(title string) (*BookData, error) {
	results, err := s.searchMultipleByQuery(fmt.Sprintf("title:%s", title), 1)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no books found")
	}
	return results[0], nil
}

// SearchByAuthor searches for books by author (returns single result)
func (s *OpenLibraryService) SearchByAuthor(author string) (*BookData, error) {
	results, err := s.searchMultipleByQuery(fmt.Sprintf("author:%s", author), 1)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no books found")
	}
	return results[0], nil
}

// SearchMultipleByISBN searches for books by ISBN (returns multiple results)
func (s *OpenLibraryService) SearchMultipleByISBN(isbn string, maxResults int) ([]*BookData, error) {
	cleanISBN := strings.ReplaceAll(isbn, "-", "")
	return s.searchMultipleByQuery(fmt.Sprintf("isbn:%s", cleanISBN), maxResults)
}

// SearchMultipleByTitle searches for books by title (returns multiple results)
func (s *OpenLibraryService) SearchMultipleByTitle(title string, maxResults int) ([]*BookData, error) {
	return s.searchMultipleByQuery(fmt.Sprintf("title:%s", title), maxResults)
}

// SearchMultipleByAuthor searches for books by author (returns multiple results)
func (s *OpenLibraryService) SearchMultipleByAuthor(author string, maxResults int) ([]*BookData, error) {
	return s.searchMultipleByQuery(fmt.Sprintf("author:%s", author), maxResults)
}

// searchMultipleByQuery performs a search query and returns multiple results
func (s *OpenLibraryService) searchMultipleByQuery(query string, maxResults int) ([]*BookData, error) {
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 100 {
		maxResults = 100 // Open Library API allows up to 100
	}

	params := url.Values{}
	params.Add("q", query)
	params.Add("limit", fmt.Sprintf("%d", maxResults))

	reqURL := fmt.Sprintf("%s/search.json?%s", s.BaseURL, params.Encode())

	resp, err := s.HTTPClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("openlibrary search failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openlibrary api returned status %d: %s", resp.StatusCode, string(body))
	}

	var result OpenLibrarySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode openlibrary search response: %w", err)
	}

	if result.NumFound == 0 || len(result.Docs) == 0 {
		return nil, fmt.Errorf("no books found")
	}

	// Convert all results to BookData
	books := make([]*BookData, 0, len(result.Docs))
	for i := range result.Docs {
		books = append(books, s.searchDocToBookData(&result.Docs[i]))
	}

	return books, nil
}

// toBookData converts Open Library response to normalized BookData
func (s *OpenLibraryService) toBookData(book *OpenLibraryBookResponse) *BookData {
	bookData := &BookData{
		Title:         book.Title,
		Publisher:     strings.Join(book.Publishers, ", "),
		PublishedDate: book.PublishDate,
		PageCount:     book.NumberOfPages,
		Categories:    book.Subjects,
		Source:        "open_library",
		RawData:       book,
	}

	// Extract ISBN
	if len(book.ISBN13) > 0 {
		bookData.ISBN13 = book.ISBN13[0]
		bookData.ISBN = book.ISBN13[0]
	} else if len(book.ISBN10) > 0 {
		bookData.ISBN = book.ISBN10[0]
	}

	// Extract language
	if len(book.Languages) > 0 {
		// Extract language code from key like "/languages/eng"
		parts := strings.Split(book.Languages[0].Key, "/")
		if len(parts) > 0 {
			bookData.Language = parts[len(parts)-1]
		}
	}

	// Extract description
	if book.Description != nil {
		switch desc := book.Description.(type) {
		case string:
			bookData.Description = desc
		case map[string]interface{}:
			if value, ok := desc["value"].(string); ok {
				bookData.Description = value
			}
		}
	}

	// Get cover images
	if len(book.Covers) > 0 {
		coverID := book.Covers[0]
		bookData.CoverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", coverID)
		bookData.ThumbnailURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", coverID)
	}

	return bookData
}

// searchDocToBookData converts search document to BookData
func (s *OpenLibraryService) searchDocToBookData(doc *OpenLibrarySearchDoc) *BookData {
	bookData := &BookData{
		Title:      doc.Title,
		Authors:    doc.AuthorName,
		Categories: doc.Subject,
		PageCount:  doc.NumberOfPagesMedian,
		Source:     "open_library",
		RawData:    doc,
	}

	// Extract ISBN
	if len(doc.ISBN) > 0 {
		for _, isbn := range doc.ISBN {
			if len(isbn) == 13 {
				bookData.ISBN13 = isbn
				bookData.ISBN = isbn
				break
			} else if len(isbn) == 10 && bookData.ISBN == "" {
				bookData.ISBN = isbn
			}
		}
	}

	// Extract publisher
	if len(doc.Publisher) > 0 {
		bookData.Publisher = doc.Publisher[0]
	}

	// Extract publish date
	if len(doc.PublishYear) > 0 {
		bookData.PublishedDate = fmt.Sprintf("%d", doc.PublishYear[0])
	}

	// Extract language
	if len(doc.Language) > 0 {
		bookData.Language = doc.Language[0]
	}

	// Get cover image
	if doc.CoverI > 0 {
		bookData.CoverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", doc.CoverI)
		bookData.ThumbnailURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", doc.CoverI)
	}

	return bookData
}

