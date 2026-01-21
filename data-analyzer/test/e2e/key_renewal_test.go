// +build e2e

package e2e

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"data-analyzer/internal/collector"
	"data-analyzer/internal/storage"
)

// mockKeyProvider simulates Discord key finder for E2E tests
type mockKeyProvider struct {
	mu         sync.Mutex
	keys       []string
	keyIndex   int
	waitTime   time.Duration
	waitCalled atomic.Int64
}

func newMockKeyProvider(keys []string, waitTime time.Duration) *mockKeyProvider {
	return &mockKeyProvider{
		keys:     keys,
		waitTime: waitTime,
	}
}

func (m *mockKeyProvider) WaitForKey(ctx context.Context) (string, error) {
	m.waitCalled.Add(1)

	// Simulate waiting for key
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(m.waitTime):
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.keyIndex >= len(m.keys) {
		return "", errors.New("no more keys available")
	}

	key := m.keys[m.keyIndex]
	m.keyIndex++
	return key, nil
}

func (m *mockKeyProvider) WaitCallCount() int64 {
	return m.waitCalled.Load()
}

// mockKeyValidator for E2E tests
type mockKeyValidator struct {
	mu         sync.Mutex
	validKeys  map[string]bool
	validateCt atomic.Int64
}

func newMockKeyValidator(validKeys []string) *mockKeyValidator {
	valid := make(map[string]bool)
	for _, k := range validKeys {
		valid[k] = true
	}
	return &mockKeyValidator{validKeys: valid}
}

func (m *mockKeyValidator) ValidateKey(key string) (bool, error) {
	m.validateCt.Add(1)
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.validKeys[key], nil
}

func (m *mockKeyValidator) ValidateCount() int64 {
	return m.validateCt.Load()
}

// expireAfterNSpider simulates a spider that returns 401 after N runs
type expireAfterNSpider struct {
	*mockSpiderRunner
	expireAfter int64
	runCount    atomic.Int64
}

func newExpireAfterNSpider(rotator *storage.FileRotator, matchesPerRun int, expireAfter int64) *expireAfterNSpider {
	return &expireAfterNSpider{
		mockSpiderRunner: newMockSpiderRunner(rotator, matchesPerRun),
		expireAfter:      expireAfter,
	}
}

func (s *expireAfterNSpider) RunContinuous(ctx context.Context) error {
	count := s.runCount.Add(1)

	// After N runs, return 401 error
	if count >= s.expireAfter {
		return collector.ErrAPIKeyExpired
	}

	return s.mockSpiderRunner.RunContinuous(ctx)
}

func (s *expireAfterNSpider) Reset() {
	s.mockSpiderRunner.Reset()
	s.runCount.Store(0)
}

