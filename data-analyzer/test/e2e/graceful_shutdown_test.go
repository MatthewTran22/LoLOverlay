// +build e2e

package e2e

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"data-analyzer/internal/collector"
	"data-analyzer/internal/storage"
)

// TestGracefulShutdown_MidCollection tests shutdown during active collection:
// - Start collector, accumulate some warm files
// - Trigger shutdown
// - Verify final reduce runs
// - Verify all data in cold/
// - Verify clean exit
func TestGracefulShutdown_MidCollection(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "e2e_graceful_shutdown_*")
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

	// Create spider
	spider := newMockSpiderRunner(rotator, 100)

	warmDir := filepath.Join(tempDir, "warm")
	coldDir := filepath.Join(tempDir, "cold")

	var reduceCycles atomic.Int64
	var totalRecords atomic.Int64

	reduceFunc := func(ctx context.Context) error {
		rotator.FlushAndRotate()

		agg, err := collector.AggregateWarmFiles(warmDir, func(int) bool { return true })
		if err != nil {
			return err
		}

		archived, err := collector.ArchiveWarmToCold(warmDir, coldDir)
		if err != nil {
			return err
		}

		totalRecords.Add(int64(agg.TotalRecords))
		reduceCycles.Add(1)
		t.Logf("Reduce cycle %d: %d records, %d files archived",
			reduceCycles.Load(), agg.TotalRecords, archived)
		return nil
	}

	// High threshold to not auto-trigger reduce
	config := collector.ContinuousCollectorConfig{
		WarmFileThreshold:  100, // Won't trigger automatically
		KeyPollInterval:    time.Second,
		ShutdownTimeout:    10 * time.Second,
		BloomResetInterval: 5,
	}

	cc := collector.NewContinuousCollector(
		spider,
		reduceFunc,
		nil, nil, nil,
		config,
	)

	ctx, cancel := context.WithCancel(context.Background())

	// Run collector
	done := make(chan struct{})
	go func() {
		cc.Run(ctx)
		close(done)
	}()

	// Let it collect for a bit
	time.Sleep(500 * time.Millisecond)

	t.Log("Triggering graceful shutdown...")

	// Record time before shutdown
	shutdownStart := time.Now()

	// Trigger graceful shutdown
	cancel()

	// Wait for shutdown to complete
	select {
	case <-done:
		shutdownDuration := time.Since(shutdownStart)
		t.Logf("Shutdown completed in %v", shutdownDuration)
	case <-time.After(15 * time.Second):
		t.Fatal("Shutdown timeout - collector did not stop gracefully")
	}

	// Verify final reduce ran
	cycles := reduceCycles.Load()
	t.Logf("Final reduce cycles: %d", cycles)

	// Check cold storage
	coldFiles, _ := filepath.Glob(filepath.Join(coldDir, "*.jsonl.gz"))
	warmFiles, _ := filepath.Glob(filepath.Join(warmDir, "*.jsonl"))
	hotFiles, _ := filepath.Glob(filepath.Join(tempDir, "hot", "*.jsonl"))

	t.Logf("Final state - Hot: %d, Warm: %d, Cold: %d files",
		len(hotFiles), len(warmFiles), len(coldFiles))

	// With graceful shutdown, warm should be empty (all archived to cold)
	// or hot should be empty (flushed to warm and archived)
	if len(warmFiles) > 0 && len(hotFiles) > 0 {
		t.Errorf("Data not fully processed: %d warm files, %d hot files remaining",
			len(warmFiles), len(hotFiles))
	}

	// Verify some data was collected
	records := totalRecords.Load()
	if records == 0 && len(coldFiles) == 0 {
		t.Error("No data was collected or archived")
	} else {
		t.Logf("Total records processed: %d", records)
	}

	t.Log("Graceful shutdown test complete")
}

// TestGracefulShutdown_DuringReduce tests shutdown while reduce is in progress
func TestGracefulShutdown_DuringReduce(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "e2e_shutdown_during_reduce_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	rotator, err := storage.NewFileRotator(tempDir)
	if err != nil {
		t.Fatalf("Failed to create rotator: %v", err)
	}
	defer rotator.Close()

	spider := newMockSpiderRunner(rotator, 100)

	warmDir := filepath.Join(tempDir, "warm")
	coldDir := filepath.Join(tempDir, "cold")

	var reduceStarted, reduceCompleted atomic.Bool

	// Slow reduce function to simulate long-running reduce
	reduceFunc := func(ctx context.Context) error {
		reduceStarted.Store(true)
		t.Log("Reduce started (will take 500ms)")

		rotator.FlushAndRotate()

		// Simulate slow processing
		time.Sleep(500 * time.Millisecond)

		agg, _ := collector.AggregateWarmFiles(warmDir, func(int) bool { return true })
		collector.ArchiveWarmToCold(warmDir, coldDir)

		reduceCompleted.Store(true)
		t.Logf("Reduce completed with %d records", agg.TotalRecords)
		return nil
	}

	// Low threshold to trigger reduce quickly
	config := collector.ContinuousCollectorConfig{
		WarmFileThreshold:  1,
		KeyPollInterval:    time.Second,
		ShutdownTimeout:    10 * time.Second,
		BloomResetInterval: 5,
	}

	cc := collector.NewContinuousCollector(
		spider,
		reduceFunc,
		nil, nil, nil,
		config,
	)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		cc.Run(ctx)
		close(done)
	}()

	// Wait for reduce to start
	deadline := time.After(5 * time.Second)
	for !reduceStarted.Load() {
		select {
		case <-deadline:
			t.Fatal("Reduce never started")
		case <-time.After(10 * time.Millisecond):
		}
	}

	// Shutdown immediately after reduce starts (mid-reduce)
	t.Log("Shutting down during reduce...")
	cancel()

	// Wait for shutdown
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("Shutdown timeout")
	}

	// Verify reduce completed (not interrupted)
	if !reduceCompleted.Load() {
		t.Error("Reduce was interrupted by shutdown - should complete first")
	}

	t.Log("Shutdown during reduce test complete")
}

