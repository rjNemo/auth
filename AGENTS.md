# Repository Guidelines

## Project Structure & Module Organization

Keep the single Go module defined in `go.mod`. Server entry lives in `main.go`, while `index.html` and `in.html` provide public and authenticated templates beside it. Use `tmp/` strictly for generated artifacts (e.g., binaries, logs); never commit its contents. When expanding the app, group domain-specific handlers and helpers in packages under the repo root.

## Approved Technologies & Dependencies

Favor the Go standard library for routing, templating, crypto, and storage. The only pre-approved third-party package is `github.com/go-chi/chi/v5` for HTTP routing if the standard mux becomes limiting. Front-end interactivity must rely on htmx and Alpine.js with semantic HTML, and Pico.css is the accepted design system. Embed templates, scripts, and styles into the binary using Go's `embed` package so deployments stay self-contained. Avoid introducing other dependencies without prior discussion and an update to this document.

## Authentication Flow Requirements

Implement email/password authentication with secure password hashing, CSRF protection, and clear failure states. Use semantic forms enhanced by htmx for progressive enhancement and Alpine.js for lightweight client behavior. Persist session state on the server, prefer HTTP-only cookies, and render authenticated views without leaking sensitive data.

## Build, Test, and Development Commands

- `go run .` starts the dev server on <http://localhost:8000>.
- `go build -o tmp/auth` emits a clean binary for manual testing.
- After every change, run `gofmt -w ./...`, `go vet ./...`, and `go test ./...` to ensure formatting, static checks, and regression coverage before you push.

## Coding Style & Naming Conventions

Trust `gofmt`; no manual formatting tweaks. Use CamelCase for exported Go identifiers and snake_case for static assets. Keep handlers slim, factor reusable logic into helpers, and add brief comments only when intent is not obvious. Template IDs and Alpine component names should describe their role (e.g., `login_form`).

## Testing Guidelines

Adopt Go’s `testing` package with table-driven cases. Name files `<feature>_test.go`, colocated with the code under test. Run `go test ./...` (and `go test -cover ./...` for major features) before opening a PR, ensuring new branches maintain or raise coverage.

## Commit & Pull Request Guidelines

Follow Conventional Commits (`feat:`, `fix:`, `chore:`). PRs must summarize the change, link relevant issues, include manual verification notes, and attach screenshots or screen recordings for UI updates.

## Living Document

This guide evolves with the project. Whenever you add a tool, workflow, or architectural rule, update this section so future contributors inherit the collective knowledge.
