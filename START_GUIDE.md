# 🚀 Bookwise API - Başlangıç Kılavuzu

## ⚠️ Önkoşullar

1. **Docker Desktop** yüklü ve çalışıyor olmalı
2. **OpenAI API Key** (zorunlu)

## 📝 Adım Adım Kurulum

### 1. Docker'ı Başlat

```bash
# macOS: Docker Desktop uygulamasını başlatın
open -a Docker

# Çalıştığını kontrol edin
docker --version
docker ps
```

### 2. .env Dosyası Oluştur

```bash
cd /Users/enes/Documents/bookwise_api

# .env dosyası oluştur
cat > .env << 'EOF'
# Server Configuration
PORT=8080
GIN_MODE=debug

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=bookwise_db
DB_SSLMODE=disable

# Google Gemini AI Configuration (ZORUNLU!)
GEMINI_API_KEY=your_gemini_key_here
GEMINI_MODEL=gemini-1.5-flash

# External APIs (Opsiyonel)
GOOGLE_BOOKS_API_KEY=

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080

# Quiz Configuration
QUIZ_QUESTIONS_COUNT=5
QUIZ_RETRY_LIMIT=3
EOF
```

**ÖNEMLİ:** `.env` dosyasındaki `GEMINI_API_KEY=your_gemini_key_here` satırını kendi Google Gemini API key'inizle değiştirin!

Google Gemini API Key almak için: https://aistudio.google.com/app/apikey (ÜCRETSİZ!)

### 3. Docker Build

```bash
cd /Users/enes/Documents/bookwise_api

# Docker image'leri build et
docker-compose build
```

Beklenen çıktı:
```
✅ Building api
✅ Successfully built...
```

### 4. Container'ları Başlat

```bash
# Detached modda başlat
docker-compose up -d

# Logları takip et
docker-compose logs -f api
```

Beklenen çıktı:
```
✅ Database connection established
✅ Database migrations completed
✅ Database indexes created
🚀 Starting quiz worker pool with 3 workers
👷 Worker #1 started
👷 Worker #2 started
👷 Worker #3 started
🚀 Server starting on :8080
```

### 5. Test Et

#### 5.1 Health Check

```bash
curl http://localhost:8080/health
```

Beklenen sonuç:
```json
{
  "status": "healthy",
  "service": "bookwise-api",
  "time": "2025-11-02T11:35:00Z"
}
```

#### 5.2 Detaylı Health Check

```bash
curl http://localhost:8080/health/detailed
```

#### 5.3 Kitap Arama (ISBN)

```bash
curl "http://localhost:8080/api/v1/books/search?q=9780262033848&type=isbn"
```

Beklenen sonuç:
```json
{
  "success": true,
  "data": {
    "id": "...",
    "title": "Introduction to Algorithms",
    "authors": ["Thomas H. Cormen", "Charles E. Leiserson"],
    "isbn": "9780262033848",
    "quiz_status": "generating"
  },
  "cache_hit": false,
  "message": "Kitap başarıyla getirildi. Quiz oluşturuluyor..."
}
```

#### 5.4 Quiz Sorgulama

Response'dan `book_id`'yi kopyalayın:

```bash
# book_id'yi değiştirin
curl "http://localhost:8080/api/v1/quiz/YOUR_BOOK_ID_HERE"
```

İlk sorguda:
```json
{
  "success": false,
  "status": "generating",
  "message": "Quiz şu anda oluşturuluyor..."
}
```

10-30 saniye sonra:
```json
{
  "success": true,
  "data": {
    "questions": [...]
  }
}
```

## 🎯 Test Senaryosu - Eksiksiz Workflow

```bash
# 1. Health check
curl http://localhost:8080/health

# 2. Kitap ara
RESPONSE=$(curl -s "http://localhost:8080/api/v1/books/search?q=9780262033848&type=isbn")
echo $RESPONSE | jq .

# 3. Book ID'yi al (jq yüklüyse)
BOOK_ID=$(echo $RESPONSE | jq -r '.data.id')
echo "Book ID: $BOOK_ID"

# 4. 15 saniye bekle (quiz oluşsun)
echo "Quiz oluşuyor, 15 saniye bekleniyor..."
sleep 15

# 5. Quiz'i getir
curl "http://localhost:8080/api/v1/quiz/$BOOK_ID" | jq .
```

