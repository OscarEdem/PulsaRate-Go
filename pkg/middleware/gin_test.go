package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aeroforge-labs/PulsaRate-Go/pkg/limiter"
	"github.com/gin-gonic/gin"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

func TestGinRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := limiter.Config{
		Capacity:   2,
		RefillRate: 1,
	}

	engine, err := limiter.NewEngine(cfg, nil)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	r := gin.New()
	r.Use(GinRateLimit(engine))
	r.GET("/api/v1/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 1st Request: PASS
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api/v1/test", nil)
	req1.RemoteAddr = "127.0.0.1:1234"
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("request 1 failed, expected 200 got %d", w1.Code)
	}

	// 2nd Request: PASS
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/test", nil)
	req2.RemoteAddr = "127.0.0.1:1234"
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("request 2 failed, expected 200 got %d", w2.Code)
	}

	// 3rd Request: FAIL (HTTP 429)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/api/v1/test", nil)
	req3.RemoteAddr = "127.0.0.1:1234"
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("request 3 expected 429, got %d", w3.Code)
	}

	if w3.Header().Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header in 429 response")
	}
}
