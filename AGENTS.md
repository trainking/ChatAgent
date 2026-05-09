# AGENTS.md

## Architecture

Three independent components, no monorepo tooling:

| Dir | Stack | Role |
|-----|-------|------|
| `server/` | Go (Gin, sqlx, Viper, Zap) | REST API, JWT auth, RBAC |
| `web/` | Vue 3 + TS + Vite + Element Plus | Agent dashboard SPA |
| `widget/` | Pre-built Vue 3 artifact | Embeddable chat widget (no source in repo — only `dist/`) |

Message flow: `web/` or `widget/` → HTTP/WS → `server/` → PostgreSQL + Redis.
Kafka is in dev docker-compose but **not implemented** in server code yet.

## Commands

Start dev infrastructure first, then run server and web:

```bash
# Start DB + Redis + Kafka
docker compose up -d

# Backend (port 8080)
cd server && make run          # or: go run ./cmd/server

# Frontend (port 3000, proxies /api -> server:8080)
cd web && npm install && npm run dev
```

**Important**: The Vite dev server proxies `/api/*` → `localhost:8080` and `/ws` → `ws://localhost:8080`. Always run both server and web for full functionality.

| Task | Server | Web |
|------|--------|-----|
| Dev | `make run` | `npm run dev` |
| Build | `make build` | `npm run build` (includes `vue-tsc --noEmit` first) |
| Test | `make test` (`go test ./...`) | *(no test script yet)* |
| Lint | `make lint` (`golangci-lint run`) | `npm run lint` (`eslint --fix`) |
| Format | — | `npm run format` (prettier) |

## First-run initialization

After starting services, open `http://localhost:3000`. The app detects no root user exists and redirects to `/init` where you create the super_admin account. This can only be done once.

## Setup prerequisites

1. Copy `.env.example` to `.env` and adjust values (`.env` is gitignored)
2. Viper loads `server/config/config.yaml` first, then overrides with env vars. Env vars use `_` separator: `DB_HOST` maps to config key `db.host`.
3. PostgreSQL and Redis must be running (`docker compose up -d` is sufficient)

## API conventions (critical)

**All responses return HTTP 200**, even errors. The response body is:

```json
{ "code": 0, "message": "success", "data": {} }
```

- `code: 0` = success, any non-zero code = error.
- The frontend `request.ts` interceptor checks `code !== 0` and shows `ElMessage.error`.
- HTTP 401 from the server triggers automatic logout and redirect to `/login`.
- Use `pkg/response` helpers (`Success`, `Error`) and `pkg/errcode` constants when adding handlers.

## Role hierarchy

`super_admin` > `admin` > `agent`

- `RequireAdmin()` middleware protects user management
- `RequireSuperAdmin()` protects role permission and system config endpoints

## Database (important)

- **No migration files.** All schema creation is inline SQL in `server/internal/database/migrate.go`. The `migrations/` directory is empty. When adding tables, add `CREATE TABLE IF NOT EXISTS` to `RunMigrations()`.
- No migration versioning/tracking — tables are created idempotently on startup.

## Code conventions

**Web (Vue/TS)**:
- `semi: false`, `singleQuote: true`, `trailingComma: all`, `printWidth: 100`
- tsconfig strict mode, bundler module resolution, `@/` alias maps to `src/`
- Run `npm run lint` before committing; `npm run build` validates types

**Server (Go)**:
- Standard `internal/` layout: `handler/` → `service/` → `repository/` → `sqlx`
- Error codes defined in `pkg/errcode/errcode.go`
- Use `config.Config` (not env vars directly) in services

## Widget

No source code is committed — only `widget/dist/widget.js` and `widget/dist/widget-demo.html`. The widget connects to `ws://localhost:8080/ws`, but the WebSocket handler (`server/internal/websocket/`) is a **planned feature, not yet implemented**. The directory exists but is empty.

## Docker compose differences

| | Root `docker-compose.yml` (dev) | `deploy/docker-compose.yml` (prod) |
|---|---|---|
| Kafka | Included | Not included |
| DB/Redis ports | Exposed (5432, 6379) | Not exposed |
| Volume mounts | Yes (live reload) | No |
| Server port | 8080:8080 | 8081:8080 |
| Web port | 80:80 | 3000:80 |
| Server env | defaults | `GIN_MODE=release`, `LOG_FORMAT=json` |

## Testing

No automated tests exist in server or web yet. `make test` runs `go test ./...` which passes with zero tests. Manual integration test scripts live in `ansible/files/` (curl-based). Test account seed script: `ansible/files/seed-accounts.sh`.

## No CI/CD

No GitHub Actions, no pre-commit hooks, no task runner. Run lint+typecheck manually before pushing.
