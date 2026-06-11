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

## Documentation

Project documentation lives in `docs/` and `README.md` — read these for context:
- **Database schema:** `docs/schema.md`
- **Setup & commands:** `README.md`

## Hard Rules

- **Migrations:** Every migration must have a working `.down.sql` rollback. Use `IF NOT EXISTS` on all CREATE statements and `IF EXISTS` on all DROP statements so up/down are idempotent.
- **Map privacy:** A user's map is visible only to themselves and their accepted friends. Enforce this authorization on every map-read endpoint (return 403 otherwise).
- **Auth:** Google OAuth required, plus email/password. Built last, but schema already supports it (`users.password_hash` nullable + `user_identities`).

## Architecture

```
cmd/nutcracker/   ← main entrypoint
internal/         ← private packages (business logic, db, handlers)
```

- `cmd/nutcracker/main.go` calls `run()` and handles top-level error reporting to stderr with exit code 1.
- Business logic goes in `internal/` packages, not in `main.go`.
- HTTP handlers, DB layer, and domain logic should each live in separate `internal/` sub-packages.
