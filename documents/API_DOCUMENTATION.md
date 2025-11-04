# 📡 Bookwise API Documentation

Complete API reference for Bookwise AI Backend.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Currently, the API does not require authentication. This will be added in future versions with JWT/Firebase Auth.

---

## Endpoints

### 1. Health Check

#### GET /health

Basic health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "service": "bookwise-api",
  "time": "2025-10-28T10:30:00Z"
}
```

#### GET /health/detailed

Detailed health check with system statistics.

**Response:**
```json
{
  "status": "healthy",
  "service": "bookwise-api",
  "uptime": "2h30m15s",
  "components": {
    "database": "healthy",
    "quiz_worker": {
      "total_books": 150,
      "pending": 5,
      "generating": 2,
      "completed": 140,
      "failed": 3,
      "queue_size": 5,
      "worker_count": 3,
      "worker_running": true
    }
  },
  "timestamp": "2025-10-28T10:30:00Z"
}
```

---

### 2. Books

#### GET /api/v1/books/search

Search for books using hybrid sources (Google Books + OpenLibrary). Returns a list of search results.

> ⚠️ **Note**: This endpoint only searches for books and returns results. It does NOT save books to the database or generate quizzes. Use `POST /api/v1/books` to save a book.

**Query Parameters:**
- `q` (required): Search query
- `type` (optional): Search type - `isbn`, `title`, or `author` (default: `title`)
- `limit` (optional): Maximum number of results (default: 10, max: 40)

**Examples:**

Search by Title:
```bash
curl "http://localhost:8080/api/v1/books/search?q=Introduction+to+Algorithms&type=title&limit=10"
```

Search by ISBN:
```bash
curl "http://localhost:8080/api/v1/books/search?q=9780262033848&type=isbn"
```

Search by Author:
```bash
curl "http://localhost:8080/api/v1/books/search?q=Thomas+Cormen&type=author&limit=5"
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "title": "Introduction to Algorithms",
      "authors": ["Thomas H. Cormen", "Charles E. Leiserson"],
      "isbn": "0262033844",
      "isbn13": "9780262033848",
      "description": "A comprehensive textbook covering...",
      "publisher": "MIT Press",
      "published_date": "2009-07-31",
      "page_count": 1312,
      "categories": ["Computers", "Algorithms"],
      "language": "en",
      "cover_url": "https://books.google.com/books/content?id=...",
      "thumbnail_url": "https://books.google.com/books/content?id=...",
      "source": "google_books"
    },
    {
      "title": "Introduction to Algorithms, 4th Edition",
      "authors": ["Thomas H. Cormen"],
      "isbn": "026204630X",
      "isbn13": "9780262046305",
      "description": "...",
      "source": "open_library"
    }
  ],
  "count": 2,
  "message": "2 kitap bulundu"
}
```

**Response (404 Not Found):**
```json
{
  "success": false,
  "error": "Kitap bulunamadı",
  "details": "no books found in any source"
}
```

---

#### POST /api/v1/books

Save a book to the database. Fetches book details from external sources using ISBN and saves it.

**Request Body:**
```json
{
  "isbn": "9780262033848",
  "generate_quiz": true
}
```

**Parameters:**
- `isbn` (required): Book ISBN (ISBN-10 or ISBN-13)
- `generate_quiz` (optional): Whether to automatically generate quiz (default: false)

**Example:**
```bash
curl -X POST "http://localhost:8080/api/v1/books" \
  -H "Content-Type: application/json" \
  -d '{
    "isbn": "9780262033848",
    "generate_quiz": true
  }'
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Introduction to Algorithms",
    "authors": ["Thomas H. Cormen", "Charles E. Leiserson"],
    "isbn": "9780262033848",
    "isbn13": "9780262033848",
    "description": "A comprehensive textbook covering...",
    "publisher": "MIT Press",
    "published_date": "2009-07-31",
    "page_count": 1312,
    "categories": ["Computers", "Algorithms"],
    "language": "en",
    "cover_url": "https://books.google.com/books/content?id=...",
    "thumbnail_url": "https://books.google.com/books/content?id=...",
    "data_sources": ["google_books", "open_library"],
    "quiz_status": "generating",
    "quiz_id": null,
    "created_at": "2025-10-28T10:30:00Z"
  },
  "message": "Kitap başarıyla kaydedildi. Quiz oluşturuluyor..."
}
```

**Response (200 OK) - Book Already Exists:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Introduction to Algorithms",
    ...
  },
  "message": "Kitap zaten kayıtlı"
}
```

**Response (404 Not Found):**
```json
{
  "success": false,
  "error": "Kitap bulunamadı",
  "details": "book not found in any source"
}
```

---

