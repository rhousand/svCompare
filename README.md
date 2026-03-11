# svCompare

A multi-user web app for evaluating and comparing sailboats using a structured scoring framework. Score up to five boats across 25 questions covering ownership history, mechanical condition, sails and rigging, systems, hull condition, electronics, and safety — then compare them side-by-side with weighted scoring to find the best value.

---

## Features

- **Structured scoring** — 25 questions across 7 sections derived from a sailboat buyer's guide, each with in-app scoring guidance on hover
- **Weighted comparison** — section scores are averaged and weighted to produce a total out of 10.0 (Ownership & History 20%, Engine & Mechanical 20%, Survey & Hull Condition 20%, Sails & Rig 15%, Systems 15%, Electronics & Safety 10%)
- **Side-by-side view** — compare up to 5 boats in a single comparison
- **Shareable links** — every comparison has a public read-only URL you can send to a broker, partner, or surveyor
- **Auto-expiry** — comparisons automatically expire 30 days after the last score edit
- **PDF export** — one-click download of the full scoring grid and weighted summary
- **Multi-user** — each user manages their own comparisons independently
- **Google OAuth ready** — authentication is built on a clean interface; local username/password is the default, Google OAuth can be dropped in without changing any other code

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.23, [Chi](https://github.com/go-chi/chi) router |
| Database | SQLite via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGo) |
| Auth | JWT in HttpOnly cookies (SameSite=Lax) |
| Frontend | Vue 3, Vite, Vue Router 4, Pinia |
| Dev environment | Nix flake + direnv, `air` for Go hot-reload |
| Deployment | Single Go binary with embedded Vue SPA, Docker (scratch image) |

---

## Getting Started

### Prerequisites

- [Nix](https://nixos.org/download) with flakes enabled, and [direnv](https://direnv.net/) — **recommended**
- Or manually: Go 1.23+, Node.js 20+, npm

### Development (with Nix)

```bash
git clone https://github.com/your-username/svCompare.git
cd svCompare
direnv allow          # enters the Nix dev shell
```

Start both servers in separate terminals:

```bash
# Terminal 1 — Go API with hot-reload at :8080
cd backend && air

# Terminal 2 — Vue dev server at :5173
cd frontend && npm install && npm run dev
```

Open [http://localhost:5173](http://localhost:5173) and log in with:

```
Username: admin
Password: admin
```

### Development (without Nix)

```bash
# Backend
cd backend
go mod download
go build -tags dev -o tmp/main .
./tmp/main

# Frontend (separate terminal)
cd frontend
npm install
npm run dev
```

---

## Configuration

All configuration is via environment variables. The dev shell sets sensible defaults automatically.

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `DATABASE_PATH` | `./data/svcompare.db` | SQLite file path (directory is created automatically) |
| `JWT_SECRET` | `dev-secret-change-in-prod` | JWT signing secret — **change this in production** |
| `GO_ENV` | `development` | Set to `production` to disable CORS and serve the embedded SPA |
| `SEED_ADMIN_USERNAME` | `admin` | Admin username, seeded only if no users exist |
| `SEED_ADMIN_PASSWORD` | `admin` | Admin password — **change this in production** |

---

## Docker

### Local testing

```bash
docker compose up --build
```

The app is served at [http://localhost:8080](http://localhost:8080). SQLite data is persisted in a named Docker volume.

### Production build

```bash
# Build and run the container
docker build -t svcompare .
docker run -p 8080:8080 \
  -e JWT_SECRET=your-secret-here \
  -e SEED_ADMIN_PASSWORD=your-password-here \
  -v svcompare_data:/data \
  svcompare
```

The Docker image is built in three stages:
1. Node 20 Alpine — builds the Vue frontend
2. Go 1.23 Alpine — compiles the Go binary with the frontend embedded
3. `scratch` — minimal final image containing only the static binary (~20MB)

### Nix

```bash
nix build              # builds the binary → ./result/bin/svcompare
nix build .#docker     # builds a Docker image tarball → ./result
docker load < result
```

> **Note:** The `vendorHash` and `npmDepsHash` in `flake.nix` are placeholders. Run `nix build` once to get the correct hashes from the error output, then update `flake.nix`.

---

## Scoring Model

Scores are entered on a **1 (Bad) to 10 (Excellent)** scale. Questions left blank are excluded from section averages.

| Section | Weight | Questions |
|---|---|---|
| Ownership & History | 20% | Q1–Q4 |
| Engine & Mechanical | 20% | Q5–Q7 |
| Sails & Rig | 15% | Q9–Q11, Q15, Q26 |
| Systems | 15% | Q12–Q14 |
| Survey & Hull Condition | 20% | Q16–Q19 |
| Electronics & Safety | 10% | Q20–Q22 |
| Transaction | — (informational) | Q23–Q25 |

**Section score** = average of scored questions in that section
**Weighted score** = section average × section weight
**Total** = sum of all weighted scores (maximum 10.0, Transaction excluded)

The scoring model is defined in two places that must stay in sync:
- Backend: [`backend/internal/scoring/scoring.go`](backend/internal/scoring/scoring.go)
- Frontend: [`frontend/src/scoring.js`](frontend/src/scoring.js)

---

## Adding Google OAuth

The `Authenticator` interface in `backend/internal/auth/auth.go` is designed to make this a drop-in:

1. Create `backend/internal/auth/oauth.go` and implement the `Authenticator` interface
2. Add `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` environment variables
3. Register the new authenticator in `backend/main.go`

No changes to handlers, middleware, or the frontend are required.

---

## Project Structure

```
svCompare/
├── backend/
│   ├── main.go                        # Server entry point, routing, background cleanup
│   ├── embed_prod.go                  # Embeds frontend/dist (default build)
│   ├── embed_dev.go                   # Stub for -tags dev (no embed)
│   ├── .air.toml                      # Hot-reload config
│   └── internal/
│       ├── auth/auth.go               # Authenticator interface, JWT, bcrypt
│       ├── db/db.go                   # Schema migrations, all CRUD, expiry cleanup
│       ├── handlers/handlers.go       # All HTTP handlers
│       ├── middleware/auth.go         # JWT cookie middleware
│       ├── models/models.go           # Shared data structs
│       └── scoring/scoring.go        # Weighted score calculation
├── frontend/
│   ├── src/
│   │   ├── scoring.js                 # Question data and section weights
│   │   ├── stores/                    # Pinia stores (auth, comparisons)
│   │   ├── views/                     # LoginView, DashboardView, ComparisonView, ShareView
│   │   └── components/               # NavBar, WeightedTable, ShareLink, TooltipIcon
│   └── vite.config.js                # Dev proxy: /api → :8080
├── flake.nix                          # Nix dev shell, package build, Docker image
├── Dockerfile                         # Multi-stage production build
├── docker-compose.yml                 # Local Docker testing
└── sailboat_buyers_guide.md           # Source material for the scoring questions
```

---

## License

[MIT](LICENSE)
