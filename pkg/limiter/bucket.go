package limiter

import (
	"sync/atomic"
	"time"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

// AtomicBucket is a lock-free, zero-allocation token bucket implementation.
// All operations are executed atomically using CPU atomic primitives.
type AtomicBucket struct {
	key        string
	capacity   int64
	tokens     int64 // Manipulated via sync/atomic
	refillRate int64 // Tokens per second
	lastRefill int64 // Unix nanoseconds (sync/atomic)
	batchSize  int64
}

// NewAtomicBucket instantiates a lock-free atomic token bucket.
func NewAtomicBucket(key string, cfg Config) (*AtomicBucket, error) {
	if cfg.Capacity <= 0 || cfg.RefillRate <= 0 {
		return nil, ErrInvalidConfig
	}

	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = cfg.Capacity / 2
		if batchSize < 1 {
			batchSize = 1
		}
	}

	now := time.Now().UnixNano()
	b := &AtomicBucket{
		key:        key,
		capacity:   cfg.Capacity,
		tokens:     cfg.Capacity,
		refillRate: cfg.RefillRate,
		lastRefill: now,
		batchSize:  batchSize,
	}

	return b, nil
}

// Allow evaluates if a single token (cost = 1) can be consumed.
func (b *AtomicBucket) Allow() bool {
	return b.AllowN(1)
}

// AllowN evaluates if N tokens can be consumed atomically.
func (b *AtomicBucket) AllowN(n int64) bool {
	if n <= 0 {
		return true
	}

	b.refill()

	for {
		curr := atomic.LoadInt64(&b.tokens)
		if curr < n {
			return false
		}
		if atomic.CompareAndSwapInt64(&b.tokens, curr, curr-n) {
			return true
		}
	}
}

// AddTokens adds tokens atomically to the local reservoir (used by Tier-2 Redis leaser).
func (b *AtomicBucket) AddTokens(n int64) int64 {
	if n <= 0 {
		return atomic.LoadInt64(&b.tokens)
	}

	for {
		curr := atomic.LoadInt64(&b.tokens)
		next := curr + n
		if next > b.capacity {
			next = b.capacity
		}
		if atomic.CompareAndSwapInt64(&b.tokens, curr, next) {
			return next
		}
	}
}

// Tokens returns the current available token count.
func (b *AtomicBucket) Tokens() int64 {
	b.refill()
	return atomic.LoadInt64(&b.tokens)
}

// Capacity returns the maximum capacity of the bucket.
func (b *AtomicBucket) Capacity() int64 {
	return b.capacity
}

// refill calculates and adds tokens based on elapsed nanoseconds.
func (b *AtomicBucket) refill() {
	now := time.Now().UnixNano()
	last := atomic.LoadInt64(&b.lastRefill)

	elapsed := now - last
	if elapsed <= 0 {
		return
	}

	// Calculate tokens earned: (elapsed_ns * refillRate) / 1e9
	deltaTokens := (elapsed * b.refillRate) / int64(time.Second)
	if deltaTokens <= 0 {
		return
	}

	// Attempt to update lastRefill timestamp atomically
	if atomic.CompareAndSwapInt64(&b.lastRefill, last, now) {
		b.AddTokens(deltaTokens)
	}
}
