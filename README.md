# Family Finance — Multi-tenant Household Finance Tracking

A family finance tracking app for managing household budgets, accounts, transactions, and financial goals.

## Project Structure

- `backend/` — Go/Fiber/GORM REST API
- `frontend/` — Next.js React SPA
- `specs.md` — Full product specification (Indonesian)
- `db.md` — Database design reference
- `fe.md` — Frontend page map
- `.phases/` — 13-phase implementation guide

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 14+

### Backend Setup

```bash
cd backend

# Copy and edit environment variables
cp .env.example .env
# Edit .env: fill JWT_SECRET + DB credentials

# Install Go dependencies
go mod tidy

# Run migrations
make migrate-up

# Start API server
make run
# Server runs on http://localhost:8080
```

### Frontend Setup

```bash
cd frontend

# Copy environment variables
cp .env.example .env

# Install dependencies
npm install

# Start dev server
npm run dev
# App runs on http://localhost:3000
```

## Available Commands

**Backend (from `backend/`):**

```bash
make run              # Start API server
make build            # Build binary
make migrate-up       # Apply all pending migrations
make migrate-down     # Rollback last migration
make migrate-status   # Show migration status
go build ./...        # Compile check
go vet ./...          # Lint check
```

**Frontend (from `frontend/`):**

```bash
npm run dev           # Start development server
npm run build         # Build for production
npm run start         # Run production build
npm run lint          # Run ESLint
```

## Documentation

- **Phase Guide:** See `.phases/00-overview.md` for how phases are organized and conventions
- **Current Phase:** Phase 01 (Setup & Fondasi) — foundation and infrastructure ready
- **API Spec:** Generated Swagger docs at `backend/docs/swagger.yaml`

## Architecture

### Backend

Strict Clean Architecture with layers:
- `delivery/http` — Fiber handlers and middleware
- `usecase` — Business logic
- `repository` — Data access interfaces and implementations
- `entity` — GORM models

### Frontend

Next.js App Router with:
- TypeScript
- Tailwind CSS + daisyUI
- API helper in `lib/api.ts` for all requests

## Multi-tenancy

Every domain table has `household_id`. **Always resolve `household_id` server-side from JWT** — never trust client-supplied household IDs. See `CLAUDE.md` section "Multi-tenancy — the one rule that matters most".

## Key Files

- `CLAUDE.md` — Repository conventions, architecture, commands
- `db.md` — Entity schemas, relationships, migration notes
- `fe.md` — Frontend routes and API consumption per page
- `.phases/*.md` — Per-phase definition of done and API contracts
