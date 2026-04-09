# vlab-api

Backend API for the Virtual Computer Lab — an interactive platform to learn computer hardware, components, and architecture.

Built with **Go**, **PostgreSQL (Neon)**, and **Docker**.

---

## Tech stack

- Go 1.25
- Chi router
- PostgreSQL via Neon (serverless)
- JWT authentication
- bcrypt password hashing
- Docker multi-stage build

---

## Project structure
vlab-api/
├── cmd/api/          → server entry point
├── internal/
│   ├── db/           → database queries and models
│   ├── handler/      → HTTP handlers
│   └── middleware/   → JWT auth and CORS
├── migrations/       → SQL migration files
├── sqlc/             → SQLC config and query files
├── .env.example      → environment variables template
├── Dockerfile
└── Makefile

---

## Getting started

### 1. Clone the repo

```bash
git clone https://github.com/joss12/vlab-api.git
cd vlab-api
```

### 2. Set up environment variables

```bash
cp .env.example .env
```

Edit `.env` and fill in your values:

### 3. Run migrations

```bash
goose -dir migrations postgres "$DATABASE_URL" up
```

### 4. Start the server

```bash
go run cmd/api/main.go
```

---

## API endpoints

### Auth

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/auth/register` | Register a new user | No |
| POST | `/auth/login` | Login and get token | No |

### Quiz

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/quiz/score` | Save a quiz score | Yes |
| GET | `/quiz/scores` | Get all scores for user | Yes |

### Progress

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/progress/{component}` | Update component progress | Yes |
| GET | `/progress` | Get all progress for user | Yes |

---

## Request examples

### Register
```bash
curl -X POST http://localhost:8089/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Login
```bash
curl -X POST http://localhost:8089/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Save quiz score
```bash
curl -X POST http://localhost:8089/quiz/score \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"score":13,"total":15,"component":"cpu"}'
```

### Update progress
```bash
curl -X POST http://localhost:8089/progress/cpu \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"tabs_visited":["overview","history"],"completed":false}'
```

---

## Docker

### Build
```bash
docker build -t vlab-api .
```
DATABASE_URL=postgres://USER:PASSWORD@HOST/DBNAME?sslmode=require
JWT_SECRET=your-random-secret
PORT=8089

### Run
```bash
docker run -p 8089:8089 \
  -e DATABASE_URL=your_neon_url \
  -e JWT_SECRET=your_secret \
  -e PORT=8089 \
  vlab-api
```

---

## Health check

```bash
curl http://localhost:8089/health
```

Expected response:
```json
{"status":"ok","service":"vlab-api"}
```

---

