# 📘 **Bookwise AI Backend – Product Requirements Document (PRD)**

---

## 🧭 1. Özet

**Bookwise Backend**, kullanıcıların kitap bilgilerine eksiksiz ve hızlı erişebildiği, yapay zekâ destekli quiz'ler oluşturabilen bir **kitap bilgi ve AI quiz servisi**dir.

### Temel Amaçlar:

1. 📚 **Hibrit Kitap Bilgi Servisi**: Farklı kaynaklardan (Google Books, Open Library, vb.) kitap verilerini birleştirerek tek, normalize, eksiksiz kitap bilgisi sunmak.

2. 🤖 **AI Quiz Üretim Motoru**: Google Gemini Flash kullanarak kitap içeriğine göre otomatik olarak akıllı quiz soruları oluşturmak ve global olarak paylaşmak (her kitap için tek quiz seti).

**Not:** Kullanıcılar aktif olarak kitap "eklemez". Backend, talep edilen kitabı farklı kaynaklardan otomatik olarak çeker, normalize eder ve sunumlar.

### Core Features Özet

| Feature                   | Açıklama                                          | Teknoloji        |
| ------------------------- | ------------------------------------------------- | ---------------- |
| 🔍 Hibrit Kitap Arama     | Çoklu kaynaktan veri toplama ve birleştirme      | Google Books API, OpenLibrary API |
| 💾 Akıllı Cache           | ISBN bazlı tekil kayıt, tekrar eden API çağrılarını engelleme | PostgreSQL + Unique Constraint |
| 🤖 AI Quiz Oluşturma      | Kitap bilgisinden otomatik quiz üretimi          | Google Gemini Flash |
| 🌐 Global Quiz Paylaşımı  | Aynı kitap için tek quiz, tüm kullanıcılara aynı sorular | PostgreSQL JSONB |
| ⚡ Asenkron İşlemler      | Quiz oluşturma background'da, kullanıcı beklemiyor | Goroutines / Worker Pool |

---

## 🎯 2. Amaç ve Hedefler

### 2.1 Ürün Amacı

* Kullanıcılar ISBN, kitap adı veya yazar ile **kitap bilgisi talep edebilmeli**.
* Sistem, talep edilen kitabı **farklı kaynaklardan (Google Books, Open Library)** otomatik olarak arayıp **birleştirilmiş, normalize edilmiş** bilgi sunmalı.
* Aynı kitap (ISBN bazlı) **sadece bir kez veritabanına kaydedilmeli** (cache mantığı).
* AI (Google Gemini Flash) kitap içeriğine göre **otomatik quiz oluşturmalı**.
* Quiz **bir kez oluşturulmalı** ve **tüm kullanıcılar tarafından paylaşılabilmeli** (global quiz sistemi).
* Veriler PostgreSQL üzerinde güvenli ve performanslı şekilde tutulmalı.

### 2.2 Başarı Kriterleri

* 📚 **Hibrit veri kalitesi**: Farklı kaynaklardan gelen verilerin %95+ başarıyla birleştirilmesi.
* ⚙️ **Tekil kitap garantisi**: Aynı ISBN'in veritabanında sadece bir kez bulunması (%100 unique constraint).
* ⚡ **Cache performansı**: Daha önce sorgulanan kitapların cache'ten dönmesi (hit ratio %70+).
* 🧠 **Quiz kalitesi**: AI'nın JSON formatında geçerli quiz üretmesi (%95+ başarı oranı).
* 💰 **Maliyet optimizasyonu**: Gemini API çağrılarının minimize edilmesi (sadece yeni kitaplar için AI çağrısı) - Gemini Flash ücretsiz tier 15 request/minute.

---

## ⚙️ 3. Sistem Mimarisi

### 3.1 Genel Yapı

```
[Flutter App] 
    ↓
 [Bookwise API (Go)]
    ↓
 ┌─────────────────────────────┐
 │ PostgreSQL (books, quizzes) │
 ├─────────────────────────────┤
 │ External APIs:              │
 │ - Google Books              │
 │ - OpenLibrary               │
 ├─────────────────────────────┤
 │ AI Service: OpenAI API      │
 ├─────────────────────────────┤
 │ Redis (optional cache)      │
 └─────────────────────────────┘
```