// TestKeyRenewal_FullCycle tests the key expiration and renewal flow:
// - Start collector with mock API
// - After some runs, mock returns 401
// - Verify reduce triggers
// - Verify waiting for key state
// - Mock provides new key
// - Verify fresh restart (state cleared)
// - Verify collection resumes from Challenger #1
func TestKeyRenewal_FullCycle(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "e2e_key_renewal_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create rotator
	rotator, err := storage.NewFileRotator(tempDir)
	if err != nil {
		t.Fatalf("Failed to create rotator: %v", err)
	}
	defer rotator.Close()

	// Create spider that expires after 5 runs
	spider := newExpireAfterNSpider(rotator, 50, 5)

	// Create key provider that returns new key after delay
	keyProvider := newMockKeyProvider([]string{"RGAPI-new-test-key"}, 100*time.Millisecond)

	// Create key validator that accepts the new key
	keyValidator := newMockKeyValidator([]string{"RGAPI-new-test-key"})

	// Track states and notifications
	var notifications []string
	var notifyMu sync.Mutex
	var reduceCycles atomic.Int64

	warmDir := filepath.Join(tempDir, "warm")
	coldDir := filepath.Join(tempDir, "cold")

	reduceFunc := func(ctx context.Context) error {
		rotator.FlushAndRotate()

		agg, err := collector.AggregateWarmFiles(warmDir, func(int) bool { return true })
		if err != nil {
			return err
		}

		collector.ArchiveWarmToCold(warmDir, coldDir)
		reduceCycles.Add(1)
		t.Logf("Reduce cycle %d: %d records", reduceCycles.Load(), agg.TotalRecords)
		return nil
	}

	notifyFunc := func(ctx context.Context, message string) error {
		notifyMu.Lock()
		notifications = append(notifications, message)
		notifyMu.Unlock()
		t.Logf("Notification: %s", message)
		return nil
	}

	// Create configuration
	config := collector.ContinuousCollectorConfig{
		WarmFileThreshold:  10, // High threshold so we trigger from key expiry
		KeyPollInterval:    50 * time.Millisecond,
		ShutdownTimeout:    5 * time.Second,
		BloomResetInterval: 5,
	}

	// Create continuous collector
	cc := collector.NewContinuousCollector(
		spider,
		reduceFunc,
		keyValidator,
		keyProvider,
		notifyFunc,
		config,
	)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Track state transitions
	var sawWaitingForKey, sawFreshRestart, sawCollecting atomic.Bool
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				state := cc.State()
				switch state {
				case collector.StateWaitingForKey:
					sawWaitingForKey.Store(true)
				case collector.StateFreshRestart:
					sawFreshRestart.Store(true)
				case collector.StateCollecting:
					// Only count collecting after fresh restart
					if sawFreshRestart.Load() {
						sawCollecting.Store(true)
					}
				}
			}
		}
	}()

	// Run collector in background
	var runErr error
	done := make(chan struct{})
	go func() {
		runErr = cc.Run(ctx)
		close(done)
	}()

	// Wait for collector to go through the cycle
	// 1. Collect -> 2. Key expires -> 3. Reduce -> 4. Wait for key -> 5. Fresh restart -> 6. Collect again
	deadline := time.After(8 * time.Second)

	// Wait until we've seen the full cycle (collecting again after fresh restart)
	for !sawCollecting.Load() {
		select {
		case <-deadline:
			t.Logf("Current state: %v", cc.State())
			t.Logf("sawWaitingForKey: %v, sawFreshRestart: %v, sawCollecting: %v",
				sawWaitingForKey.Load(), sawFreshRestart.Load(), sawCollecting.Load())

			// Check what we got
			if !sawWaitingForKey.Load() {
				t.Error("Never reached WAITING_FOR_KEY state")
			}
			if !sawFreshRestart.Load() {
				t.Error("Never reached FRESH_RESTART state")
			}
			cancel()
			<-done
			return
		case <-time.After(50 * time.Millisecond):
			// Continue checking
		}
	}

	// Success! We went through the full cycle
	t.Log("Full key renewal cycle completed successfully")

	// Verify key provider was called
	if keyProvider.WaitCallCount() == 0 {
		t.Error("Key provider was never called")
	}
	t.Logf("Key provider wait called %d times", keyProvider.WaitCallCount())

	// Verify key validator was called
	if keyValidator.ValidateCount() == 0 {
		t.Error("Key validator was never called")
	}
	t.Logf("Key validated %d times", keyValidator.ValidateCount())

	// Verify notifications were sent
	notifyMu.Lock()
	if len(notifications) == 0 {
		t.Error("No notifications were sent")
	} else {
		t.Logf("Notifications sent: %v", notifications)
		// Check for expected notification content
		hasExpiredNotif := false
		hasStartedNotif := false
		for _, n := range notifications {
			if contains(n, "expired") {
				hasExpiredNotif = true
			}
			if contains(n, "started") {
				hasStartedNotif = true
			}
		}
		if !hasExpiredNotif {
			t.Error("Missing key expiration notification")
		}
		if !hasStartedNotif {
			t.Error("Missing session started notification")
		}
	}
	notifyMu.Unlock()

	// Verify spider was reset (fresh start)
	if !spider.WasReset() {
		t.Error("Spider was not reset on fresh restart")
	}

	// Shutdown
	cancel()
	<-done

	if runErr != nil && runErr != context.Canceled && runErr != context.DeadlineExceeded {
		t.Errorf("Unexpected error: %v", runErr)
	}
}

// TestKeyRenewal_InvalidKeyRetry tests that invalid keys are rejected
func TestKeyRenewal_InvalidKeyRetry(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "e2e_invalid_key_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	rotator, err := storage.NewFileRotator(tempDir)
	if err != nil {
		t.Fatalf("Failed to create rotator: %v", err)
	}
	defer rotator.Close()

	// Spider expires immediately
	spider := newExpireAfterNSpider(rotator, 10, 1)

	// Key provider provides invalid key first, then valid key
	keyProvider := newMockKeyProvider([]string{
		"RGAPI-invalid-key",
		"RGAPI-valid-key",
	}, 50*time.Millisecond)

	// Only accept the second key
	keyValidator := newMockKeyValidator([]string{"RGAPI-valid-key"})

	warmDir := filepath.Join(tempDir, "warm")
	coldDir := filepath.Join(tempDir, "cold")

	reduceFunc := func(ctx context.Context) error {
		rotator.FlushAndRotate()
		agg, _ := collector.AggregateWarmFiles(warmDir, func(int) bool { return true })
		collector.ArchiveWarmToCold(warmDir, coldDir)
		t.Logf("Reduced %d records", agg.TotalRecords)
		return nil
	}

	config := collector.ContinuousCollectorConfig{
		WarmFileThreshold:  10,
		KeyPollInterval:    50 * time.Millisecond,
		ShutdownTimeout:    5 * time.Second,
		BloomResetInterval: 5,
	}

	cc := collector.NewContinuousCollector(
		spider,
		reduceFunc,
		keyValidator,
		keyProvider,
		nil,
		config,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Track fresh restart
	var sawFreshRestart atomic.Bool
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if cc.State() == collector.StateFreshRestart {
					sawFreshRestart.Store(true)
					cancel() // We got what we wanted
					return
				}
			}
		}
	}()

	done := make(chan struct{})
	go func() {
		cc.Run(ctx)
		close(done)
	}()

	<-done

	// Verify validator was called at least twice (once for invalid, once for valid)
	validateCount := keyValidator.ValidateCount()
	if validateCount < 2 {
		t.Errorf("Expected at least 2 validation calls, got %d", validateCount)
	}
	t.Logf("Validation count: %d", validateCount)

	// Should have reached fresh restart with valid key
	if !sawFreshRestart.Load() {
		t.Error("Never reached FRESH_RESTART state (valid key not accepted)")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
