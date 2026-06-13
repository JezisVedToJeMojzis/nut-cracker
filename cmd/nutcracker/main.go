package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nutcracker/internal/config"
	"nutcracker/internal/db"
	"nutcracker/internal/seed"
	"nutcracker/internal/server"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Root context cancelled on SIGINT/SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Apply database migrations before connecting the pool.
	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		return fmt.Errorf("migrating database: %w", err)
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	// Seed reference data (countries) on first run.
	if n, err := seed.CountriesIfEmpty(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "warning: seeding countries failed: %v\n", err)
	} else if n > 0 {
		fmt.Printf("nutcracker: seeded %d countries\n", n)
	}

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           server.New(pool, cfg).Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Run the server in a goroutine so we can wait for shutdown signals.
	serverErr := make(chan error, 1)
	go func() {
		fmt.Printf("nutcracker: listening on %s\n", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for either a server error or a shutdown signal.
	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		fmt.Println("\nnutcracker: shutting down...")
	}

	// Give in-flight requests up to 10s to complete.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	return nil
}