### 3.2 Teknoloji Stack

| Katman           | Teknoloji                                 |
| ---------------- | ----------------------------------------- |
| Backend Language | Go (Golang 1.22+)                         |
| Framework        | Gin / Fiber                               |
| Database         | PostgreSQL 16                             |
| ORM              | GORM                                      |
| Cache            | Redis (isteğe bağlı)                      |
| AI Integration   | Google Gemini API (gemini-1.5-flash)      |
| Auth             | JWT (opsiyonel Firebase Auth integration) |
| External APIs    | Google Books API, OpenLibrary API         |

### 3.3 Sistem Akışı (User Journey)

```
1. Kullanıcı Flutter app'te kitap arar (ISBN/Title/Author)
   ↓
2. Flutter → GET /books/search?q=isbn:9780262033848&type=isbn
   ↓
3. Backend ISBN kontrolü yapar
   ├─ Varsa → DB'den döner (Cache Hit) ✅
   └─ Yoksa → Adım 4'e geç
   ↓
4. Google Books API çağrısı
   ├─ Başarılı → Veriyi al
   └─ Başarısız → OpenLibrary'ye fallback
   ↓
5. OpenLibrary API çağrısı
   ├─ Başarılı → Veriyi al
   └─ Başarısız → Hata döndür
   ↓
6. İki kaynağı birleştir (merge)
   ↓
7. Normalize et ve DB'ye kaydet
   ↓
8. Asenkron AI quiz oluşturma tetikle (background job)
   ↓
9. Kullanıcıya kitap bilgisini döndür
   ↓
10. (Background) AI quiz oluşur ve DB'ye kaydedilir
    ↓
11. Kullanıcı /quiz/:bookId ile quiz'i alabilir
```

---

## 🧩 4. Özellikler ve Gereksinimler

### 4.1 Kitap Bilgisi Sorgulama ve Hibrit Veri Toplama

#### Tanım

Kullanıcı **ISBN, kitap adı veya yazar** ile kitap bilgisi talep eder.
Backend şu akışı izler:

1. **Cache Kontrolü**: Veritabanında bu ISBN'e sahip kitap var mı?
   - ✅ Varsa → Mevcut veriyi döndür (cache hit)
   - ❌ Yoksa → Adım 2'ye geç

2. **Hibrit Veri Toplama**: 
   - Google Books API'den kitap bilgilerini çek
   - Open Library API'den kitap bilgilerini çek
   - İki kaynağı birleştir, eksik alanları tamamla

3. **Normalizasyon**: Birleştirilmiş veriyi standart formata çevir

4. **Kayıt ve Döndürme**: 
   - Veritabanına kaydet
   - AI quiz oluşturma işlemini tetikle (asenkron)
   - Kullanıcıya normalize edilmiş kitap bilgisini döndür

#### Fonksiyonel Gereksinimler

