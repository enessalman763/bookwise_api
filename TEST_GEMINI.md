# 🧪 Gemini Flash Test Rehberi

## 1️⃣ Gemini API Key Alma (Ücretsiz!)

### Adımlar:
1. **Google AI Studio**'ya git: https://aistudio.google.com/app/apikey
2. **"Create API Key"** butonuna tıkla
3. API key'i kopyala (örnek: `AIzaSyXXXXXXXXXXXXXXXXXXXXXXXXXX`)

### Limitler (Ücretsiz):
- ✅ **15 request/dakika**
- ✅ **1 million token/dakika**
- ✅ **1500 request/gün**

---

## 2️⃣ .env Dosyasını Güncelle

```bash
# .env dosyasını düzenle
nano .env

# veya
code .env
```

Bu satırı değiştir:
```bash
# ÖNCE:
GEMINI_API_KEY=test-key-replace-this

# SONRA:
GEMINI_API_KEY=AIzaSyXXXXXXXXXXXXXXXXXXXXXXXXXX  # Gerçek key'ini yapıştır
```

---

## 3️⃣ Container'ları Yeniden Başlat

```bash
docker-compose down
docker-compose up -d
```

---

## 4️⃣ Test Et!

### Test 1: Yeni Kitap Ara (Quiz Oluşturma)

```bash
# 1984 kitabını ara (quiz generate edilecek)
curl "http://localhost:8080/api/v1/books/search?q=9780451524935&type=isbn" | jq '.'
```

**Beklenen:**
```json
{
  "success": true,
  "data": {
    "title": "1984",
    "authors": ["George Orwell"],
    "quiz_status": "pending"
  },
  "message": "Kitap başarıyla getirildi. Quiz oluşturuluyor..."
}
```

### Test 2: Quiz'i Kontrol Et

```bash
# Kitap ID'sini al
BOOK_ID=$(curl -s "http://localhost:8080/api/v1/books/search?q=9780451524935&type=isbn" | jq -r '.data.id')

# 5-10 saniye bekle (quiz generate olsun)
sleep 10

# Quiz'i getir
curl "http://localhost:8080/api/v1/quiz/$BOOK_ID" | jq '.'
```

**Beklenen:**
```json
{
  "success": true,
  "data": {
    "quiz": [
      {
        "question": "What is the main character's name in 1984?",
        "options": ["Winston Smith", "Julia", "O'Brien", "Big Brother"],
        "answer": "Winston Smith",
        "explanation": "..."
      }
    ]
  }
}
```

### Test 3: Logları İzle

```bash
# Quiz generation sürecini izle
docker-compose logs -f api
```

**Başarılı log örneği:**
```
✅ Generating quiz with Gemini: gemini-1.5-flash
✅ Quiz generated successfully by Gemini in 2.3s
✅ Quiz saved to database: 1984 (quiz_id: xxx)
```

---

## 5️⃣ Hata Durumları

### Hata 1: "API key not valid"
```bash
# Çözüm: Gemini API key'ini kontrol et
echo $GEMINI_API_KEY
```

### Hata 2: "Resource exhausted"
```bash
# Çözüm: Rate limit aşıldı, 1 dakika bekle
# Ücretsiz plan: 15 req/min
```

### Hata 3: Quiz "failed" durumunda
```bash
# Retry mekanizması devreye girecek
# Veya manuel retry:
curl -X POST "http://localhost:8080/api/v1/admin/retry-failed-quizzes"
```

---

## 6️⃣ Sistem Durumu Kontrolü

```bash
# Genel durum
curl http://localhost:8080/health/detailed | jq '.'

# Quiz istatistikleri
curl http://localhost:8080/health/detailed | jq '.components.quiz_worker'
```

**Sağlıklı çıktı:**
```json
{
  "completed": 5,
  "failed": 0,
  "generating": 0,
  "pending": 0,
  "worker_running": true
}
```

---

## 🎯 Başarı Kriterleri

✅ **Adım 1**: Gemini API key alındı  
✅ **Adım 2**: .env dosyası güncellendi  
✅ **Adım 3**: Container'lar yeniden başlatıldı  
✅ **Adım 4**: Yeni kitap arandı  
✅ **Adım 5**: Quiz başarıyla oluşturuldu  
✅ **Adım 6**: API endpoint'leri çalışıyor  

---

## 🚀 Hızlı Test Scripti

```bash
#!/bin/bash

echo "🧪 Gemini Flash Test Başlatılıyor..."

# 1. Yeni kitap ekle
echo "📚 1984 kitabını ekliyorum..."
RESPONSE=$(curl -s "http://localhost:8080/api/v1/books/search?q=9780451524935&type=isbn")
echo $RESPONSE | jq '.'

BOOK_ID=$(echo $RESPONSE | jq -r '.data.id')

# 2. Quiz oluşmasını bekle
echo "⏳ Quiz oluşması bekleniyor (10 saniye)..."
sleep 10

# 3. Quiz'i getir
echo "📝 Quiz getiriliyor..."
curl -s "http://localhost:8080/api/v1/quiz/$BOOK_ID" | jq '.'

echo "✅ Test tamamlandı!"
```

**Kullanım:**
```bash
chmod +x test_gemini.sh
./test_gemini.sh
```

---

## 📊 Performans Metrikleri

### Gemini Flash (gemini-1.5-flash)
- **Ortalama Yanıt Süresi**: 1-3 saniye
- **Token/Quiz**: ~500-1000 tokens
- **Maliyet**: ✨ **ÜCRETSIZ** (ücretsiz plan dahilinde)

### OpenAI Comparison
- **GPT-4o-mini**: ~$0.15/1M input tokens
- **Gemini Flash**: **$0** (ücretsiz plan)
- **Hız**: Gemini ~%30 daha hızlı

---

## 🆘 Destek

Sorun yaşarsan:
1. Logları kontrol et: `docker-compose logs api`
2. API key'i kontrol et: Gerçek Gemini key'i kullanıyor musun?
3. Rate limit kontrolü: 15 req/min limiti aşıldı mı?

**Daha fazla bilgi**: `README.md` ve `START_GUIDE.md` dosyalarına bak.

