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

### Frontend

```bash
cd web
npm install
npm run dev      # http://localhost:5173
```

The frontend (SvelteKit) proxies `/api/*` to the Go backend on `:8080`. Sign up
or log in (cookie session). Left-click a country to crack it, right-click to
remove a crack; toggle **Count mode** to track how many people per country.

## Deployment

The app deploys as a **single Docker service**: the Go server serves the built
frontend and the API on one origin (`/api`), runs DB migrations and seeds
countries on startup.

- **Database:** create a Postgres instance (e.g. [Neon](https://neon.tech)) and
  copy its connection string (use `sslmode=require`).
- **Host (Render):** the repo includes a `Dockerfile` and `render.yaml`. Create a
  Blueprint from the repo, then set env vars in the dashboard:
  - `DATABASE_URL` — the Neon connection string
  - `APP_BASE_URL` — the public Render URL (e.g. `https://nutcracker.onrender.com`)
  - `COOKIE_SECURE=true`
  - `RESEND_API_KEY` — optional (empty = emails logged to server logs)

Render injects `PORT` automatically. HTTPS is provided, which also makes the PWA
installable on phones.

Build the image locally:

```bash
docker build -t nutcracker .
docker run -p 8080:8080 --env-file .env nutcracker
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
