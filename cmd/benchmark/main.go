package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aeroforge-labs/PulsaRate-Go/pkg/limiter"
	"github.com/aeroforge-labs/PulsaRate-Go/pkg/middleware"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

func main() {
	targetURL := "http://localhost:8080/api/v1/resource"
	totalRequests := 50000
	concurrency := 100

	// Check if server is already running on :8080, if not spin up an in-memory test server
	var testServer *httptest.Server
	client := &http.Client{Timeout: 500 * time.Millisecond}
	_, err := client.Get(targetURL)
	if err != nil {
		fmt.Println("No external server detected on :8080. Initializing internal high-speed PulsaRate server...")
		
		cfg := limiter.Config{
			Capacity:   10,
			RefillRate: 2,
			BatchSize:  5,
		}
		engine, _ := limiter.NewEngine(cfg, nil)
		
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/resource", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		})
		
		testServer = httptest.NewServer(middleware.RateLimit(engine)(mux))
		defer testServer.Close()
		targetURL = testServer.URL + "/api/v1/resource"
	}

	fmt.Printf("\nStarting PulsaRate Load Test & Efficiency Proof...\n")
	fmt.Printf("  Target URL     : %s\n", targetURL)
	fmt.Printf("  Total Requests : %d\n", totalRequests)
	fmt.Printf("  Concurrency    : %d goroutines\n", concurrency)
	fmt.Println("--------------------------------------------------")

	loadClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: concurrency,
			MaxConnsPerHost:     concurrency,
		},
		Timeout: 5 * time.Second,
	}

	var (
		allowedCount   int64
		throttledCount int64
		errorCount     int64
	)

	requestsPerWorker := totalRequests / concurrency
	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				resp, err := loadClient.Get(targetURL)
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					continue
				}

				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()

				switch resp.StatusCode {
				case http.StatusOK:
					atomic.AddInt64(&allowedCount, 1)
				case http.StatusTooManyRequests:
					atomic.AddInt64(&throttledCount, 1)
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)

	rps := float64(totalRequests) / duration.Seconds()
	avgLatency := duration / time.Duration(totalRequests)

	fmt.Println("\n==================================================")
	fmt.Println("  ⚡ PULSARATE-GO EFFICIENCY PROOF RESULTS")
	fmt.Println("==================================================")
	fmt.Printf("  Duration             : %v\n", duration)
	fmt.Printf("  Throughput           : %.2f Requests/Sec\n", rps)
	fmt.Printf("  Average Latency      : %v per request\n", avgLatency)
	fmt.Printf("  Allowed (200 OK)     : %d\n", atomic.LoadInt64(&allowedCount))
	fmt.Printf("  Throttled (429)      : %d\n", atomic.LoadInt64(&throttledCount))
	fmt.Printf("  Network Errors       : %d\n", atomic.LoadInt64(&errorCount))
	fmt.Println("==================================================")
}
