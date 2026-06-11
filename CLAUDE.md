# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Nut Cracker** — A PWA where users create accounts, view a world map, and color countries to mark "I've cracked nuts with someone from this country." Users can view each other's maps and add friends.

**Planned stack:**
- **Backend:** Go (this repo)
- **Frontend:** Svelte (PWA)
- **Database:** PostgreSQL
- **Hosting:** Free tier (TBD)

## Session Context

At the start of each session, read all previous messages in the conversation thread to restore context. Key decisions made so far:
- Project is a CLI entrypoint for now; web server will be added under `cmd/nutcracker/`
- Module name: `nutcracker`
- Git remote: https://github.com/JezisVedToJeMojzis/nut-cracker
- Developer: Samuel Mojzis (mojzissamuel@gmail.com)

## Conventions

- **Commit messages:** Do NOT add "Co-Authored-By: Claude" or any Claude/AI attribution to commits.

## Commands

```bash
# Build
go build ./...

# Run
go run ./cmd/nutcracker

# Test
go test ./...

# Single package test
go test ./internal/somepackage/...

# Start / stop the PostgreSQL container
docker compose up -d
docker compose down

# Database migrations (golang-migrate CLI; binary at $(go env GOPATH)/bin/migrate)
migrate -path internal/db/migrations -database "$DATABASE_URL" up
migrate -path internal/db/migrations -database "$DATABASE_URL" down 1
# Create a new migration pair:
migrate create -ext sql -dir internal/db/migrations -seq <name>
```

## Database Schema

Five tables (migration `000001_init`):
- `users` — accounts (`password_hash` nullable for OAuth-only users)
- `user_identities` — external logins (Google, etc.); PK `(provider, provider_user_id)`
- `countries` — ISO 3166-1 alpha-2 reference list (`code`, `name`)
- `user_countries` — colored countries; `cracks` int (>=1) = people cracked with from that country; PK `(user_id, country_code)`
- `friendships` — request/accept model (`status`: pending|accepted)

**Hard rule:** a user's map is visible only to themselves and their accepted friends. Enforce this authorization on every map-read endpoint (return 403 otherwise).

**Auth:** Google OAuth required, plus email/password. Build auth last, but schema already supports it.

## Architecture

```
cmd/nutcracker/   ← main entrypoint
internal/         ← private packages (business logic, db, handlers)
```

- `cmd/nutcracker/main.go` calls `run()` and handles top-level error reporting to stderr with exit code 1.
- Business logic goes in `internal/` packages, not in `main.go`.
- HTTP handlers, DB layer, and domain logic should each live in separate `internal/` sub-packages.
