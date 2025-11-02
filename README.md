# 📘 Bookwise AI Backend

**Bookwise Backend**, AI destekli hibrit kitap bilgi servisi ve quiz oluşturma motoru.

## 🎯 Özellikler

- 🔍 **Hibrit Kitap Arama**: Google Books ve OpenLibrary kaynaklarını birleştirerek eksiksiz kitap bilgisi
- 🤖 **AI Quiz Üretimi**: Google Gemini Flash ile otomatik quiz oluşturma (ÜCRETSIZ!)
- 💾 **Akıllı Cache**: ISBN bazlı tekil kayıt, gereksiz API çağrılarını engelleme
- 🌐 **Global Quiz Paylaşımı**: Her kitap için tek quiz, tüm kullanıcılara aynı sorular
- ⚡ **Asenkron İşlemler**: Background worker ile quiz oluşturma

## 🏗️ Teknoloji Stack

- **Language**: Go 1.22+
- **Framework**: Gin
- **Database**: PostgreSQL 16
- **ORM**: GORM
- **AI**: Google Gemini API (gemini-1.5-flash) - ÜCRETSIZ!
- **External APIs**: Google Books API, OpenLibrary API

## 📋 Gereksinimler

- Go 1.22+
- PostgreSQL 16
- Google Gemini API Key (zorunlu - ücretsiz tier mevcut)
- Google Books API Key (opsiyonel)

## 🚀 Kurulum

### 1. Projeyi Klonlayın

```bash
git clone <repository-url>
cd bookwise_api
```

### 2. Bağımlılıkları Yükleyin

```bash
go mod download
```

### 3. Environment Değişkenlerini Ayarlayın

`.env.example` dosyasını `.env` olarak kopyalayın ve düzenleyin:

```bash
cp .env.example .env
```

Gerekli değişkenleri ayarlayın:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=bookwise_db

# Google Gemini (ZORUNLU - ÜCRETSIZ!)
GEMINI_API_KEY=your_gemini_api_key_here

# Google Books (Opsiyonel)
GOOGLE_BOOKS_API_KEY=your_google_books_api_key_here
```

### 4. PostgreSQL Veritabanı Oluşturun

```bash
createdb bookwise_db
```

### 5. Uygulamayı Çalıştırın

```bash
go run cmd/server/main.go
```

veya

```bash
make run
```

Uygulama `http://localhost:8080` adresinde çalışacaktır.

## 🐳 Docker ile Kurulum

### Docker Compose ile Çalıştırma

```bash
# .env dosyasını oluşturun ve GEMINI_API_KEY'i ekleyin
echo "GEMINI_API_KEY=your_key_here" > .env

# Container'ları başlatın
docker-compose up -d

# Logları görüntüleyin
docker-compose logs -f api
```

### Container'ları Durdurma

```bash
docker-compose down
```

## 📡 API Endpoints

### Health Check

```http
GET /health
GET /health/detailed
```

### Books

#### Kitap Arama (Hibrit Kaynak)
```http
GET /api/v1/books/search?q={query}&type={isbn|title|author}
```

**Örnek:**
```bash
curl "http://localhost:8080/api/v1/books/search?q=9780262033848&type=isbn"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Introduction to Algorithms",
    "authors": ["Thomas H. Cormen", "Charles E. Leiserson"],
    "isbn": "9780262033848",
    "description": "...",
    "quiz_status": "generating"
  },
  "cache_hit": false,
  "message": "Kitap başarıyla getirildi. Quiz oluşturuluyor..."
}
```

#### Kitap Detayı (UUID ile)
```http
GET /api/v1/books/:id
```

#### Kitap Detayı (ISBN ile)
```http
GET /api/v1/books/isbn/:isbn
```

#### Kitap Listesi
```http
GET /api/v1/books?page=1&limit=10
```

### Quiz

#### Quiz Getir (Book ID ile)
```http
GET /api/v1/quiz/:bookId
```