#### GET /api/v1/books/:id

Get book details by UUID.

**Path Parameters:**
- `id` (required): Book UUID

**Example:**
```bash
curl "http://localhost:8080/api/v1/books/550e8400-e29b-41d4-a716-446655440000"
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Introduction to Algorithms",
    ...
  }
}
```

**Response (404 Not Found):**
```json
{
  "success": false,
  "error": "Kitap bulunamadı"
}
```

---

#### GET /api/v1/books/isbn/:isbn

Get book details by ISBN.

**Path Parameters:**
- `isbn` (required): Book ISBN (ISBN-10 or ISBN-13)

**Example:**
```bash
curl "http://localhost:8080/api/v1/books/isbn/9780262033848"
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Introduction to Algorithms",
    ...
  }
}
```

**Response (404 Not Found):**
```json
{
  "success": false,
  "error": "Kitap bulunamadı",
  "message": "Bu ISBN ile kayıtlı kitap bulunamadı. /books/search?q=9780262033848&type=isbn ile arama yapabilirsiniz."
}
```

---

#### GET /api/v1/books

List all books with pagination.

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 10, max: 100)

**Example:**
```bash
curl "http://localhost:8080/api/v1/books?page=1&limit=20"
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Introduction to Algorithms",
      ...
    },
    ...
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

---

#### POST /api/v1/books/:id/generate-quiz

Manually trigger quiz generation for a specific book.

**Path Parameters:**
- `id` (required): Book UUID

**Example:**
```bash
curl -X POST "http://localhost:8080/api/v1/books/550e8400-e29b-41d4-a716-446655440000/generate-quiz"
```

**Response (202 Accepted) - Quiz Generation Started:**
```json
{
  "success": true,
  "message": "Quiz oluşturma işlemi başlatıldı. Lütfen birkaç saniye sonra kontrol edin.",
  "status": "generating"
}
```

**Response (200 OK) - Quiz Already Exists:**
```json
{
  "success": true,
  "message": "Quiz zaten oluşturulmuş. Yeni quiz oluşturulsun mu?",
  "status": "completed"
}
```

**Response (202 Accepted) - Already Generating:**
```json
{
  "success": false,
  "message": "Quiz şu anda oluşturuluyor. Lütfen bekleyin.",
  "status": "generating"
}
```

**Response (404 Not Found):**
```json
{
  "success": false,
  "error": "Kitap bulunamadı"
}
```

---

### 3. Quiz

#### GET /api/v1/quiz/:bookId

Get quiz for a book by book ID.

**Path Parameters:**
- `bookId` (required): Book UUID

**Example:**
```bash
curl "http://localhost:8080/api/v1/quiz/550e8400-e29b-41d4-a716-446655440000"
```

**Response (200 OK) - Quiz Completed:**
```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440111",
    "book_id": "550e8400-e29b-41d4-a716-446655440000",
    "questions": [
      {
        "question": "Big O notasyonu ne için kullanılır?",
        "options": [
          "A) Algoritmanın doğruluğunu ölçmek",
          "B) Algoritmanın zaman karmaşıklığını ifade etmek",
          "C) Algoritmanın bellek kullanımını hesaplamak",
          "D) Algoritmanın okunabilirliğini değerlendirmek"
        ],
        "answer": "B) Algoritmanın zaman karmaşıklığını ifade etmek",
        "explanation": "Big O notasyonu, algoritmaların asimptotik zaman karmaşıklığını tanımlar."
      },
      ...
    ],
    "ai_model": "gpt-4o-mini",
    "created_at": "2025-10-28T10:31:30Z"
  }
}
```

**Response (202 Accepted) - Quiz Pending:**
```json
{
  "success": false,
  "status": "pending",
  "message": "Quiz henüz oluşturulmadı. Lütfen daha sonra tekrar deneyin."
}
```

**Response (202 Accepted) - Quiz Generating:**
```json
{
  "success": false,
  "status": "generating",
  "message": "Quiz şu anda oluşturuluyor. Lütfen birkaç saniye sonra tekrar deneyin."
}
```

**Response (500 Internal Server Error) - Quiz Failed:**
```json
{
  "success": false,
  "status": "failed",
  "error": "Quiz oluşturulamadı. Lütfen destek ekibiyle iletişime geçin."
}
```

**Response (404 Not Found):**
```json
{
  "success": false,
  "error": "Kitap bulunamadı"
}
```

---

#### GET /api/v1/quiz/id/:id

Get quiz by quiz ID.

**Path Parameters:**
- `id` (required): Quiz UUID

**Example:**
```bash
curl "http://localhost:8080/api/v1/quiz/id/660e8400-e29b-41d4-a716-446655440111"
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440111",
    "book_id": "550e8400-e29b-41d4-a716-446655440000",
    "questions": [...],
    "ai_model": "gpt-4o-mini",
    "created_at": "2025-10-28T10:31:30Z"
  }
}
```

**Response (404 Not Found):**
```json
{
  "success": false,
  "error": "Quiz bulunamadı"
}
```

---

## Status Codes

| Code | Description |
|------|-------------|
| 200  | OK - Request successful |
| 201  | Created - Resource created successfully |
| 202  | Accepted - Request accepted but processing not complete (quiz generating) |
| 400  | Bad Request - Invalid parameters |
| 404  | Not Found - Resource not found |
| 500  | Internal Server Error - Server error |

---

## Quiz Status

Books have a `quiz_status` field that indicates the quiz generation state:

| Status | Description |
|--------|-------------|
| `pending` | Quiz not yet generated (in queue) |
| `generating` | Quiz is currently being generated by AI |
| `completed` | Quiz successfully generated |
| `failed` | Quiz generation failed (will be retried) |

---

## Rate Limiting

Currently, there are no rate limits. This will be implemented in future versions.

---

## CORS

CORS is enabled for origins specified in the `ALLOWED_ORIGINS` environment variable.

Default: `http://localhost:3000`

