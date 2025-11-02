# 🚀 Bookwise API - Quick Start Guide

5 dakikada Bookwise API'yi çalıştırın!

## Hızlı Başlangıç (Docker ile - Önerilen)

### 1. Gereksinimler

- Docker ve Docker Compose yüklü
- OpenAI API Key

### 2. Kurulum

```bash
# Projeyi klonlayın
git clone <repository-url>
cd bookwise_api

# .env dosyası oluşturun
echo "OPENAI_API_KEY=your_openai_key_here" > .env

# Container'ları başlatın
docker-compose up -d
```

### 3. Test Edin

```bash
# Health check
curl http://localhost:8080/health

# Kitap arayın
curl "http://localhost:8080/api/v1/books/search?q=9780262033848&type=isbn"
```

✅ Başarılı! API çalışıyor.

---

## Manuel Kurulum (Docker Olmadan)

### 1. Gereksinimler

- Go 1.22+
- PostgreSQL 16
- OpenAI API Key

### 2. PostgreSQL Kurulumu

```bash
# macOS
brew install postgresql@16
brew services start postgresql@16

# Ubuntu/Debian
sudo apt install postgresql-16
sudo systemctl start postgresql

# Veritabanı oluştur
createdb bookwise_db
```

### 3. Proje Kurulumu

```bash
# Bağımlılıkları yükle
go mod download

# .env dosyası oluştur
cp .env.example .env
```

`.env` dosyasını düzenleyin:

```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=bookwise_db
OPENAI_API_KEY=your_openai_key_here
```

### 4. Çalıştırın

```bash
# Direkt çalıştır
go run cmd/server/main.go

# veya build edip çalıştır
make build
./bin/bookwise-api
```

---

## İlk Testler

### 1. Health Check

```bash
curl http://localhost:8080/health
```

Beklenen sonuç:
```json
{
  "status": "healthy",
  "service": "bookwise-api",
  "time": "2025-10-28T10:30:00Z"
}
```

### 2. Kitap Arama (ISBN ile)

```bash
curl "http://localhost:8080/api/v1/books/search?q=9780262033848&type=isbn"
```

Bu:
1. Google Books ve OpenLibrary'den kitap bilgisini getirir
2. Veritabanına kaydeder
3. Asenkron olarak AI quiz oluşturur
4. Kitap bilgisini döndürür

### 3. Quiz'i Kontrol Et

Response'dan `book_id`'yi alın, sonra:

```bash
curl "http://localhost:8080/api/v1/quiz/{book_id}"
```

Quiz hala oluşturuluyorsa `"status": "generating"` döner.
Tamamlandıysa quiz sorularını göreceksiniz.

---

## Yaygın Sorunlar

### ❌ "Database connection failed"

```bash
# PostgreSQL çalışıyor mu kontrol et
pg_isready

# Veritabanı var mı kontrol et
psql -l | grep bookwise_db

# Yoksa oluştur
createdb bookwise_db
```

### ❌ "OPENAI_API_KEY is not set"

`.env` dosyasında `OPENAI_API_KEY` değişkenini ayarlayın.

### ❌ Port 8080 kulanımda

`.env` dosyasında `PORT` değişkenini değiştirin:

```env
PORT=3000
```

---

## Sonraki Adımlar

1. 📖 [README.md](README.md) - Detaylı dokümantasyon
2. 📡 [API_DOCUMENTATION.md](documents/API_DOCUMENTATION.md) - API referansı
3. 📋 [PRD.md](documents/PRD.md) - Ürün gereksinimler belgesi

---

## Örnek İş Akışı

```bash
# 1. Kitap ara
curl "http://localhost:8080/api/v1/books/search?q=9780262033848&type=isbn" \
  | jq '.data.id'

# Çıktı: "550e8400-e29b-41d4-a716-446655440000"

# 2. Birkaç saniye bekle (quiz oluşsun)
sleep 10

# 3. Quiz'i al
curl "http://localhost:8080/api/v1/quiz/550e8400-e29b-41d4-a716-446655440000" \
  | jq '.data.questions'

# Çıktı: Quiz soruları
```

---

## Geliştirme Modu

```bash
# Değişiklikleri izle ve otomatik yeniden başlat
# (air tool gerekli: go install github.com/cosmtrek/air@latest)
air

# veya
make run
```

---

## Production Deployment

```bash
# Production modunda çalıştır
GIN_MODE=release go run cmd/server/main.go

# veya Docker ile
docker-compose -f docker-compose.yml up -d
```

---

🎉 **Tebrikler!** Bookwise API artık çalışıyor.

Sorularınız için: [GitHub Issues](https://github.com/yourusername/bookwise_api/issues)