* `GET /books/search?q={query}&type={isbn|title|author}` → Kitap bilgisi getir (hibrit kaynak)
* `GET /books/:id` → Kitap detayını getir (cache'den)
* `GET /books/isbn/:isbn` → ISBN ile direkt kitap getir
* Normalize edilmiş veri modeli döndürülmeli.
* İlk sorgulamada otomatik AI quiz tetiklenmeli.

#### Veri Modeli

```go
type Book struct {
  ID            uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
  Title         string         `gorm:"not null"`
  Authors       pq.StringArray `gorm:"type:text[]"`
  ISBN          string         `gorm:"uniqueIndex;not null"`
  ISBN13        string         
  Description   string         `gorm:"type:text"`
  Publisher     string
  PublishedDate string
  PageCount     int
  Categories    pq.StringArray `gorm:"type:text[]"`
  Language      string
  CoverURL      string
  ThumbnailURL  string
  SourceData    datatypes.JSON `gorm:"type:jsonb"` // Ham veri (debugging için)
  DataSources   pq.StringArray `gorm:"type:text[]"` // ["google_books", "open_library"]
  QuizID        *uuid.UUID     `gorm:"type:uuid"`
  QuizStatus    string         `gorm:"default:'pending'"` // "pending", "generating", "completed", "failed"
  CreatedAt     time.Time
  UpdatedAt     time.Time
}
```

**Hibrit Birleştirme Kuralları:**
- **Öncelik:** Google Books > OpenLibrary
- **Eksik alanlar:** Diğer kaynaktan tamamlanır
- **Çakışma:** Google Books verisi tercih edilir

#### Kaynak API Öncelik Sırası

1. Google Books
2. OpenLibrary

#### Neden Hibrit Sistem?

| Sorun                          | Çözüm                                                  |
| ------------------------------ | ------------------------------------------------------ |
| Google Books bazı kitaplarda eksik bilgi | OpenLibrary'den tamamlanır                      |
| OpenLibrary bazı kitaplarda eski kapak | Google Books'tan güncel kapak alınır               |
| Tek kaynak down olursa        | Diğer kaynak fallback görevi görür                     |
| Veri kalitesi tutarsızlığı    | İki kaynağın birleşimi daha eksiksiz sonuç verir       |

**Örnek Hibrit Birleştirme:**
```
Google Books → title, authors, description, cover_url
OpenLibrary → page_count (eksikse), publisher (eksikse)
Sonuç → Tüm alanlar dolu bir kitap objesi
```

---

### 4.2 AI Quiz Oluşturma

#### Tanım

Yeni eklenen kitap için sistem Google Gemini API'ye kitap bilgilerini gönderir.
Model (gemini-1.5-flash), kitap hakkında JSON formatında 5 adet quiz sorusu döner.
Quiz sadece bir kez oluşturulur ve global paylaşılır.

#### Fonksiyonel Gereksinimler

* `POST /quiz/generate` → (internal) kitap bilgisine göre quiz üretir.
* `GET /quiz/:bookId` → kitabın quiz'ini getirir.

#### Prompt Formatı

```text
Kitabın bilgileri:
{book_info_json}

Bu kitap hakkında 5 adet çoktan seçmeli quiz sorusu oluştur.
JSON formatında dön:
{
  "quiz": [
    {
      "question": "...",
      "options": ["A", "B", "C", "D"],
      "answer": "...",
      "explanation": "..."
    }
  ]
}
```

#### Örnek Quiz Modeli

```go
type Quiz struct {
  ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
  BookID      uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex"` // Her kitap için bir quiz
  Questions   datatypes.JSON `gorm:"type:jsonb;not null"`
  AIModel     string         `gorm:"default:'gemini-1.5-flash'"`
  Status      string         `gorm:"default:'completed'"` // "completed", "failed", "retrying"
  RetryCount  int            `gorm:"default:0"`
  ErrorLog    string         `gorm:"type:text"`
  CreatedAt   time.Time
  UpdatedAt   time.Time
}
```

#### Örnek Quiz JSON Formatı

```json
{
  "quiz": [
    {
      "question": "Kitabın ana teması nedir?",
      "options": ["A) Aşk", "B) Savaş", "C) Bilim", "D) Doğa"],
      "answer": "C) Bilim",
      "explanation": "Kitap bilimsel gelişmelerin toplum üzerindeki etkisini ele alır."
    },
    {
      "question": "Yazarın kullandığı anlatım tekniği hangisidir?",
      "options": ["A) Birinci şahıs", "B) Üçüncü şahıs", "C) Çoklu bakış açısı", "D) Mektup"],
      "answer": "B) Üçüncü şahıs",
      "explanation": "Yazar olayları dışarıdan gözleyen bir anlatıcı kullanır."
    }
  ]
}
```

---

### 4.3 Global Quiz Paylaşımı

#### Tanım

Aynı ISBN'e sahip kitaplar tek bir quiz'e bağlı olmalı.

#### Kurallar

* Quiz sadece **ilk ekleme sırasında** oluşturulur.
* Diğer kullanıcılar aynı kitabı eklerse mevcut quiz gösterilir.
* Quiz tekrar generate edilmez (AI maliyeti azaltılır).

#### DB Constraint

```sql
ALTER TABLE books ADD CONSTRAINT unique_isbn UNIQUE (isbn);
```

---

### 4.4 Hata Yönetimi ve Retry Mekanizması

#### Gereksinimler

* AI JSON döndürmezse:

  * Response schema doğrulanır.
  * Hatalıysa 3 defaya kadar retry yapılır.
  * Yine başarısızsa quiz boş kaydedilir ve "status = pending" flag'i atanır.
* Bu flag'li kayıtlar cron job ile yeniden denenebilir.

---

### 4.5 Güvenlik

* Tüm istekler JWT token ile doğrulanmalı (veya Firebase Auth ID Token).
* OpenAI API anahtarı `.env` dosyasında saklanmalı.
* CORS yönetimi: Sadece Flutter app domain'leri izinli.

---

### 4.6 Performans ve Ölçeklenebilirlik

* Kitap bilgileri ve quiz sonuçları **PostgreSQL JSONB** alanlarında saklanmalı.
* **Redis cache**: Popüler kitapların quiz sonuçlarını cache'le.
* Indexler:

  ```sql
  CREATE INDEX idx_books_isbn ON books(isbn);
  CREATE INDEX idx_quizzes_bookid ON quizzes(book_id);
  ```

---

## 📡 5. API Endpoint Özeti

| Method | Endpoint                                  | Açıklama                                                        |
| ------ | ----------------------------------------- | --------------------------------------------------------------- |
| `GET`  | `/books/search?q={query}&type={type}`     | Hibrit kaynaklardan kitap bilgisi getir (varsa cache, yoksa oluştur) |
| `GET`  | `/books/isbn/:isbn`                       | ISBN ile direkt kitap getir                                      |
| `GET`  | `/books/:id`                              | UUID ile kitap detayını getir                                    |
| `GET`  | `/quiz/:bookId`                           | Kitabın quiz'ini getir                                           |
| `POST` | `/quiz/generate` (internal)               | Kitap bilgisine göre quiz oluştur (background job)               |

**Not:** Kullanıcılar `POST /books/add` kullanmaz. Sistem otomatik olarak `GET /books/search` ile kitap bilgilerini toplar ve cache'ler.

### 5.1 Örnek API Response

#### GET /books/search?q=9780262033848&type=isbn

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
    "description": "A comprehensive textbook covering the full spectrum of modern algorithms...",
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

#### GET /quiz/:bookId

**Response (200 OK):**
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
    "ai_model": "gemini-1.5-flash",
    "created_at": "2025-10-28T10:31:30Z"
  }
}
```

