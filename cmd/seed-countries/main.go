// Command seed-countries populates the countries table from the ISO 3166-1
// dataset. Safe to run repeatedly (upsert).
package main

import (
	"context"
	"fmt"
	"os"

	"nutcracker/internal/config"
	"nutcracker/internal/db"
	"nutcracker/internal/seed"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	n, err := seed.Countries(ctx, pool)
	if err != nil {
		return err
	}
	fmt.Printf("seeded %d countries\n", n)
	return nil
}
