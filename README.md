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

   | Target                   | Description                                                                                                                    |
   | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------ |
   | `make run`               | Start the HTTP server with the current environment.                                                                            |
   | `make dev`               | Launch [Air](https://github.com/cosmtrek/air) for live reload (requires `air` on PATH).                                        |
   | `make build`             | Compile to `./bin/auth-server`.                                                                                                |
   | `make test`              | Run `go test ./... -cover -count=1`.                                                                                           |
   | `make migrate-status`    | Show Goose migration status for the configured database.                                                                       |
   | `make migrate-up`        | Apply pending migrations to the database at `AUTH_DATABASE_URL` (defaults to `postgres://localhost/auth_dev?sslmode=disable`). |
   | `make migrate-down`      | Roll back the most recent migration in the target database.                                                                    |
   | `make migrate-reset`     | Reset the schema by rolling back all migrations, then re-applying them.                                                        |
   | `make migrate-new name=` | Create a timestamped SQL migration (e.g. `make migrate-new name=add_users`).                                                   |
   | `make sqlc-generate`     | Regenerate data-access code from SQL queries via `sqlc`.                                                                       |

3. Visit the login page (default <http://localhost:8000>) and authenticate with
   the demo credentials displayed on screen.

## Configuration

Settings are sourced from environment variables (see [.env](./.env)).

| Variable                    | Required    | Default       | Description                                                                          |
| --------------------------- | ----------- | ------------- | ------------------------------------------------------------------------------------ |
| `AUTH_SESSION_SECRET`       | Yes         | —             | Base64-encoded secret used to sign session cookies.                                  |
| `AUTH_DATABASE_URL`         | Yes         | —             | PostgreSQL connection string (e.g. `postgres://localhost/auth_dev?sslmode=disable`). |
| `AUTH_LISTEN_ADDR`          | No          | `:8000`       | Address the HTTP server binds to.                                                    |
| `AUTH_ENV`                  | No          | `development` | Environment label, controls logger source annotation.                                |
| `AUTH_LOG_MODE`             | No          | `text`        | Structured log encoder (`text` or `json`).                                           |
| `AUTH_GOOGLE_CLIENT_ID`     | Conditional | —             | Google OAuth 2.0 client ID; required when enabling Google social login.              |
| `AUTH_GOOGLE_CLIENT_SECRET` | Conditional | —             | Google OAuth 2.0 client secret matching the ID above.                                |
| `AUTH_GOOGLE_REDIRECT_URL`  | Conditional | —             | Registered redirect URL (e.g. `http://localhost:8000/login/google/callback`).        |

## Database Tooling

Migrations live in [`internal/driver/db/migrations`](./internal/driver/db/migrations)
and are managed with [Goose](https://github.com/pressly/goose).
Point `AUTH_DATABASE_URL` at your PostgreSQL instance—`postgres://localhost/auth_dev?sslmode=disable`
is a good local default—then use the Makefile helpers
(`make migrate-up`, `make migrate-status`, etc.) to evolve the schema.
The same DSN drives [`sqlc`](https://sqlc.dev/) generation with `make sqlc-generate`,
which reads [`internal/driver/db/sqlc.yaml`](./internal/driver/db/sqlc.yaml) and
emits typed data-access code alongside the queries.

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

## Deployment

Use Docker Compose to run the application and its PostgreSQL dependency on a VPS.
The database service is kept on the private Compose network (no host port published).

1. Provision secrets as environment variables
   (or in an env file referenced via `docker compose --env-file`):
   - `AUTH_SESSION_SECRET` must be a base64-encoded random value.
   - `POSTGRES_PASSWORD` and optional `POSTGRES_USER`/`POSTGRES_DB` override the
     database credentials referenced by `AUTH_DATABASE_URL`.
   - Google OAuth values are optional but required for social login.
2. Build images with `make compose-build` (or `docker compose build`).
3. Start the stack in the background: `docker compose up -d`.
4. Monitor logs with `docker compose logs -f app`.

To run administrative commands, exec into the containers
(e.g. `docker compose exec db psql`).

## License

MIT
