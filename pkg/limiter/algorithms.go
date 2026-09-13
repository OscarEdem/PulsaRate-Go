package limiter

import (
	"sync"
	"sync/atomic"
	"time"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

// LeakyBucket enforces constant-rate processing by leaking requests at fixed intervals.
type LeakyBucket struct {
	capacity   int64
	leakRate   time.Duration // Time interval per token leak
	waterLevel int64         // Current queued requests (sync/atomic)
	lastLeak   int64         // Unix nanoseconds (sync/atomic)
}

// NewLeakyBucket creates a new LeakyBucket algorithm instance.
func NewLeakyBucket(capacity int64, requestsPerSec int64) (*LeakyBucket, error) {
	if capacity <= 0 || requestsPerSec <= 0 {
		return nil, ErrInvalidConfig
	}

	leakInterval := time.Second / time.Duration(requestsPerSec)
	now := time.Now().UnixNano()

	return &LeakyBucket{
		capacity:   capacity,
		leakRate:   leakInterval,
		waterLevel: 0,
		lastLeak:   now,
	}, nil
}

// Allow evaluates if a request can enter the leaky bucket queue.
func (b *LeakyBucket) Allow() bool {
	now := time.Now().UnixNano()
	last := atomic.LoadInt64(&b.lastLeak)
	elapsed := time.Duration(now - last)

	if elapsed >= b.leakRate {
		leaked := int64(elapsed / b.leakRate)
		if leaked > 0 {
			if atomic.CompareAndSwapInt64(&b.lastLeak, last, now) {
				for {
					curr := atomic.LoadInt64(&b.waterLevel)
					next := curr - leaked
					if next < 0 {
						next = 0
					}
					if atomic.CompareAndSwapInt64(&b.waterLevel, curr, next) {
						break
					}
				}
			}
		}
	}

	for {
		curr := atomic.LoadInt64(&b.waterLevel)
		if curr >= b.capacity {
			return false
		}
		if atomic.CompareAndSwapInt64(&b.waterLevel, curr, curr+1) {
			return true
		}
	}
}

// SlidingWindowLog tracks precise request timestamps within a sliding time window.
type SlidingWindowLog struct {
	mu       sync.Mutex
	capacity int64
	window   time.Duration
	logs     []int64
}

// NewSlidingWindowLog instantiates a SlidingWindowLog rate limiter.
func NewSlidingWindowLog(capacity int64, window time.Duration) (*SlidingWindowLog, error) {
	if capacity <= 0 || window <= 0 {
		return nil, ErrInvalidConfig
	}

	return &SlidingWindowLog{
		capacity: capacity,
		window:   window,
		logs:     make([]int64, 0, capacity),
	}, nil
}

// Allow evaluates if a request falls within the allowed sliding window limit.
func (s *SlidingWindowLog) Allow() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixNano()
	boundary := now - s.window.Nanoseconds()

	// Evict outdated timestamps outside current window
	validIndex := 0
	for i, ts := range s.logs {
		if ts > boundary {
			validIndex = i
			break
		}
		if i == len(s.logs)-1 {
			validIndex = len(s.logs)
		}
	}
	s.logs = s.logs[validIndex:]

	if int64(len(s.logs)) >= s.capacity {
		return false
	}

	s.logs = append(s.logs, now)
	return true
}
