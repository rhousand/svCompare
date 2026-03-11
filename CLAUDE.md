# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture

Full-stack sailboat comparison web app. A single Go binary embeds the Vue SPA and serves both the API and the frontend.

| Layer | Tech |
|---|---|
| Backend | Go 1.23, Chi router, `modernc.org/sqlite` (pure Go, CGo-free) |
| Frontend | Vue 3 + Vite + Vue Router 4 + Pinia, no UI framework |
| Auth | JWT in HttpOnly cookies (SameSite=Lax). `Authenticator` interface designed for Google OAuth. |
| Database | SQLite. Timestamps stored as Unix integers. WAL mode enabled. |
| Dev | Nix flake + direnv. `air` for Go hot-reload. |
| Deploy | Docker (multi-stage). Single binary + volume-mounted SQLite file. |

## Development

Enter the dev shell first (requires Nix + direnv):

```bash
direnv allow          # or: nix develop
```

Then start both servers in separate terminals:

```bash
# Terminal 1 — Go API at :8080
cd backend && air

# Terminal 2 — Vue dev server at :5173 (proxies /api to :8080)
cd frontend && npm install && npm run dev
```

Open `http://localhost:5173`. Default credentials: `admin` / `admin`.

## Key Commands

```bash
# Go backend
cd backend
go build -tags dev .      # dev build (no frontend embed needed)
go build .                # prod build (requires frontend/dist to exist)
go test ./...

# Vue frontend
cd frontend
npm run dev               # dev server at :5173
npm run build             # build to frontend/dist

# Manual production build (no Nix/Docker)
cd frontend && npm run build
cp -r dist ../backend/frontend/dist
cd ../backend && go build -o svcompare .
./svcompare               # serves at :8080

# Docker (local test)
docker compose up --build

# Nix package build (update hashes first — see flake.nix comments)
nix build              # binary → ./result/bin/svcompare
nix build .#docker     # Docker image tarball → ./result
docker load < result
```

## Environment Variables

| Variable | Default | Notes |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `DATABASE_PATH` | `./data/svcompare.db` | SQLite file path (directory auto-created) |
| `JWT_SECRET` | `dev-secret-change-in-prod` | **Change in production** |
| `GO_ENV` | `development` | `development` enables CORS for Vite |
| `SEED_ADMIN_USERNAME` | `admin` | Seeded only when no users exist |
| `SEED_ADMIN_PASSWORD` | `admin` | Seeded only when no users exist |

## Key Files

| File | Purpose |
|---|---|
| `backend/internal/scoring/scoring.go` | **Single source of truth** for section weights and question IDs |
| `frontend/src/scoring.js` | Frontend copy — must stay in sync with the backend |
| `backend/internal/db/db.go` | All schema migrations, CRUD, and expiry cleanup |
| `backend/internal/auth/auth.go` | `Authenticator` interface — add Google OAuth here |
| `backend/embed_prod.go` | Embeds `frontend/dist` (default build tag) |
| `backend/embed_dev.go` | Stub used with `go build -tags dev` (no embed needed) |
| `backend/.air.toml` | Air config — builds with `-tags dev` |

## Scoring Model

25 questions across 7 sections from `sailboat_buyers_guide.md`:

| Section | Weight | Questions |
|---|---|---|
| Ownership & History | 20% | Q1–Q4 |
| Engine & Mechanical | 20% | Q5–Q8 |
| Sails & Rig | 15% | Q9–Q11 |
| Systems | 15% | Q12–Q15 |
| Survey & Hull Condition | 20% | Q16–Q19 |
| Electronics & Safety | 10% | Q20–Q22 |
| Transaction | 0% (informational) | Q23–Q25 |

Section score = average of scored questions (1–10, skipping unscored).
Weighted score = section average × weight.
Total = sum of weighted scores (Transaction excluded).
Comparisons auto-expire 30 days after the last score save.

## Adding Google OAuth

1. Implement the `Authenticator` interface in `backend/internal/auth/oauth.go`
2. Add `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` env vars
3. Register the new authenticator in `backend/main.go`
4. No other changes needed — all handlers use the interface

## Nix Hash Updates

After adding/changing dependencies, update the placeholder hashes in `flake.nix`:

```bash
# Go vendor hash
nix build 2>&1 | grep "got:"

# npm deps hash
nix build .#frontend 2>&1 | grep "got:"
```