## 🔍 Container Yönetimi

### Container Durumunu Kontrol Et

```bash
docker-compose ps
```

### Logları İzle

```bash
# Tüm servisler
docker-compose logs -f

# Sadece API
docker-compose logs -f api

# Sadece PostgreSQL
docker-compose logs -f postgres
```

### Container'ları Durdur

```bash
docker-compose down
```

### Container'ları Yeniden Başlat

```bash
docker-compose restart api
```

### Veritabanını Sıfırla

```bash
# Container'ları durdur ve volume'leri sil
docker-compose down -v

# Yeniden başlat
docker-compose up -d
```

## 🐛 Sorun Giderme

### Problem 1: Docker daemon çalışmıyor

```bash
# Hata: Cannot connect to the Docker daemon
# Çözüm: Docker Desktop'ı başlatın
open -a Docker
```

### Problem 2: GEMINI_API_KEY hatası

```bash
# Hata: "GEMINI_API_KEY" variable is not set
# Çözüm: .env dosyasını kontrol edin
cat .env | grep GEMINI_API_KEY

# Key ekleyin
nano .env
# veya
code .env
```

### Problem 3: Port zaten kullanımda

```bash
# Hata: port is already allocated
# Çözüm: Port'u kullanan servisi durdurun
lsof -i :8080
kill -9 <PID>

# veya .env'de PORT'u değiştirin
PORT=3000
```

### Problem 4: Database connection failed

```bash
# PostgreSQL container'ının çalıştığını kontrol edin
docker-compose ps postgres

# PostgreSQL loglarını kontrol edin
docker-compose logs postgres

# Container'ları yeniden başlatın
docker-compose restart postgres
```

### Problem 5: Quiz oluşturulmuyor

```bash
# Worker loglarını kontrol edin
docker-compose logs -f api | grep "Worker"

# Worker stats'ı kontrol edin
curl http://localhost:8080/health/detailed | jq '.components.quiz_worker'

# Gemini API key'i kontrol edin
curl http://localhost:8080/health/detailed
```

## 📊 Database'e Bağlanma

```bash
# PostgreSQL container'ına bağlan
docker-compose exec postgres psql -U postgres -d bookwise_db

# SQL komutları
\dt                           # Tabloları listele
SELECT * FROM books;          # Kitapları listele
SELECT * FROM quizzes;        # Quiz'leri listele
SELECT COUNT(*) FROM books;   # Kitap sayısı
\q                            # Çıkış
```

## 🧹 Temizlik

```bash
# Container'ları ve volume'leri sil
docker-compose down -v

# Docker image'lerini sil
docker-compose down --rmi all

# Tüm Docker kaynaklarını temizle
docker system prune -a --volumes
```

## 📝 Notlar

1. **İlk çalıştırma** biraz zaman alabilir (dependencies download, build)
2. **Quiz oluşturma** 10-30 saniye sürebilir (OpenAI API)
3. **Google Books API Key** opsiyoneldir, yoksa sadece OpenLibrary kullanılır
4. **Cache sistemi** çalışıyor - aynı ISBN'i tekrar aradığınızda cache'den döner

## 🎉 Başarılı Kurulum Kontrolü

Eğer şunları görüyorsanız her şey hazır:

- ✅ `curl http://localhost:8080/health` → `"status": "healthy"`
- ✅ `docker-compose ps` → api ve postgres `Up`
- ✅ Book search çalışıyor
- ✅ Quiz oluşuyor

## 🚀 Sonraki Adımlar

1. **Postman Collection** import edin → `Bookwise_API.postman_collection.json`
2. **API Documentation** okuyun → `documents/API_DOCUMENTATION.md`
3. **Postman Guide** inceleyin → `documents/POSTMAN_GUIDE.md`

---

**İyi çalışmalar!** 🎯

Sorular için: [GitHub Issues](https://github.com/yourusername/bookwise_api/issues)

