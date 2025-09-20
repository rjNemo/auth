# Auth Demo

Auth Demo showcases a fully server-rendered email/password authentication flow
with secure session management, CSRF protection, structured logging, and embedded
templates/assets for single-binary deployment.

## Capabilities

- Email/password signup and login backed by salted hashing and reusable auth services.
- CSRF-protected session middleware with signed cookies and automatic token rotation.
- Structured logging (text or JSON) and environment-driven configuration for
  production parity.
- Embedded templates styled with Pico.css and progressively enhanced with htmx
  and Alpine.js.

## Getting Started

1. Review or adjust the defaults in [.env](./.env). To load them in POSIX shells,
   run `set -a; . ./.env; set +a`.
2. Use the targets in the [Makefile](./Makefile):

   | Target       | Description                                                                             |
   | ------------ | --------------------------------------------------------------------------------------- |
   | `make run`   | Start the HTTP server with the current environment.                                     |
   | `make dev`   | Launch [Air](https://github.com/cosmtrek/air) for live reload (requires `air` on PATH). |
   | `make build` | Compile to `./bin/auth-server`.                                                         |
   | `make test`  | Run `go test ./... -cover -count=1`.                                                    |

3. Visit the login page (default <http://localhost:8000>) and authenticate with
   the demo credentials displayed on screen.

## Configuration

Settings are sourced from environment variables (see [.env](./.env)).

| Variable                    | Required    | Default       | Description                                                                   |
| --------------------------- | ----------- | ------------- | ----------------------------------------------------------------------------- |
| `AUTH_SESSION_SECRET`       | Yes         | —             | Base64-encoded secret used to sign session cookies.                           |
| `AUTH_LISTEN_ADDR`          | No          | `:8000`       | Address the HTTP server binds to.                                             |
| `AUTH_ENV`                  | No          | `development` | Environment label, controls logger source annotation.                         |
| `AUTH_LOG_MODE`             | No          | `text`        | Structured log encoder (`text` or `json`).                                    |
| `AUTH_GOOGLE_CLIENT_ID`     | Conditional | —             | Google OAuth 2.0 client ID; required when enabling Google social login.       |
| `AUTH_GOOGLE_CLIENT_SECRET` | Conditional | —             | Google OAuth 2.0 client secret matching the ID above.                         |
| `AUTH_GOOGLE_REDIRECT_URL`  | Conditional | —             | Registered redirect URL (e.g. `http://localhost:8000/login/google/callback`). |

## Project Layout

- `cmd/server` — application entrypoint.
- `internal/config` — environment-backed configuration loader.
- `internal/driver/logging` — `slog` helpers for text/JSON output.
- `internal/service/auth` — authentication domain logic, hashing, validation.
- `internal/server` — router, middleware, handlers, session store.
- `web/templates` — embedded HTML templates.

## Built With

- [Go](https://go.dev/doc/) — standard library HTTP, templates, crypto, and `embed`.
- [Chi](https://github.com/go-chi/chi) — lightweight router and middleware stack.
- [htmx](https://htmx.org/) — progressive enhancement via HTML attributes.
- [Alpine.js](https://alpinejs.dev/) — declarative client-side interactions.
- [Pico.css](https://picocss.com/) — minimal, semantic-first styling.

## License

MIT
