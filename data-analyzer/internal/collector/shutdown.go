package collector

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler creates a context that is cancelled on SIGTERM or SIGINT.
// It also calls the provided shutdown function before cancelling.
func SetupSignalHandler(shutdownFunc func(context.Context)) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		log.Printf("[Signal] Received %v, initiating graceful shutdown...", sig)

		// Call shutdown function if provided
		if shutdownFunc != nil {
			shutdownFunc(ctx)
		}

		// Cancel context
		cancel()

		// Handle second signal - force exit
		sig = <-sigCh
		log.Printf("[Signal] Received second %v, forcing exit", sig)
		os.Exit(1)
	}()

	return ctx
}

// WaitForSignal blocks until a signal is received, then calls shutdown and returns
func WaitForSignal(ctx context.Context, shutdownFunc func(context.Context)) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-ctx.Done():
		log.Println("[Signal] Context cancelled")
	case sig := <-sigCh:
		log.Printf("[Signal] Received %v, initiating graceful shutdown...", sig)
		if shutdownFunc != nil {
			shutdownFunc(ctx)
		}
	}
}