---

## Error Handling

All errors follow this format:

```json
{
  "success": false,
  "error": "Error message",
  "details": "Additional error details (optional)"
}
```

---

## Best Practices

1. **Two-Step Process**: 
   - First, search for books using `GET /books/search` to get a list of results
   - Then, save the desired book using `POST /books` with the ISBN
2. **Quiz Generation**: 
   - Quiz is NOT generated automatically unless you set `generate_quiz: true` in POST /books
   - You can manually trigger quiz generation using `POST /books/:id/generate-quiz`
3. **ISBN Search**: Always use ISBN search when you have the ISBN for best accuracy
4. **Cache**: Check if a book exists using `/books/isbn/:isbn` before calling `/books/search`
5. **Quiz Polling**: If quiz status is `generating`, poll `/quiz/:bookId` every 5-10 seconds
6. **Error Handling**: Always check the `success` field in responses

---

## Examples with cURL

### Complete Workflow (Recommended)

```bash
# 1. Search for books by title (returns list)
curl "http://localhost:8080/api/v1/books/search?q=Introduction+to+Algorithms&type=title&limit=10"

# 2. Select a book from results and save it (using ISBN)
curl -X POST "http://localhost:8080/api/v1/books" \
  -H "Content-Type: application/json" \
  -d '{
    "isbn": "9780262033848",
    "generate_quiz": true
  }'

# 3. Get book details (extract ID from previous response)
curl "http://localhost:8080/api/v1/books/550e8400-e29b-41d4-a716-446655440000"

# 4. Wait for quiz to be generated (check quiz_status)
# 5. Get quiz
curl "http://localhost:8080/api/v1/quiz/550e8400-e29b-41d4-a716-446655440000"
```

### Manual Quiz Generation Workflow

```bash
# 1. Save book without quiz
curl -X POST "http://localhost:8080/api/v1/books" \
  -H "Content-Type: application/json" \
  -d '{
    "isbn": "9780262033848",
    "generate_quiz": false
  }'

# 2. Later, manually trigger quiz generation
curl -X POST "http://localhost:8080/api/v1/books/550e8400-e29b-41d4-a716-446655440000/generate-quiz"

# 3. Poll for quiz
curl "http://localhost:8080/api/v1/quiz/550e8400-e29b-41d4-a716-446655440000"
```

### List All Books

```bash
curl "http://localhost:8080/api/v1/books?page=1&limit=10"
```

### Health Check

```bash
curl "http://localhost:8080/health/detailed"
```

---

## Postman Collection

A Postman collection is available for easier API testing. Import the collection and set the `base_url` variable to `http://localhost:8080`.

---

## SDK / Client Libraries

Client libraries for popular languages coming soon:
- [ ] JavaScript/TypeScript
- [ ] Dart/Flutter
- [ ] Python
- [ ] Java

---

**Last Updated**: November 4, 2025

---

## 🔄 Recent Changes

### November 4, 2025
- **Breaking Change**: `GET /books/search` now returns a list of books instead of a single book
- **Breaking Change**: `GET /books/search` no longer saves books to database or generates quizzes automatically
- **New Endpoint**: `POST /books` - Save a book to the database with optional quiz generation
- **New Endpoint**: `POST /books/:id/generate-quiz` - Manually trigger quiz generation for a specific book
- **Improved**: Search results now return multiple books from both Google Books and Open Library
- **Improved**: Better control over quiz generation (manual vs automatic)

