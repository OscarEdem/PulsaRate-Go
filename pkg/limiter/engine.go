package limiter

import (
	"context"
	"sync"
	"time"

	pulsaSync "github.com/OscarEdem/PulsaRate-Go/pkg/sync"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

// Engine is the central rate-limiting coordinator managing Tier-1 local buckets and Tier-2 Redis synchronizers.
type Engine struct {
	mu      sync.RWMutex
	buckets map[string]*AtomicBucket
	cfg     Config
	leaser  *pulsaSync.RedisLeaser
}

// NewEngine instantiates a PulsaRate Engine.
func NewEngine(cfg Config, leaser *pulsaSync.RedisLeaser) (*Engine, error) {
	if cfg.Capacity <= 0 || cfg.RefillRate <= 0 {
		return nil, ErrInvalidConfig
	}

	return &Engine{
		buckets: make(map[string]*AtomicBucket),
		cfg:     cfg,
		leaser:  leaser,
	}, nil
}

// Allow evaluates if a single request (cost = 1) is permitted for key.
func (e *Engine) Allow(key string) bool {
	return e.AllowN(key, 1)
}

// AllowN evaluates if a request with token cost N is permitted.
func (e *Engine) AllowN(key string, n int64) bool {
	res := e.Reserve(key, n)
	return res.Allowed
}

// Reserve checks local token availability, triggering Tier-2 lease refill if required.
func (e *Engine) Reserve(key string, n int64) Result {
	bucket := e.getOrCreateBucket(key)

	allowed := bucket.AllowN(n)
	remaining := bucket.Tokens()

	// If local tokens drop below batchSize and leaser is configured, trigger async background refill
	if remaining < e.cfg.BatchSize && e.leaser != nil {
		go e.asyncLeaseRefill(key, bucket)
	}

	return Result{
		Allowed:   allowed,
		Remaining: remaining,
		ResetIn:   time.Second,
	}
}

func (e *Engine) getOrCreateBucket(key string) *AtomicBucket {
	e.mu.RLock()
	bucket, exists := e.buckets[key]
	e.mu.RUnlock()

	if exists {
		return bucket
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	bucket, exists = e.buckets[key]
	if exists {
		return bucket
	}

	bucket, _ = NewAtomicBucket(key, e.cfg)
	e.buckets[key] = bucket

	return bucket
}

func (e *Engine) asyncLeaseRefill(key string, bucket *AtomicBucket) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	granted, err := e.leaser.RequestLease(ctx, key, e.cfg.BatchSize, e.cfg.Capacity, e.cfg.RefillRate)
	if err == nil && granted > 0 {
		bucket.AddTokens(granted)
	}
}
