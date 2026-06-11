# Nut Cracker

A PWA where users create accounts, view a world map, and color countries to mark
"I've cracked nuts with someone from this country." Users can add friends and view
each other's maps (friends-only).

## Tech Stack

- **Backend:** Go
- **Frontend:** Svelte (PWA) *(planned)*
- **Database:** PostgreSQL (via Docker)

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.26+
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI:
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```

### Setup

```bash
# 1. Copy environment template and adjust if needed
cp .env.example .env

# 2. Start PostgreSQL
docker compose up -d

# 3. Apply database migrations
migrate -path internal/db/migrations -database "$DATABASE_URL" up

# 4. Seed the countries table (downloads the ISO 3166-1 list, idempotent)
go run ./cmd/seed-countries

# 5. Run the server
go run ./cmd/nutcracker
```

The server listens on `:8080` by default. Health check:

```bash
curl http://localhost:8080/health
```

## Common Commands

```bash
go build ./...                 # build
go test ./...                  # run tests
docker compose up -d           # start database
docker compose down            # stop database
```

## Documentation

- [API reference](docs/api.md)
- [Database schema](docs/schema.md)

## Project Structure

```
cmd/nutcracker/        ← main entrypoint
internal/
  config/              ← env/config loading
  db/                  ← connection pool + migrations
  server/              ← HTTP routing and handlers
docs/                  ← project documentation
```
