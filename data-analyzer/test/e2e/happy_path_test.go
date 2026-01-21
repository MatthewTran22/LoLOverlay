// +build e2e

package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"data-analyzer/internal/collector"
	"data-analyzer/internal/storage"
)

// mockSpiderRunner simulates match collection for E2E tests
type mockSpiderRunner struct {
	mu              sync.Mutex
	matchesPerRun   int
	runsCompleted   atomic.Int64
	seeded          bool
	reset           bool
	onRunContinuous func() error
	rotator         *storage.FileRotator
	onRotation      func() // Called when a file rotation occurs
}

func newMockSpiderRunner(rotator *storage.FileRotator, matchesPerRun int) *mockSpiderRunner {
	return &mockSpiderRunner{
		matchesPerRun: matchesPerRun,
		rotator:       rotator,
	}
}

func (m *mockSpiderRunner) RunContinuous(ctx context.Context) error {
	m.runsCompleted.Add(1)

	if m.onRunContinuous != nil {
		return m.onRunContinuous()
	}

	// Simulate writing matches to rotator
	for i := 0; i < m.matchesPerRun; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if m.rotator != nil {
				// Write 10 participants per match (standard LoL match)
				for p := 0; p < 10; p++ {
					match := storage.RawMatch{
						MatchID:      fmt.Sprintf("NA1_%s_%d_%d", time.Now().Format("150405"), i, p),
						GameVersion:  "15.24.123",
						ChampionID:   1 + p,
						ChampionName: "Champion" + fmt.Sprintf("%d", p),
						TeamPosition: []string{"TOP", "JUNGLE", "MIDDLE", "BOTTOM", "UTILITY"}[p%5],
						Win:          p < 5, // First 5 win, last 5 lose
						Item0:        3089,
						Item1:        3157,
					}
					if err := m.rotator.WriteLine(match); err != nil {
						return err
					}
				}
				// Signal match complete - this triggers rotation check
				matchesBefore, _ := m.rotator.Stats()
				if err := m.rotator.MatchComplete(); err != nil {
					return err
				}
				matchesAfter, _ := m.rotator.Stats()

				// Check if rotation occurred (counter reset to 0)
				if matchesAfter < matchesBefore && m.onRotation != nil {
					m.onRotation()
				}
			}
		}
	}

	// Small delay to simulate API calls
	time.Sleep(10 * time.Millisecond)
	return nil
}

func (m *mockSpiderRunner) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reset = true
	m.runsCompleted.Store(0)
}

func (m *mockSpiderRunner) SeedFromChallenger(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seeded = true
	return nil
}

func (m *mockSpiderRunner) WasSeeded() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.seeded
}

func (m *mockSpiderRunner) WasReset() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.reset
}

// mockDataPusher captures push calls for verification
type mockDataPusher struct {
	mu        sync.Mutex
	pushes    []*collector.AggData
	pushDelay time.Duration
	pushErr   error
}

func (m *mockDataPusher) PushAggData(ctx context.Context, data *collector.AggData) error {
	if m.pushDelay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(m.pushDelay):
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pushErr != nil {
		return m.pushErr
	}

	m.pushes = append(m.pushes, data)
	return nil
}

func (m *mockDataPusher) PushCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.pushes)
}

