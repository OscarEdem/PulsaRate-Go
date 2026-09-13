package limiter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

func TestAtomicBucket_BasicAllow(t *testing.T) {
	cfg := Config{
		Capacity:   5,
		RefillRate: 10,
	}

	b, err := NewAtomicBucket("test_client", cfg)
	if err != nil {
		t.Fatalf("unexpected error creating bucket: %v", err)
	}

	// Consume 5 allowed tokens
	for i := 0; i < 5; i++ {
		if !b.Allow() {
			t.Errorf("expected request %d to be allowed", i+1)
		}
	}

	// 6th request should be rejected (depleted)
	if b.Allow() {
		t.Errorf("expected 6th request to be denied")
	}
}

func TestAtomicBucket_ConcurrentSafety(t *testing.T) {
	cfg := Config{
		Capacity:   1000,
		RefillRate: 1, // Slow refill so capacity dominates
	}

	b, err := NewAtomicBucket("concurrent_client", cfg)
	if err != nil {
		t.Fatalf("unexpected error creating bucket: %v", err)
	}

	var allowedCount int64
	var wg sync.WaitGroup
	workers := 50
	requestsPerWorker := 40 // Total 2000 requests against capacity 1000

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				if b.Allow() {
					syncAdd(&allowedCount, 1)
				}
			}
		}()
	}

	wg.Wait()

	if allowedCount > 1005 { // Allow slight buffer for minor elapsed refill
		t.Errorf("allowed %d requests, expected <= 1005 for capacity 1000", allowedCount)
	}
}

func TestAtomicBucket_Refill(t *testing.T) {
	cfg := Config{
		Capacity:   2,
		RefillRate: 20, // 20 tokens per sec -> 2 tokens per 100ms
	}

	b, err := NewAtomicBucket("refill_client", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Drain bucket
	b.Allow()
	b.Allow()
	if b.Allow() {
		t.Errorf("expected bucket to be empty")
	}

	// Wait 150ms for refill
	time.Sleep(150 * time.Millisecond)

	if !b.Allow() {
		t.Errorf("expected bucket to have refilled tokens")
	}
}

func TestLeakyBucket(t *testing.T) {
	lb, err := NewLeakyBucket(3, 100)
	if err != nil {
		t.Fatalf("failed to create leaky bucket: %v", err)
	}

	for i := 0; i < 3; i++ {
		if !lb.Allow() {
			t.Errorf("expected request %d to be allowed by leaky bucket", i+1)
		}
	}

	if lb.Allow() {
		t.Errorf("expected 4th request to exceed capacity")
	}
}

func TestSlidingWindowLog(t *testing.T) {
	swl, err := NewSlidingWindowLog(2, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create sliding window log: %v", err)
	}

	for i := 0; i < 2; i++ {
		if !swl.Allow() {
			t.Errorf("expected request %d to be allowed within sliding window", i+1)
		}
	}

	if swl.Allow() {
		t.Errorf("expected 3rd request to be rejected")
	}
}

func BenchmarkAtomicBucket_Allow(b *testing.B) {
	cfg := Config{
		Capacity:   int64(b.N + 1000),
		RefillRate: 1000000,
	}
	bucket, _ := NewAtomicBucket("bench_client", cfg)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bucket.Allow()
	}
}

func syncAdd(addr *int64, delta int64) {
	for {
		curr := atomic.LoadInt64(addr)
		if atomic.CompareAndSwapInt64(addr, curr, curr+delta) {
			return
		}
	}
}
