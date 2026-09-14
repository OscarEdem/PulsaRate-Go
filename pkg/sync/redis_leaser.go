package sync

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

var (
	ErrCircuitOpen = errors.New("circuit breaker open: falling back to local limit")
)

// CircuitState represents the current state of the Redis circuit breaker.
type CircuitState int32

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

// RedisClient defines the contract for Redis operations needed by the Leaser.
type RedisClient interface {
	EvalSHAOrScript(ctx context.Context, script string, keys []string, args ...interface{}) (int64, error)
	Ping(ctx context.Context) error
}

// LeaserConfig contains setup options for the Redis Tier-2 Leaser.
type LeaserConfig struct {
	MaxLatencyThreshold time.Duration // e.g. 50ms
	MaxErrorThreshold   int64         // Consec errors before opening circuit
	CooldownWindow      time.Duration // e.g. 5s before Half-Open probe
}

// RedisLeaser manages Tier-2 asynchronous batch token reservations from Redis with circuit breaking.
type RedisLeaser struct {
	client          RedisClient
	cfg             LeaserConfig
	state           int32 // CircuitState (atomic)
	consecErrors    int64
	lastStateChange int64 // Unix nanoseconds
}

// NewRedisLeaser instantiates a RedisLeaser.
func NewRedisLeaser(client RedisClient, cfg LeaserConfig) *RedisLeaser {
	if cfg.MaxLatencyThreshold <= 0 {
		cfg.MaxLatencyThreshold = 50 * time.Millisecond
	}
	if cfg.MaxErrorThreshold <= 0 {
		cfg.MaxErrorThreshold = 5
	}
	if cfg.CooldownWindow <= 0 {
		cfg.CooldownWindow = 5 * time.Second
	}

	return &RedisLeaser{
		client:          client,
		cfg:             cfg,
		state:           int32(StateClosed),
		lastStateChange: time.Now().UnixNano(),
	}
}

// RequestLease requests a batch lease of tokens from Redis for a key.
func (rl *RedisLeaser) RequestLease(ctx context.Context, key string, batchSize, capacity, refillRate int64) (int64, error) {
	if !rl.allowRequest() {
		return 0, ErrCircuitOpen
	}

	start := time.Now()
	nowSec := start.Unix()

	granted, err := rl.client.EvalSHAOrScript(
		ctx,
		BatchLeaseLuaScript,
		[]string{key},
		batchSize, capacity, refillRate, nowSec,
	)

	latency := time.Since(start)

	if err != nil || latency > rl.cfg.MaxLatencyThreshold {
		rl.recordFailure()
		if err != nil {
			return 0, err
		}
		// If latency exceeded threshold, return granted tokens but record warning
	} else {
		rl.recordSuccess()
	}

	return granted, nil
}

// State returns the current circuit breaker state.
func (rl *RedisLeaser) State() CircuitState {
	return CircuitState(atomic.LoadInt32(&rl.state))
}

func (rl *RedisLeaser) allowRequest() bool {
	st := rl.State()
	if st == StateClosed {
		return true
	}

	now := time.Now().UnixNano()
	lastChange := atomic.LoadInt64(&rl.lastStateChange)

	if st == StateOpen {
		if now-lastChange > rl.cfg.CooldownWindow.Nanoseconds() {
			if atomic.CompareAndSwapInt32(&rl.state, int32(StateOpen), int32(StateHalfOpen)) {
				atomic.StoreInt64(&rl.lastStateChange, now)
				return true // Allow single probe request
			}
		}
		return false
	}

	return true // Half-Open state allows request
}

func (rl *RedisLeaser) recordSuccess() {
	atomic.StoreInt64(&rl.consecErrors, 0)
	if rl.State() == StateHalfOpen {
		atomic.StoreInt32(&rl.state, int32(StateClosed))
		atomic.StoreInt64(&rl.lastStateChange, time.Now().UnixNano())
	}
}

func (rl *RedisLeaser) recordFailure() {
	errs := atomic.AddInt64(&rl.consecErrors, 1)
	if errs >= rl.cfg.MaxErrorThreshold || rl.State() == StateHalfOpen {
		if atomic.CompareAndSwapInt32(&rl.state, rl.state, int32(StateOpen)) {
			atomic.StoreInt64(&rl.lastStateChange, time.Now().UnixNano())
		}
	}
}
