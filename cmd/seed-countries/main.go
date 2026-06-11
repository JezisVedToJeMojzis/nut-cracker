// Command seed-countries populates the countries table from the authoritative
// ISO 3166-1 dataset, downloaded at runtime. Safe to run repeatedly (upsert).
package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"nutcracker/internal/config"
	"nutcracker/internal/db"
)

const datasetURL = "https://raw.githubusercontent.com/lukes/ISO-3166-Countries-with-Regional-Codes/master/all/all.csv"

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

	records, err := fetchCountries()
	if err != nil {
		return fmt.Errorf("fetching countries: %w", err)
	}

	var count int
	for _, c := range records {
		_, err := pool.Exec(ctx,
			`INSERT INTO countries (code, name) VALUES ($1, $2)
			 ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name`,
			c.code, c.name)
		if err != nil {
			return fmt.Errorf("upserting %s: %w", c.code, err)
		}
		count++
	}

	fmt.Printf("seeded %d countries\n", count)
	return nil
}

type country struct {
	code string
	name string
}

func fetchCountries() ([]country, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(datasetURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	r := csv.NewReader(resp.Body)

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}
	nameIdx, codeIdx := columnIndex(header, "name"), columnIndex(header, "alpha-2")
	if nameIdx < 0 || codeIdx < 0 {
		return nil, fmt.Errorf("expected columns not found in dataset")
	}

	var out []country
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, country{code: row[codeIdx], name: row[nameIdx]})
	}
	return out, nil
}

func columnIndex(header []string, name string) int {
	for i, h := range header {
		if h == name {
			return i
		}
	}
	return -1
}