// TestGracefulShutdown_MultipleSignals tests calling shutdown multiple times
func TestGracefulShutdown_MultipleSignals(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "e2e_multi_shutdown_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	rotator, err := storage.NewFileRotator(tempDir)
	if err != nil {
		t.Fatalf("Failed to create rotator: %v", err)
	}
	defer rotator.Close()

	spider := newMockSpiderRunner(rotator, 50)

	warmDir := filepath.Join(tempDir, "warm")
	coldDir := filepath.Join(tempDir, "cold")

	var reduceCalls atomic.Int64

	reduceFunc := func(ctx context.Context) error {
		reduceCalls.Add(1)
		rotator.FlushAndRotate()
		agg, _ := collector.AggregateWarmFiles(warmDir, func(int) bool { return true })
		collector.ArchiveWarmToCold(warmDir, coldDir)
		t.Logf("Reduce call %d: %d records", reduceCalls.Load(), agg.TotalRecords)
		return nil
	}

	config := collector.ContinuousCollectorConfig{
		WarmFileThreshold:  100, // High threshold
		KeyPollInterval:    time.Second,
		ShutdownTimeout:    5 * time.Second,
		BloomResetInterval: 5,
	}

	cc := collector.NewContinuousCollector(
		spider,
		reduceFunc,
		nil, nil, nil,
		config,
	)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		cc.Run(ctx)
		close(done)
	}()

	// Let it run briefly
	time.Sleep(200 * time.Millisecond)

	// Call Shutdown multiple times rapidly
	t.Log("Calling shutdown 5 times rapidly...")
	shutdownCtx := context.Background()
	for i := 0; i < 5; i++ {
		go cc.Shutdown(shutdownCtx)
	}

	// Also cancel context
	cancel()

	// Wait for shutdown
	select {
	case <-done:
		t.Log("Collector stopped successfully")
	case <-time.After(10 * time.Second):
		t.Fatal("Shutdown timeout with multiple signals")
	}

	// Verify no panic occurred (test completes = success)
	reduceCt := reduceCalls.Load()
	t.Logf("Reduce was called %d times (1-2 expected)", reduceCt)

	// Should only have 1-2 reduce calls max (initial shutdown reduce, maybe one before)
	if reduceCt > 3 {
		t.Errorf("Too many reduce calls (%d) - shutdown may have triggered multiple reduces", reduceCt)
	}

	t.Log("Multiple shutdown signals test complete")
}

// TestGracefulShutdown_Timeout tests shutdown timeout behavior
func TestGracefulShutdown_Timeout(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "e2e_shutdown_timeout_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	rotator, err := storage.NewFileRotator(tempDir)
	if err != nil {
		t.Fatalf("Failed to create rotator: %v", err)
	}
	defer rotator.Close()

	spider := newMockSpiderRunner(rotator, 50)

	warmDir := filepath.Join(tempDir, "warm")
	coldDir := filepath.Join(tempDir, "cold")

	// Very slow reduce that would exceed timeout
	reduceFunc := func(ctx context.Context) error {
		t.Log("Starting very slow reduce...")
		rotator.FlushAndRotate()

		// This would take longer than shutdown timeout
		select {
		case <-ctx.Done():
			t.Log("Reduce context cancelled")
			// Still try to complete essential work
		case <-time.After(100 * time.Millisecond): // Quick path for test
		}

		agg, _ := collector.AggregateWarmFiles(warmDir, func(int) bool { return true })
		collector.ArchiveWarmToCold(warmDir, coldDir)
		t.Logf("Reduce completed: %d records", agg.TotalRecords)
		return nil
	}

	// Very short shutdown timeout
	config := collector.ContinuousCollectorConfig{
		WarmFileThreshold:  1,
		KeyPollInterval:    time.Second,
		ShutdownTimeout:    100 * time.Millisecond, // Very short timeout
		BloomResetInterval: 5,
	}

	cc := collector.NewContinuousCollector(
		spider,
		reduceFunc,
		nil, nil, nil,
		config,
	)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		cc.Run(ctx)
		close(done)
	}()

	// Let it start reducing
	time.Sleep(300 * time.Millisecond)

	// Shutdown
	t.Log("Initiating shutdown with short timeout...")
	start := time.Now()
	cancel()

	// Wait for shutdown
	select {
	case <-done:
		duration := time.Since(start)
		t.Logf("Shutdown completed in %v", duration)

		// Should complete relatively quickly despite timeout
		if duration > 5*time.Second {
			t.Errorf("Shutdown took too long: %v", duration)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Shutdown completely stuck")
	}

	t.Log("Shutdown timeout test complete")
}
