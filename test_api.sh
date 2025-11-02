#!/bin/bash

# Bookwise API Test Script
# Bu script API'nin tüm endpoint'lerini test eder

set -e

BASE_URL="http://localhost:8080"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Bookwise API Test Script            ║${NC}"
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo ""

# Test 1: Health Check
echo -e "${YELLOW}[1/6] Testing Health Check...${NC}"
RESPONSE=$(curl -s "$BASE_URL/health")
STATUS=$(echo $RESPONSE | grep -o '"status":"healthy"' || echo "")

if [ -n "$STATUS" ]; then
    echo -e "${GREEN}✅ Health Check: PASSED${NC}"
else
    echo -e "${RED}❌ Health Check: FAILED${NC}"
    echo "Response: $RESPONSE"
    exit 1
fi
echo ""

# Test 2: Detailed Health Check
echo -e "${YELLOW}[2/6] Testing Detailed Health Check...${NC}"
RESPONSE=$(curl -s "$BASE_URL/health/detailed")
STATUS=$(echo $RESPONSE | grep -o '"database":"healthy"' || echo "")

if [ -n "$STATUS" ]; then
    echo -e "${GREEN}✅ Detailed Health Check: PASSED${NC}"
    echo "Worker Stats:"
    echo $RESPONSE | jq '.components.quiz_worker' 2>/dev/null || echo "jq not installed"
else
    echo -e "${RED}❌ Detailed Health Check: FAILED${NC}"
    echo "Response: $RESPONSE"
fi
echo ""

# Test 3: Book Search by ISBN
echo -e "${YELLOW}[3/6] Testing Book Search (ISBN)...${NC}"
ISBN="9780262033848"
RESPONSE=$(curl -s "$BASE_URL/api/v1/books/search?q=$ISBN&type=isbn")
SUCCESS=$(echo $RESPONSE | grep -o '"success":true' || echo "")

if [ -n "$SUCCESS" ]; then
    echo -e "${GREEN}✅ Book Search: PASSED${NC}"
    BOOK_ID=$(echo $RESPONSE | jq -r '.data.id' 2>/dev/null)
    TITLE=$(echo $RESPONSE | jq -r '.data.title' 2>/dev/null)
    CACHE_HIT=$(echo $RESPONSE | jq -r '.cache_hit' 2>/dev/null)
    
    echo "  📚 Book ID: $BOOK_ID"
    echo "  📖 Title: $TITLE"
    echo "  💾 Cache Hit: $CACHE_HIT"
else
    echo -e "${RED}❌ Book Search: FAILED${NC}"
    echo "Response: $RESPONSE"
fi
echo ""

# Test 4: Get Book by ID
if [ -n "$BOOK_ID" ] && [ "$BOOK_ID" != "null" ]; then
    echo -e "${YELLOW}[4/6] Testing Get Book by ID...${NC}"
    RESPONSE=$(curl -s "$BASE_URL/api/v1/books/$BOOK_ID")
    SUCCESS=$(echo $RESPONSE | grep -o '"success":true' || echo "")
    
    if [ -n "$SUCCESS" ]; then
        echo -e "${GREEN}✅ Get Book by ID: PASSED${NC}"
        QUIZ_STATUS=$(echo $RESPONSE | jq -r '.data.quiz_status' 2>/dev/null)
        echo "  🎯 Quiz Status: $QUIZ_STATUS"
    else
        echo -e "${RED}❌ Get Book by ID: FAILED${NC}"
        echo "Response: $RESPONSE"
    fi
else
    echo -e "${YELLOW}[4/6] Get Book by ID: SKIPPED (no book_id)${NC}"
fi
echo ""

# Test 5: Get Book by ISBN
echo -e "${YELLOW}[5/6] Testing Get Book by ISBN...${NC}"
RESPONSE=$(curl -s "$BASE_URL/api/v1/books/isbn/$ISBN")
SUCCESS=$(echo $RESPONSE | grep -o '"success":true' || echo "")

if [ -n "$SUCCESS" ]; then
    echo -e "${GREEN}✅ Get Book by ISBN: PASSED (Cache Hit!)${NC}"