---

## 🧪 6. Test Senaryoları

| ID  | Test                                              | Beklenen Sonuç                                                     |
| --- | ------------------------------------------------- | ------------------------------------------------------------------ |
| T01 | Aynı ISBN ile iki kez sorgu yapılınca             | İkinci istek cache'den döner, API çağrısı yapılmaz, quiz aynı kalır |
| T02 | Google Books API offline                          | OpenLibrary fallback devreye girer, veri yine de gelir             |
| T03 | Her iki harici API offline                        | Hata mesajı döner: "Kitap bilgisi şu an alınamıyor"                |
| T04 | AI quiz generation hatalı JSON döner              | Retry mekanizması çalışır (3 deneme)                               |
| T05 | Hibrit birleştirmede Google Books'ta olmayan alan | OpenLibrary'den tamamlanır (ör. page_count)                        |
| T06 | Quiz henüz oluşturulmamış kitap sorgulanınca      | Kitap bilgisi anında döner, quiz asenkron oluşturulur              |

---

## 🧱 7. Gelecek Aşamalar (Future Scope)

* **Kullanıcı İstatistikleri**: Quiz skorlarını saklama (`user_quizzes` tablosu)
* **Sosyal Özellikler**: Kitaplara yorum ve puanlama sistemi
* **Veri Zenginleştirme**: Daha fazla kitap kaynağı entegrasyonu (Goodreads, Kitapyurdu)
* **Quiz Kalite Kontrolü**: Admin panel üzerinden quiz moderasyonu ve düzenleme
* **Çok Dil Desteği**: Çok dilli quiz üretimi (Türkçe, İngilizce)
* **Akıllı Öneri**: Kitap okuma geçmişine göre AI destekli kitap önerileri

---

## 📅 8. Zaman Planı (Öneri)

