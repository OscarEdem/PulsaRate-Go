package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OscarEdem/PulsaRate-Go/pkg/limiter"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

func TestRateLimitMiddleware_AllowAndDeny(t *testing.T) {
	cfg := limiter.Config{
		Capacity:   2,
		RefillRate: 1,
	}

	engine, err := limiter.NewEngine(cfg, nil)
	if err != nil {
		t.Fatalf("unexpected error creating engine: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	ts := httptest.NewServer(RateLimit(engine)(handler))
	defer ts.Close()

	client := ts.Client()

	// 1st Request: Should succeed
	resp1, err := client.Get(ts.URL)
	if err != nil || resp1.StatusCode != http.StatusOK {
		t.Errorf("request 1 failed, status: %v", resp1.StatusCode)
	}

	// 2nd Request: Should succeed
	resp2, err := client.Get(ts.URL)
	if err != nil || resp2.StatusCode != http.StatusOK {
		t.Errorf("request 2 failed, status: %v", resp2.StatusCode)
	}

	// 3rd Request: Should be throttled with 429
	resp3, err := client.Get(ts.URL)
	if err != nil || resp3.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %v", resp3.StatusCode)
	}

	if resp3.Header.Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header to be present")
	}
}