**Response:**
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
      }
    ],
    "ai_model": "gpt-4o-mini",
    "created_at": "2025-10-28T10:31:30Z"
  }
}
```

#### Quiz Getir (Quiz ID ile)
```http
GET /api/v1/quiz/id/:id
```

## 🔄 Sistem Akışı

```
1. Kullanıcı kitap arar (ISBN/Title/Author)
   ↓
2. Backend ISBN kontrolü yapar
   ├─ Varsa → Cache'den döner ✅
   └─ Yoksa → External API'lere gider
   ↓
3. Google Books API + OpenLibrary API
   ↓
4. Verileri birleştir ve normalize et
   ↓
5. Veritabanına kaydet
   ↓
6. Asenkron quiz oluşturma tetikle
   ↓
7. Kullanıcıya kitap bilgisini döndür
   ↓
8. (Background) AI quiz oluşur
```

## 🧪 Test

```bash
# Tüm testleri çalıştır
go test -v ./...

# Test coverage
go test -cover ./...
```

## 📊 Performans Hedefleri

| Metrik | Hedef |
|--------|-------|
| Kitap sorgulama süresi | < 2 saniye |
| Cache hit ratio | > %70 |
| AI quiz oluşturma | < 30 saniye |
| API uptime | > %99.5 |

## 🛠️ Geliştirme

### Proje Yapısı

```
bookwise_api/
├── cmd/
│   └── server/          # Main application
│       └── main.go
├── config/              # Configuration management
│   └── config.go
├── internal/
│   ├── database/        # Database connection & migrations
│   │   └── database.go
│   ├── handlers/        # HTTP handlers
│   │   ├── books.go
│   │   ├── quiz.go
│   │   └── health.go
│   ├── models/          # Data models
│   │   ├── book.go
│   │   └── quiz.go
│   └── services/        # Business logic
│       ├── googlebooks.go
│       ├── openlibrary.go
│       ├── bookmerger.go
│       ├── quizgenerator.go
│       └── quizworker.go
├── documents/           # Documentation
│   └── PRD.md
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Makefile Komutları

```bash
make help          # Tüm komutları listele
make build         # Uygulamayı derle
make run           # Uygulamayı çalıştır
make test          # Testleri çalıştır
make clean         # Build artifactlarını temizle
make deps          # Bağımlılıkları indir
make docker-build  # Docker image oluştur
make docker-up     # Docker container'ları başlat
make docker-down   # Docker container'ları durdur
make docker-logs   # Docker loglarını göster
```

## 🔐 Güvenlik

- Tüm hassas bilgiler (API keys) `.env` dosyasında saklanmalı
- `.env` dosyası asla commit edilmemeli (`.gitignore`'da)
- Production ortamında `GIN_MODE=release` kullanılmalı
- CORS ayarları production domain'lere göre yapılandırılmalı

## 📈 Monitoring

Detaylı health check endpoint'i sistem durumunu gösterir:

```bash
curl http://localhost:8080/health/detailed
```

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
  }
}
```

## 🐛 Troubleshooting

### Problem: Database connection failed

```bash
# PostgreSQL'in çalıştığından emin olun
systemctl status postgresql

# Veritabanının oluşturulduğundan emin olun
createdb bookwise_db
```

### Problem: OpenAI API hatası

- `OPENAI_API_KEY` environment variable'ının doğru ayarlandığından emin olun
- OpenAI hesabınızda yeterli kredi olduğunu kontrol edin
- Rate limit hatası alıyorsanız retry mekanizması devreye girecektir

### Problem: Quiz oluşturulmuyor

- Worker loglarını kontrol edin
- `/health/detailed` endpoint'inden worker durumunu kontrol edin
- Veritabanında `quiz_status='failed'` olan kitapları kontrol edin

## 📝 Lisans

MIT License

## 👥 Katkıda Bulunma

1. Fork yapın
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Commit edin (`git commit -m 'Add amazing feature'`)
4. Push edin (`git push origin feature/amazing-feature`)
5. Pull Request açın

## 📞 İletişim

Sorularınız için: [GitHub Issues](https://github.com/yourusername/bookwise_api/issues)

---

**Bookwise AI Backend** - Made with ❤️ and Go

