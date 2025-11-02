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

Search for books using hybrid sources (Google Books + OpenLibrary).

**Query Parameters:**
- `q` (required): Search query
- `type` (required): Search type - `isbn`, `title`, or `author`

**Examples:**

Search by ISBN:
```bash
curl "http://localhost:8080/api/v1/books/search?q=9780262033848&type=isbn"
```

Search by Title:
```bash
curl "http://localhost:8080/api/v1/books/search?q=Introduction+to+Algorithms&type=title"
```

Search by Author:
```bash
curl "http://localhost:8080/api/v1/books/search?q=Thomas+Cormen&type=author"
```

**Response (200 OK):**
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
  "cache_hit": false,
  "message": "Kitap başarıyla getirildi. Quiz oluşturuluyor..."
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

1. **ISBN Search**: Always use ISBN search when you have the ISBN for best accuracy
2. **Cache**: Check if a book exists using `/books/isbn/:isbn` before calling `/books/search`
3. **Quiz Polling**: If quiz status is `generating`, poll `/quiz/:bookId` every 5-10 seconds
4. **Error Handling**: Always check the `success` field in responses

---

## Examples with cURL

### Complete Workflow

```bash
# 1. Search for a book
curl "http://localhost:8080/api/v1/books/search?q=9780262033848&type=isbn"

# 2. Get book details (extract ID from previous response)
curl "http://localhost:8080/api/v1/books/550e8400-e29b-41d4-a716-446655440000"

# 3. Wait for quiz to be generated (check quiz_status)
# 4. Get quiz
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

**Last Updated**: October 28, 2025