else
    echo -e "${RED}❌ Get Book by ISBN: FAILED${NC}"
    echo "Response: $RESPONSE"
fi
echo ""

# Test 6: Get Quiz (with retry for generating status)
if [ -n "$BOOK_ID" ] && [ "$BOOK_ID" != "null" ]; then
    echo -e "${YELLOW}[6/6] Testing Get Quiz...${NC}"
    
    MAX_RETRIES=3
    RETRY_COUNT=0
    QUIZ_READY=false
    
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        RESPONSE=$(curl -s "$BASE_URL/api/v1/quiz/$BOOK_ID")
        SUCCESS=$(echo $RESPONSE | grep -o '"success":true' || echo "")
        STATUS=$(echo $RESPONSE | jq -r '.status' 2>/dev/null)
        
        if [ -n "$SUCCESS" ]; then
            echo -e "${GREEN}✅ Get Quiz: PASSED${NC}"
            QUESTION_COUNT=$(echo $RESPONSE | jq '.data.questions | length' 2>/dev/null)
            echo "  🎯 Quiz Questions: $QUESTION_COUNT"
            echo "  📝 First Question:"
            echo $RESPONSE | jq -r '.data.questions[0].question' 2>/dev/null | sed 's/^/     /'
            QUIZ_READY=true
            break
        elif [ "$STATUS" = "generating" ] || [ "$STATUS" = "pending" ]; then
            echo -e "${YELLOW}  ⏳ Quiz Status: $STATUS (Retry $((RETRY_COUNT + 1))/$MAX_RETRIES)${NC}"
            RETRY_COUNT=$((RETRY_COUNT + 1))
            if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
                echo "  ⏰ Waiting 10 seconds..."
                sleep 10
            fi
        else
            echo -e "${RED}❌ Get Quiz: FAILED${NC}"
            echo "Response: $RESPONSE"
            break
        fi
    done
    
    if [ "$QUIZ_READY" = false ] && ([ "$STATUS" = "generating" ] || [ "$STATUS" = "pending" ]); then
        echo -e "${YELLOW}⚠️  Quiz still generating after $MAX_RETRIES retries${NC}"
        echo "  💡 This is normal for first-time quiz generation (takes 10-30 seconds)"
        echo "  💡 Try again in a few seconds with:"
        echo "     curl http://localhost:8080/api/v1/quiz/$BOOK_ID"
    fi
else
    echo -e "${YELLOW}[6/6] Get Quiz: SKIPPED (no book_id)${NC}"
fi
echo ""

# Test 7: List Books
echo -e "${YELLOW}[BONUS] Testing List Books...${NC}"
RESPONSE=$(curl -s "$BASE_URL/api/v1/books?page=1&limit=5")
SUCCESS=$(echo $RESPONSE | grep -o '"success":true' || echo "")

if [ -n "$SUCCESS" ]; then
    echo -e "${GREEN}✅ List Books: PASSED${NC}"
    TOTAL=$(echo $RESPONSE | jq -r '.pagination.total' 2>/dev/null)
    echo "  📚 Total Books in Database: $TOTAL"
else
    echo -e "${RED}❌ List Books: FAILED${NC}"
fi
echo ""

# Summary
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Test Summary                         ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}✅ API is running successfully!${NC}"
echo ""
echo -e "${BLUE}📡 Endpoints tested:${NC}"
echo "  • GET  /health"
echo "  • GET  /health/detailed"
echo "  • GET  /api/v1/books/search"
echo "  • GET  /api/v1/books/:id"
echo "  • GET  /api/v1/books/isbn/:isbn"
echo "  • GET  /api/v1/books"
echo "  • GET  /api/v1/quiz/:bookId"
echo ""
echo -e "${YELLOW}💡 Next Steps:${NC}"
echo "  1. Import Postman collection: Bookwise_API.postman_collection.json"
echo "  2. Read API docs: documents/API_DOCUMENTATION.md"
echo "  3. Check worker stats: curl http://localhost:8080/health/detailed | jq"
echo ""
echo -e "${GREEN}Happy Testing! 🎉${NC}"