| Aşama                                     | Süre  | Açıklama                                                       |
| ----------------------------------------- | ----- | -------------------------------------------------------------- |
| 1️⃣ Temel Go API + DB setup               | 2 gün | Gin, GORM, PostgreSQL yapılandırması                            |
| 2️⃣ Hibrit kitap bilgi servisi            | 5 gün | Google Books + OpenLibrary entegrasyonu, veri birleştirme       |
| 3️⃣ Cache ve tekil kitap sistemi          | 2 gün | ISBN bazlı unique constraint, cache kontrolü                    |
| 4️⃣ AI quiz üretim motoru                 | 4 gün | Google Gemini entegrasyonu, JSON parse, retry mekanizması       |
| 5️⃣ Global quiz paylaşım sistemi          | 2 gün | Quiz-Book ilişkilendirme, asenkron quiz oluşturma               |
| 6️⃣ Test + Hata yönetimi                  | 2 gün | Unit test, integration test, fallback mekanizmaları             |
| 7️⃣ Deployment + Dokümantasyon            | 2 gün | Docker, API dokümantasyonu, Postman collection                  |

**Toplam Süre:** ~19 gün

### Öncelik Sırası:
1. **MVP (Minimum Viable Product)**: Aşama 1-4 (~13 gün)
2. **Production Ready**: Tüm aşamalar (~19 gün)

---

## 🔒 9. Ek Notlar

* **Kod standartları:** GoLint, GoVet, idiomatic Go best practices
* **Deployment:** Docker + systemd veya Kubernetes (ileride)
* **Logging:** Zap veya Logrus
* **Env yönetimi:** `godotenv`
* **Rate limiting:** Fiber middleware (DDOS önleme)

---

## ⚠️ 10. Risk Analizi ve Çözümler

| Risk                                  | Olasılık | Etki  | Çözüm                                                    |
| ------------------------------------- | -------- | ----- | -------------------------------------------------------- |
| Google Books API rate limit           | Yüksek   | Orta  | Redis cache + OpenLibrary fallback                       |
| Gemini API rate limit                 | Düşük    | Orta  | Free tier: 15 req/min, cache sistemi + retry              |
| AI'nın geçersiz JSON döndürmesi       | Orta     | Orta  | Strict JSON schema validation + 3x retry                 |
| Aynı kitabın farklı ISBN'lerle gelmesi| Yüksek   | Düşük | ISBN-10 ve ISBN-13 normalizasyonu                        |
| Kitap bilgisi hiçbir kaynakta yok     | Düşük    | Orta  | 404 Not Found + kullanıcıya "manuel ekle" önerisi       |
| Quiz oluşturma 30+ saniye sürüyor     | Orta     | Yüksek| Asenkron background job, kullanıcı beklemez              |
| Veritabanı performans sorunları       | Düşük    | Yüksek| Index stratejisi (ISBN, BookID), JSONB indexleme         |

---

## 📝 11. Doküman Sürümü

| Versiyon | Tarih      | Değişiklik                                                   | Yazar |
| -------- | ---------- | ------------------------------------------------------------ | ----- |
| 1.0      | 28.10.2025 | İlk PRD oluşturuldu                                           | -     |
| 1.1      | 28.10.2025 | Hibrit sistem odaklı güncelleme - kitap ekleme yerine sorgu  | -     |

---

## 📊 12. Performans Hedefleri

| Metrik                      | Hedef        | Ölçüm Yöntemi                  |
| --------------------------- | ------------ | ------------------------------ |
| Kitap sorgulama süresi      | < 2 saniye   | API response time              |
| Cache hit ratio             | > %70        | Redis/DB metrics               |
| AI quiz oluşturma süresi    | < 30 saniye  | Background job duration        |
| API uptime                  | > %99.5      | Monitoring tools               |
| Eşzamanlı kullanıcı kapasitesi | 100+ user | Load testing                   |
| Database query latency      | < 100ms      | PostgreSQL slow query log      |

---

**Son Güncelleme:** 28 Ekim 2025  
**Doküman Sahibi:** Bookwise Backend Team  
**Durum:** ✅ Onaylandı - Geliştirme Başlayabilir

