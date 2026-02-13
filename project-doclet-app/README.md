# Doclet

![Doclet Logo](./docs/doclet-logo.png)

Anonymous real-time collaboration on rich-text documents

## Architecture overview

System components:

- Frontend (React + TipTap + Yjs): editor UI, presence, and WebSocket client.
- Document service (Go): REST API for basic document metadata CRUD.
- Collaboration service (Go): WebSocket hub for realtime updates and presence.
- NATS: broadcasts updates across service replicas.
- PostgreSQL: stores document metadata and snapshots.

![Component Diagram](./docs/component-diagram.png)

## Local setup

### 1) Start dependencies
```sh
docker compose up postgres nats
```

### 2) Set environment variables
```sh
cp .env.example .env
```
Edit `.env` as needed. Defaults assume local Postgres + NATS.

### 3) Run the Go services
```sh
export $(grep -v '^#' .env | xargs)
go run ./cmd/document
```
In another terminal:
```sh
export $(grep -v '^#' .env | xargs)
go run ./cmd/collab
```

### 4a) Run the frontend with hot reloading
```sh
cd webapp-react-frontend
npm install
npm run dev
```

Open `http://localhost:5173`.

### 4b) Frontend config
Runtime config is read from `webapp-react-frontend/public/config.json`. Use `webapp-react-frontend/config.example.json` as a template and copy to `webapp-react-frontend/config.json` for local overrides.

---
## Alternative: Run full stack with Docker

This runs Postgres, NATS, the backend Go services, and the frontend with docker-compose. This is not set up to
hot reload.

```sh
docker compose --profile app up -d
```
Access the frontend at `http://localhost:5173`

## Useful commands
- `go test ./...`
- `docker compose ps`
- `docker compose down -v`

## Notes
- Document schema is code-first (Gorm). Migrations are generated via `go run ./services/document/cmd/atlas`.
- Snapshots are saved via NATS on a short debounce from the editor.
