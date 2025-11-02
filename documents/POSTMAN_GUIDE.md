# 📮 Postman Collection Guide

Bookwise API için Postman collection kullanım kılavuzu.

## 📥 Import Etme

### 1. Postman'ı Açın

Postman uygulamasını başlatın (yoksa [indir](https://www.postman.com/downloads/))

### 2. Collection'ı Import Edin

**Yöntem 1: File Import**
```
1. Postman'da "Import" butonuna tıklayın
2. "Upload Files" sekmesini seçin
3. "Bookwise_API.postman_collection.json" dosyasını seçin
4. "Import" butonuna tıklayın
```

**Yöntem 2: Drag & Drop**
```
Bookwise_API.postman_collection.json dosyasını
Postman penceresine sürükleyip bırakın
```

### 3. Environment'ı Import Edin

```
1. Postman'da "Import" butonuna tıklayın
2. "Bookwise_API.postman_environment.json" dosyasını import edin
3. Sağ üstten "Bookwise API - Local" environment'ını seçin
```

## 🚀 Hızlı Başlangıç

### 1. API'yi Çalıştırın

```bash
cd bookwise_api
docker-compose up -d
# veya
go run cmd/server/main.go
```

### 2. Health Check

```
1. Collection'da "Health Check" klasörünü açın
2. "Basic Health Check" isteğini seçin
3. "Send" butonuna tıklayın
4. ✅ "status": "healthy" görmelisiniz
```

### 3. İlk Kitap Araması

```
1. "Books" klasöründen "Search Book by ISBN" isteğini seçin
2. "Send" butonuna tıklayın
3. Response'da kitap bilgilerini göreceksiniz
4. Response'dan book_id'yi kopyalayın
```

### 4. Quiz Sorgulama

```
1. "Quiz" klasöründen "Get Quiz by Book ID" isteğini seçin
2. URL'deki :bookId değerini kopyaladığınız ID ile değiştirin
3. "Send" butonuna tıklayın
4. Quiz henüz hazır değilse "generating" mesajı alırsınız
5. 10-30 saniye sonra tekrar deneyin
```

## 📂 Collection Yapısı

```
Bookwise API
├── Health Check
│   ├── Basic Health Check
│   └── Detailed Health Check
├── Books
│   ├── Search Book by ISBN
│   ├── Search Book by Title
│   ├── Search Book by Author
│   ├── Get Book by ID
│   ├── Get Book by ISBN
│   └── List Books
└── Quiz
    ├── Get Quiz by Book ID
    └── Get Quiz by Quiz ID
```

## 🔧 Environment Variables

Collection'da kullanılan değişkenler:

| Variable | Default Value | Description |
|----------|--------------|-------------|
| `base_url` | `http://localhost:8080` | API base URL |
| `book_id` | `""` | Book UUID (manuel set edilir) |
| `quiz_id` | `""` | Quiz UUID (manuel set edilir) |

### Environment'ı Değiştirme

**Local Development:**
```json
{
  "base_url": "http://localhost:8080"
}
```

**Docker:**
```json
{
  "base_url": "http://localhost:8080"
}
```

**Production:**
```json
{
  "base_url": "https://api.bookwise.com"
}
```

## 📝 Örnek İş Akışları

### Workflow 1: Yeni Kitap Arama ve Quiz

```
1. Search Book by ISBN
   → q: 9780262033848
   → type: isbn
   → Response'dan book_id'yi al

2. Get Book by ID
   → :id parametresine book_id'yi yapıştır
   → quiz_status'u kontrol et

3. Get Quiz by Book ID (quiz_status="completed" ise)
   → :bookId parametresine book_id'yi yapıştır
   → Quiz sorularını gör
```

### Workflow 2: Mevcut Kitap Sorgulama

```
1. Get Book by ISBN
   → :isbn parametresine ISBN'i gir
   → Kitap varsa direkt döner
   → Yoksa 404 alırsınız, /search kullanın

2. Get Quiz by Book ID
   → book_id'yi kullanarak quiz'i çek
```

### Workflow 3: Tüm Kitapları Listeleme

```
1. List Books
   → page: 1
   → limit: 10
   → Pagination bilgilerini gör

2. Sayfa değiştir
   → page: 2
   → limit: 20
```

## 🧪 Test Senaryoları

### Test 1: ISBN ile Kitap Bulma

```
Request: Search Book by ISBN
Query: q=9780262033848&type=isbn
Expected: 200 OK, cache_hit: false (ilk arama)
         200 OK, cache_hit: true (ikinci arama)
```

### Test 2: Geçersiz ISBN

```
Request: Search Book by ISBN
Query: q=9999999999999&type=isbn
Expected: 404 Not Found
```

### Test 3: Quiz Oluşturma Süreci

```
1. Search Book → quiz_status: "pending"
2. 5 saniye sonra Get Quiz → status: "generating"
3. 30 saniye sonra Get Quiz → success: true + sorular
```

### Test 4: Pagination

```
Request: List Books
Query: page=1&limit=5
Expected: 5 kitap + pagination info

Query: page=2&limit=5
Expected: Sonraki 5 kitap
```

## 🎨 Response Örnekleri

Collection'daki her request için örnek response'lar hazır:
- ✅ Success responses
- ❌ Error responses
- 🔄 Different status states

## 💡 İpuçları

### Otomatik Environment Variable Set

Request'lerden dönen değerleri otomatik olarak environment'a kaydetmek için:

**Tests** sekmesinde:
```javascript
// Book search sonrası
var jsonData = pm.response.json();
if (jsonData.success && jsonData.data.id) {
    pm.environment.set("book_id", jsonData.data.id);
}

// Quiz get sonrası
var jsonData = pm.response.json();
if (jsonData.success && jsonData.data.id) {
    pm.environment.set("quiz_id", jsonData.data.id);
}
```

### Response Validasyonu

Otomatik test eklemek için **Tests** sekmesi:

```javascript
// Status code kontrolü
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

// JSON format kontrolü
pm.test("Response is JSON", function () {
    pm.response.to.be.json;
});

// Success field kontrolü
pm.test("Success is true", function () {
    var jsonData = pm.response.json();
    pm.expect(jsonData.success).to.eql(true);
});
```

## 🔄 Collection Runner

Tüm testleri otomatik çalıştırmak için:

```
1. Collection'a sağ tıklayın
2. "Run collection" seçin
3. İstediğiniz request'leri seçin
4. "Run Bookwise API" butonuna tıklayın
5. Tüm request'ler sırayla çalışacak
```

## 📊 Monitoring

Postman Monitoring ile API'yi periyodik olarak test edebilirsiniz:

```
1. Collection'a sağ tıklayın
2. "Monitor collection" seçin
3. Test sıklığını ayarlayın (örn: her 5 dakika)
4. Email bildirimleri aktif edin
```

## 🐛 Troubleshooting

### Problem: Connection refused

```
✅ Çözüm:
- API'nin çalıştığından emin olun: curl http://localhost:8080/health
- Port'un doğru olduğunu kontrol edin
- Docker container'ın ayakta olduğunu kontrol edin: docker ps
```

### Problem: 404 Not Found

```
✅ Çözüm:
- URL'in doğru olduğunu kontrol edin
- base_url environment variable'ını kontrol edin
- API endpoint'in doğru olduğunu doğrulayın
```

### Problem: Timeout

```
✅ Çözüm:
- Postman settings'den timeout süresini artırın (Settings > General > Request timeout)
- API response time'ını kontrol edin: /health/detailed
- Database connection'ı kontrol edin
```

## 📦 Export & Share

### Collection'ı Export Etme

```
1. Collection'a sağ tıklayın
2. "Export" seçin
3. Collection v2.1 (recommended) seçin
4. JSON dosyasını kaydedin
```

### Team ile Paylaşma

```
1. Collection'ı Postman workspace'e publish edin
2. Workspace'e team üyelerini invite edin
3. Collection otomatik olarak senkronize olur
```

## 🔗 Faydalı Linkler

- [Postman Documentation](https://learning.postman.com/docs/getting-started/introduction/)
- [API Documentation](API_DOCUMENTATION.md)
- [README](../README.md)

---

**Son Güncelleme:** 28 Ekim 2025

Happy Testing! 📮✨

