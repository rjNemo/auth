# Authentication Implementation Plan

## Project Restructure

- Move the entrypoint to `cmd/server/main.go` and create `internal/server` for router, middleware, and session bootstrap logic.
- Add `internal/auth` for credential hashing, validation helpers, and user/session abstractions.
- Relocate templates and static assets to `web/` and embed them with `embed.FS` so the binary stays self-contained.

## Routing & Middleware

- Adopt `github.com/go-chi/chi/v5` with recovery, request logging, and session-loading middleware.
- Maintain an in-memory session store backed by HTTP-only cookies; compare secrets with `crypto/subtle` and rotate session IDs on login/logout.

## Authentication Core

- Implement password hashing with per-user random salts (`crypto/rand`, `sha256`) and base64 encoding for storage.
- Define a user repository interface seeded with an in-memory implementation until persistence is added.
- Generate CSRF tokens tied to session state and validate them for every mutating request.
- Externalize session secrets via configuration so environments use predictable/rotatable keys.
- Add integration tests covering login/logout flows and CSRF protections to prevent regressions.

## Templates & Frontend

- Replace static HTML with `html/template` views built on semantic markup and Pico.css styling.
- Enhance forms using htmx for progressive submission and Alpine.js for light client interactivity (loading states, error visibility).
- Serve embedded assets via helper endpoints or template functions using the embedded filesystem.

## Handlers & Flows

- Create `GET /signin`, `POST /signin`, `GET /signup`, `POST /signup`, `POST /signout`, and protected dashboard routes.
- Return precise HTTP status codes and htmx-compatible fragments for validation errors.
- Guard authenticated routes with middleware that checks session validity before rendering protected pages.

## Testing Strategy

- Cover hashing and CSRF helpers with table-driven unit tests.
- Use `net/http/httptest` to verify happy-path login, signup, logout, invalid credential handling, CSRF failures, and session persistence.
- Run `go test -cover ./...` to ensure the new logic maintains regression coverage.
- Flesh out dedicated service tests for lookup flows and extend dashboard coverage once integration scaffolding is available.
- Add structured logging: text encoder for development, JSON for production deployments.
- Consolidate templates with a base layout to remove duplication across pages.
- Introduce configuration loading that sources environment variables, validates them, and exposes typed settings at startup.
