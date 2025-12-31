# Firehose 🔥

A load testing UI for Postgres connection poolers with real-time latency and throughput metrics.

![Firehose Dashboard](docs/screenshot.png)

## Features

- **Adjustable Load** — Control connections, read QPS, and write QPS with live sliders
- **Real-time Metrics** — Latency (P50/P99), throughput, and error rates streamed via WebSocket
- **High Throughput** — Go backend with goroutines can push tens of thousands of QPS
- **Clean UI** — React dashboard with live-updating charts

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 14+ (or your connection pooler pointing to Postgres)

### 1. Set Up the Database

```bash
psql -h localhost -U postgres -d pooler_demo -f init.sql
```

### 2. Start the Backend

```bash
cd backend
cp .env.example .env  # Edit with your database URL
go run .
```

### 3. Start the Frontend

```bash
cd frontend
npm install
npm run dev
```

Open [http://localhost:5173](http://localhost:5173) and start blasting.

## Configuration

Environment variables for the backend:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://localhost:5432/pooler_demo` | Postgres connection string (point this at your pooler) |
| `HTTP_PORT` | `8080` | HTTP/WebSocket server port |
| `MAX_CONNECTIONS` | `500` | Maximum allowed connections |
| `MAX_READ_QPS` | `50000` | Maximum read queries per second |
| `MAX_WRITE_QPS` | `10000` | Maximum write queries per second |

## Architecture

```
┌────────────────┐      HTTP/WS       ┌────────────────┐
│                │◄──────────────────►│                │
│  React + Vite  │                    │   Go Backend   │
│                │                    │                │
└────────────────┘                    └───────┬────────┘
                                              │
                                              ▼
                                     ┌────────────────┐
                                     │  Your Pooler   │
                                     └───────┬────────┘
                                              │
                                              ▼
                                     ┌────────────────┐
                                     │   PostgreSQL   │
                                     └────────────────┘
```

The Go backend spawns worker goroutines that execute queries at the configured rate. Metrics are collected, aggregated every 100ms, and streamed to the frontend via WebSocket.

See [DESIGN.md](DESIGN.md) for the full technical design.

## Docker Compose

```bash
docker-compose up
```

This starts Postgres, your pooler, the backend, and frontend. Access the dashboard at [http://localhost:3000](http://localhost:3000).

## Development

```bash
# Run backend with hot reload
cd backend
go install github.com/air-verse/air@latest
air

# Run frontend
cd frontend
npm run dev
```

## Workload Details

**Reads** — Random point selects by primary key:
```sql
SELECT id, username, email, created_at FROM users WHERE id = $1
```

**Writes** — Inserts with generated data:
```sql
INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id
```
