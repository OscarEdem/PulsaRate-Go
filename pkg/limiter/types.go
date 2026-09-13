package limiter

import (
	"errors"
	"time"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

var (
	// ErrRateLimited is returned when a request exceeds allowed quota.
	ErrRateLimited = errors.New("rate limit exceeded")

	// ErrInvalidConfig is returned when limiter parameters are misconfigured.
	ErrInvalidConfig = errors.New("invalid limiter configuration")
)

// Limiter defines the core interface for rate limiting evaluation.
type Limiter interface {
	// Allow evaluates if a single request (cost = 1) is allowed.
	Allow(key string) bool

	// AllowN evaluates if a request with specific token cost is allowed.
	AllowN(key string, n int64) bool

	// Reserve checks availability and reserves tokens, returning a Result.
	Reserve(key string, n int64) Result
}

// Result captures the evaluation state of a rate limit check.
type Result struct {
	Allowed   bool          `json:"allowed"`
	Remaining int64         `json:"remaining"`
	ResetIn   time.Duration `json:"reset_in"`
}

// Config defines initialization parameters for a LocalBucket.
type Config struct {
	Capacity   int64         `json:"capacity"`     // Maximum token capacity (burst limit)
	RefillRate int64         `json:"refill_rate"`  // Tokens added per second
	BatchSize  int64         `json:"batch_size"`   // Lease batch reservation size for Tier-2 sync
	Window     time.Duration `json:"window"`       // Time window (default 1s)
}
