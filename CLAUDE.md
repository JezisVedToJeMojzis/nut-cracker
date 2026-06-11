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
```

## Architecture

```
cmd/nutcracker/   ← main entrypoint
internal/         ← private packages (business logic, db, handlers)
```

- `cmd/nutcracker/main.go` calls `run()` and handles top-level error reporting to stderr with exit code 1.
- Business logic goes in `internal/` packages, not in `main.go`.
- HTTP handlers, DB layer, and domain logic should each live in separate `internal/` sub-packages.
