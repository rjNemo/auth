# Repository Guidelines

## Project Structure & Module Organization

Keep the single Go module defined in `go.mod`. The executable entrypoint lives in `cmd/server/main.go`; reusable HTTP and auth helpers belong under `internal/` (e.g., `internal/server`, `internal/auth`). Store templates and static assets beneath `web/` and embed them into the binary so deployments stay self-contained. Reserve `tmp/` for generated artifacts (binaries, logs) and keep it out of version control.

## Approved Technologies & Dependencies

Favor the Go standard library for routing, templating, crypto, and storage. The only pre-approved third-party package is `github.com/go-chi/chi/v5` when the default mux falls short. Pico.css provides styling, while htmx and Alpine.js deliver client interactivity; stick to semantic HTML. `golangci-lint` is the endorsed aggregator when linting beyond `go vet`. Introduce anything else only after discussion and an update here.

## Authentication Flow Requirements

Implement email/password authentication with secure password hashing, CSRF protection, and clear failure states. Use semantic forms enhanced by htmx for progressive enhancement and Alpine.js for lightweight client behavior. Persist session state on the server with HTTP-only cookies and render authenticated views without leaking sensitive data.

## Build, Lint, and Test Commands

- `go run ./cmd/server` starts the dev server on <http://localhost:8000>.
- `go build ./...` (and `go build -o tmp/auth ./cmd/server`) validates compilation before any formatting or linting step.
- After a successful build, run `gofmt -w ./...`, `go vet ./...`, `golangci-lint run` (if configured), and `go test ./...` to keep style, static checks, and regressions in check.

## Coding Style & Naming Conventions

Trust `gofmt`; avoid manual formatting. Use CamelCase for exported Go identifiers and snake_case for embedded assets. Keep handlers slim, factor shared logic into helpers, and add concise comments only when intent needs clarification. Promote named constants/variables over magic numbers or strings. Template IDs and Alpine component names should reflect their role (e.g., `login_form`). Name handlers that render full pages with a `PageHandler` suffix and reserve the plain `Handler` suffix for non-page actions.

## Testing Guidelines

Adopt Go’s `testing` package with table-driven cases. Name files `<feature>_test.go`, colocated with the code under test. Run `go test ./... -cover -count=1` before opening a PR so coverage is measured on fresh binaries; we don’t enforce a target, but avoid notable drops when adding code.

## Commit & Pull Request Guidelines

Follow Conventional Commits (`feat:`, `fix:`, `chore:`). PRs must summarize the change, link relevant issues, include manual verification notes (commands executed, browsers checked), and attach screenshots or recordings for UI updates.

## Living Document

This guide evolves with the project. Whenever you add a tool, workflow, or architectural rule, update this section so future contributors inherit the collective knowledge.