// TestHappyPath_FullCollectionCycle tests the complete happy path:
// - Start continuous collector
// - Collect matches until 10 warm files
// - Verify reduce triggers
// - Verify collector resumes
// - Run for 3 reduce cycles
func TestHappyPath_FullCollectionCycle(t *testing.T) {
	// Create temp directory for storage
	tempDir, err := os.MkdirTemp("", "e2e_happy_path_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create file rotator
	rotator, err := storage.NewFileRotator(tempDir)
	if err != nil {
		t.Fatalf("Failed to create rotator: %v", err)
	}
	defer rotator.Close()

	// Track reduce cycles
	var reduceCycles atomic.Int64
	var lastAggData *collector.AggData
	var aggMu sync.Mutex

	warmDir := filepath.Join(tempDir, "warm")
	coldDir := filepath.Join(tempDir, "cold")

	// Create mock spider that writes to rotator
	spider := newMockSpiderRunner(rotator, 100) // 100 matches per run

	// We'll set the rotation callback after creating the collector
	var cc *collector.ContinuousCollector

	// Create reduce function that tracks calls
	reduceFunc := func(ctx context.Context) error {
		// Flush hot file
		if _, err := rotator.FlushAndRotate(); err != nil {
			t.Logf("FlushAndRotate error (expected on first run): %v", err)
		}

		// Aggregate
		agg, err := collector.AggregateWarmFiles(warmDir, func(itemID int) bool {
			return itemID >= 1000
		})
		if err != nil {
			return err
		}

		aggMu.Lock()
		lastAggData = agg
		aggMu.Unlock()

		// Archive
		_, err = collector.ArchiveWarmToCold(warmDir, coldDir)
		if err != nil {
			return err
		}

		reduceCycles.Add(1)
		t.Logf("Reduce cycle %d complete: %d records", reduceCycles.Load(), agg.TotalRecords)
		return nil
	}

	// Create configuration with low threshold for faster testing
	config := collector.ContinuousCollectorConfig{
		WarmFileThreshold:  2, // Trigger reduce every 2 warm files (2000 matches)
		KeyPollInterval:    time.Second,
		ShutdownTimeout:    10 * time.Second,
		BloomResetInterval: 5,
	}

	// Create continuous collector
	cc = collector.NewContinuousCollector(
		spider,
		reduceFunc,
		nil, // no key validator needed for happy path
		nil, // no key provider needed
		nil, // no notify function needed
		config,
	)

	// Wire up rotation callback to increment warm file counter
	spider.onRotation = func() {
		cc.IncrementWarmFileCount()
		t.Logf("File rotated, warm file count incremented")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run collector in background
	var runErr error
	done := make(chan struct{})
	go func() {
		runErr = cc.Run(ctx)
		close(done)
	}()

	// Wait for at least 3 reduce cycles
	targetCycles := int64(3)
	deadline := time.After(25 * time.Second)

	for reduceCycles.Load() < targetCycles {
		select {
		case <-deadline:
			t.Errorf("Timeout waiting for %d reduce cycles, only got %d", targetCycles, reduceCycles.Load())
			cancel()
			<-done
			return
		case <-time.After(100 * time.Millisecond):
			// Continue checking
		}
	}

	// Verify we hit the target
	cycles := reduceCycles.Load()
	if cycles < targetCycles {
		t.Errorf("Expected at least %d reduce cycles, got %d", targetCycles, cycles)
	}
	t.Logf("Completed %d reduce cycles", cycles)

	// Verify spider was seeded
	if !spider.WasSeeded() {
		t.Error("Spider was not seeded from Challenger")
	}

	// Verify data was aggregated
	aggMu.Lock()
	if lastAggData == nil {
		t.Error("No aggregated data produced")
	} else {
		if lastAggData.TotalRecords == 0 {
			t.Error("Aggregated data has no records")
		}
		t.Logf("Last aggregation: %d records, %d champion stats",
			lastAggData.TotalRecords, len(lastAggData.ChampionStats))
	}
	aggMu.Unlock()

	// Verify cold storage has archived files
	coldFiles, _ := filepath.Glob(filepath.Join(coldDir, "*.jsonl.gz"))
	if len(coldFiles) == 0 {
		t.Error("No archived files in cold storage")
	}
	t.Logf("Cold storage has %d archived files", len(coldFiles))

	// Trigger graceful shutdown
	cancel()
	<-done

	if runErr != nil && runErr != context.Canceled && runErr != context.DeadlineExceeded {
		t.Errorf("Unexpected error from Run: %v", runErr)
	}

	t.Log("Happy path test complete")
}

// TestHappyPath_WithTursoPush tests the full path including async Turso push
func TestHappyPath_WithTursoPush(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "e2e_turso_push_*")
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

	// Create mock pusher
	mockPusher := &mockDataPusher{pushDelay: 50 * time.Millisecond}
	tursoPusher := collector.NewTursoPusher(mockPusher)

	// Start pusher
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tursoPusher.Start(ctx)

	warmDir := filepath.Join(tempDir, "warm")
	coldDir := filepath.Join(tempDir, "cold")

	// Track reduce completion times
	var reduceCycles atomic.Int64

	// Create reduce function that queues Turso push
	reduceFunc := func(reduceCtx context.Context) error {
		// Flush hot
		rotator.FlushAndRotate()

		// Aggregate
		agg, err := collector.AggregateWarmFiles(warmDir, func(int) bool { return true })
		if err != nil {
			return err
		}

		// Archive
		collector.ArchiveWarmToCold(warmDir, coldDir)

		// Queue push (async)
		if agg.TotalRecords > 0 {
			if err := tursoPusher.Push(reduceCtx, agg); err != nil {
				t.Logf("Push queue error: %v", err)
			}
		}

		reduceCycles.Add(1)
		return nil
	}

	// Create spider
	spider := newMockSpiderRunner(rotator, 100)

	// Create collector with low threshold
	config := collector.ContinuousCollectorConfig{
		WarmFileThreshold: 2,
		KeyPollInterval:   time.Second,
		ShutdownTimeout:   10 * time.Second,
	}

	cc := collector.NewContinuousCollector(
		spider,
		reduceFunc,
		nil, nil, nil,
		config,
	)

	// Wire up rotation callback
	spider.onRotation = func() {
		cc.IncrementWarmFileCount()
	}

	// Run collector
	done := make(chan struct{})
	go func() {
		cc.Run(ctx)
		close(done)
	}()

	// Wait for a few reduce cycles
	time.Sleep(5 * time.Second)

	// Shutdown
	cancel()
	<-done

	// Wait for pusher to complete
	tursoPusher.Wait()

	// Verify pushes happened
	pushCount := mockPusher.PushCount()
	cycles := reduceCycles.Load()

	t.Logf("Completed %d reduce cycles, %d pushes", cycles, pushCount)

	if cycles == 0 {
		t.Error("No reduce cycles completed")
	}

	// Pushes should match reduce cycles (each reduce queues a push if data exists)
	if pushCount == 0 && cycles > 0 {
		t.Error("No Turso pushes completed despite reduce cycles")
	}

	t.Log("Turso push integration test complete")
}
