package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aeroforge-labs/PulsaRate-Go/pkg/limiter"
	"github.com/aeroforge-labs/PulsaRate-Go/pkg/middleware"
	"github.com/aeroforge-labs/PulsaRate-Go/pkg/observability"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

func main() {
	cfg := limiter.Config{
		Capacity:   10,
		RefillRate: 2, // 2 tokens/sec
		BatchSize:  5,
	}

	engine, err := limiter.NewEngine(cfg, nil)
	if err != nil {
		log.Fatalf("failed to initialize PulsaRate engine: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/resource", func(w http.ResponseWriter, r *http.Request) {
		observability.DefaultMetrics.IncAllowed()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success","message":"Resource accessed successfully"}`))
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		allowed, rejected, leases := observability.DefaultMetrics.Stats()
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(fmt.Sprintf(
			"# HELP rate_limit_allowed_total Total allowed requests\n"+
				"# TYPE rate_limit_allowed_total counter\n"+
				"rate_limit_allowed_total %d\n\n"+
				"# HELP rate_limit_rejected_total Total throttled requests\n"+
				"# TYPE rate_limit_rejected_total counter\n"+
				"rate_limit_rejected_total %d\n\n"+
				"# HELP redis_leases_total Total Redis lease sync calls\n"+
				"# TYPE redis_leases_total counter\n"+
				"redis_leases_total %d\n",
			allowed, rejected, leases,
		)))
	})

	rateLimitedHandler := middleware.RateLimit(engine)(mux)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      rateLimitedHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("==================================================")
	fmt.Println("⚡ PulsaRate-Go Proxy Server Running on :8080")
	fmt.Println("   API Endpoint : http://localhost:8080/api/v1/resource")
	fmt.Println("   Metrics      : http://localhost:8080/metrics")
	fmt.Println("==================================================")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failure: %v", err)
	}
}
