# AGENTS.md

This file provides guidance to AI coding assistants working with this repository.

## Build Commands

```bash
make dev          # Run with hot reload (generate + run)
make build        # Build binary to bin/server
make run          # Build and run binary
make generate     # Generate templ and sqlc code
make check        # Run go vet and build
make tidy         # go mod tidy
```

Code generation is required after modifying `.templ` or `.sql` files:
- `make generate-templ` - Regenerate from `internal/templates/*.templ`
- `make generate-sqlc` - Regenerate from `internal/db/queries.sql`

## Environment Variables

- `YOUTUBE_API_KEY` (required) - YouTube Data API v3 key
- `GOOGLE_CLIENT_SECRET` - Path to OAuth client secret JSON (default: `client_secret.json`)
- `DB_PATH` - SQLite database path (default: `data.db`)
- `PORT` - Server port (default: `8080`)

## Architecture

**Go web server** using standard `net/http` with HTMX for interactivity.

```
cmd/server/main.go     # Entry point, routing, schema init
internal/
  handlers/            # HTTP handlers (one file per resource)
  templates/           # templ components (generates *_templ.go)
  youtube/client.go    # YouTube Data API wrapper
  auth/oauth.go        # Google OAuth2 flow
  db/                  # sqlc-generated database layer
```

**Tech stack:**
- **templ** - Type-safe HTML templating (Go 1.25 tool directive)
- **sqlc** - SQL-first database access (SQLite via modernc.org/sqlite)
- **HTMX** - Client-side interactivity without JS frameworks
- **Tailwind CSS** - Styling via CDN

**Data flow:** Handlers receive requests → call youtube client or db queries → render templ components → HTMX updates DOM fragments.

**Database:** Schema defined inline in `cmd/server/main.go`. Queries in `internal/db/queries.sql` with sqlc annotations.
