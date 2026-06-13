// Package seed populates reference data (the countries table) from the
// authoritative ISO 3166-1 dataset, downloaded at runtime.
package seed

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const datasetURL = "https://raw.githubusercontent.com/lukes/ISO-3166-Countries-with-Regional-Codes/master/all/all.csv"

// CountriesIfEmpty seeds the countries table only when it is empty. Returns the
// number of countries inserted (0 if it was already populated).
func CountriesIfEmpty(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM countries`).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting countries: %w", err)
	}
	if count > 0 {
		return 0, nil
	}
	return Countries(ctx, pool)
}

// Countries upserts the full ISO 3166-1 country list. Idempotent.
func Countries(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	records, err := fetchCountries(ctx)
	if err != nil {
		return 0, fmt.Errorf("fetching countries: %w", err)
	}
	var n int
	for _, c := range records {
		if _, err := pool.Exec(ctx,
			`INSERT INTO countries (code, name) VALUES ($1, $2)
			 ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name`,
			c.code, c.name); err != nil {
			return 0, fmt.Errorf("upserting %s: %w", c.code, err)
		}
		n++
	}
	return n, nil
}

type country struct {
	code string
	name string
}

func fetchCountries(ctx context.Context) ([]country, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, datasetURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
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
		return nil, fmt.Errorf("expected columns not found")
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
