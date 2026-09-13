package sync

import (
	"context"
	"errors"
	"testing"
	"time"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

type mockRedis struct {
	evalFunc func(ctx context.Context, script string, keys []string, args ...interface{}) (int64, error)
}

func (m *mockRedis) EvalSHAOrScript(ctx context.Context, script string, keys []string, args ...interface{}) (int64, error) {
	if m.evalFunc != nil {
		return m.evalFunc(ctx, script, keys, args...)
	}
	return 50, nil
}

func (m *mockRedis) Ping(ctx context.Context) error {
	return nil
}

func TestRedisLeaser_NormalOperation(t *testing.T) {
	mock := &mockRedis{
		evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) (int64, error) {
			return 50, nil
		},
	}

	leaser := NewRedisLeaser(mock, LeaserConfig{
		MaxLatencyThreshold: 50 * time.Millisecond,
		MaxErrorThreshold:   3,
	})

	granted, err := leaser.RequestLease(context.Background(), "user_123", 50, 100, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if granted != 50 {
		t.Errorf("expected 50 granted tokens, got %d", granted)
	}

	if leaser.State() != StateClosed {
		t.Errorf("expected state CLOSED, got %s", leaser.State())
	}
}

func TestRedisLeaser_CircuitTripAndRecovery(t *testing.T) {
	failCount := 0
	mock := &mockRedis{
		evalFunc: func(ctx context.Context, script string, keys []string, args ...interface{}) (int64, error) {
			failCount++
			if failCount <= 3 {
				return 0, errors.New("redis connection refused")
			}
			return 50, nil
		},
	}

	leaser := NewRedisLeaser(mock, LeaserConfig{
		MaxErrorThreshold: 3,
		CooldownWindow:    50 * time.Millisecond,
	})

	// Trigger 3 failures
	for i := 0; i < 3; i++ {
		_, _ = leaser.RequestLease(context.Background(), "user_123", 50, 100, 10)
	}

	if leaser.State() != StateOpen {
		t.Errorf("expected state OPEN after 3 failures, got %s", leaser.State())
	}

	// Immediate next call should fail-fast with ErrCircuitOpen
	_, err := leaser.RequestLease(context.Background(), "user_123", 50, 100, 10)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}

	// Wait for cooldown window to expire
	time.Sleep(60 * time.Millisecond)

	// Probe call in Half-Open state (4th request returns 50, nil)
	granted, err := leaser.RequestLease(context.Background(), "user_123", 50, 100, 10)
	if err != nil {
		t.Fatalf("expected probe request to succeed, got %v", err)
	}

	if granted != 50 {
		t.Errorf("expected 50 granted tokens in half-open probe, got %d", granted)
	}

	if leaser.State() != StateClosed {
		t.Errorf("expected state CLOSED after successful probe, got %s", leaser.State())
	}
}
