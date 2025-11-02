package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// GoogleBooksService handles Google Books API integration
type GoogleBooksService struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// NewGoogleBooksService creates a new Google Books service
func NewGoogleBooksService(apiKey string) *GoogleBooksService {
	return &GoogleBooksService{
		APIKey:  apiKey,
		BaseURL: "https://www.googleapis.com/books/v1",
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GoogleBooksResponse represents the response from Google Books API
type GoogleBooksResponse struct {
	Kind       string `json:"kind"`
	TotalItems int    `json:"totalItems"`
	Items      []GoogleBookItem `json:"items"`
}

// GoogleBookItem represents a single book item from Google Books
type GoogleBookItem struct {
	ID         string `json:"id"`
	VolumeInfo GoogleVolumeInfo `json:"volumeInfo"`
}

// GoogleVolumeInfo contains the book information
type GoogleVolumeInfo struct {
	Title               string   `json:"title"`
	Authors             []string `json:"authors"`
	Publisher           string   `json:"publisher"`
	PublishedDate       string   `json:"publishedDate"`
	Description         string   `json:"description"`
	IndustryIdentifiers []GoogleIdentifier `json:"industryIdentifiers"`
	PageCount           int      `json:"pageCount"`
	Categories          []string `json:"categories"`
	Language            string   `json:"language"`
	ImageLinks          GoogleImageLinks `json:"imageLinks"`
}

// GoogleIdentifier represents ISBN identifiers
type GoogleIdentifier struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

// GoogleImageLinks contains image URLs
type GoogleImageLinks struct {
	SmallThumbnail string `json:"smallThumbnail"`
	Thumbnail      string `json:"thumbnail"`
	Small          string `json:"small"`
	Medium         string `json:"medium"`
	Large          string `json:"large"`
	ExtraLarge     string `json:"extraLarge"`
}

// BookData represents normalized book data from any source
type BookData struct {
	Title         string
	Authors       []string
	ISBN          string
	ISBN13        string
	Description   string
	Publisher     string
	PublishedDate string
	PageCount     int
	Categories    []string
	Language      string
	CoverURL      string
	ThumbnailURL  string
	Source        string
	RawData       interface{}
}

// SearchByISBN searches for a book by ISBN
func (s *GoogleBooksService) SearchByISBN(isbn string) (*BookData, error) {
	return s.search(fmt.Sprintf("isbn:%s", isbn))
}

// SearchByTitle searches for a book by title
func (s *GoogleBooksService) SearchByTitle(title string) (*BookData, error) {
	return s.search(fmt.Sprintf("intitle:%s", title))
}

// SearchByAuthor searches for books by author
func (s *GoogleBooksService) SearchByAuthor(author string) (*BookData, error) {
	return s.search(fmt.Sprintf("inauthor:%s", author))
}

// search performs the actual search query
func (s *GoogleBooksService) search(query string) (*BookData, error) {
	params := url.Values{}
	params.Add("q", query)
	params.Add("maxResults", "1")
	if s.APIKey != "" {
		params.Add("key", s.APIKey)
	}

	reqURL := fmt.Sprintf("%s/volumes?%s", s.BaseURL, params.Encode())

	resp, err := s.HTTPClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("google books api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google books api returned status %d: %s", resp.StatusCode, string(body))
	}

	var result GoogleBooksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode google books response: %w", err)
	}

	if result.TotalItems == 0 || len(result.Items) == 0 {
		return nil, fmt.Errorf("no books found")
	}

	// Convert first result to normalized BookData
	return s.toBookData(&result.Items[0]), nil
}

// toBookData converts Google Books response to normalized BookData
func (s *GoogleBooksService) toBookData(item *GoogleBookItem) *BookData {
	volume := item.VolumeInfo
	
	bookData := &BookData{
		Title:         volume.Title,
		Authors:       volume.Authors,
		Description:   volume.Description,
		Publisher:     volume.Publisher,
		PublishedDate: volume.PublishedDate,
		PageCount:     volume.PageCount,
		Categories:    volume.Categories,
		Language:      volume.Language,
		Source:        "google_books",
		RawData:       item,
	}

	// Extract ISBN and ISBN13
	for _, identifier := range volume.IndustryIdentifiers {
		switch identifier.Type {
		case "ISBN_10":
			if bookData.ISBN == "" {
				bookData.ISBN = identifier.Identifier
			}
		case "ISBN_13":
			bookData.ISBN13 = identifier.Identifier
			// Prefer ISBN-13 as primary ISBN
			if bookData.ISBN == "" {
				bookData.ISBN = identifier.Identifier
			}
		}
	}

	// Get best quality images
	if volume.ImageLinks.Large != "" {
		bookData.CoverURL = volume.ImageLinks.Large
	} else if volume.ImageLinks.Medium != "" {
		bookData.CoverURL = volume.ImageLinks.Medium
	} else if volume.ImageLinks.Small != "" {
		bookData.CoverURL = volume.ImageLinks.Small
	}

	if volume.ImageLinks.Thumbnail != "" {
		bookData.ThumbnailURL = volume.ImageLinks.Thumbnail
	} else if volume.ImageLinks.SmallThumbnail != "" {
		bookData.ThumbnailURL = volume.ImageLinks.SmallThumbnail
	}

	return bookData
}

