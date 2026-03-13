# Perjadin Protokol Kemnaker RI

This repository contains the official **Perjadin Protokol** system for **Kementerian Ketenagakerjaan Republik Indonesia**. It is an enterprise-grade web application designed to manage official travel records, expense reporting, and protocol activities.

## 🧱 Architecture & Tech Stack

The project follows a **Monorepo** structure separating the frontend and backend, orchestrated via containerization.

| Component | Technology | Description |
| :--- | :--- | :--- |
| **Frontend** | [SvelteKit](https://kit.svelte.dev/) | SSR/CSR hybrid web application using Node.js adapter. |
| **Backend** | [Go (Golang)](https://go.dev/) | RESTful API using [Echo](https://echo.labstack.com/) framework. |
| **Database** | [PostgreSQL](https://www.postgresql.org/) | Primary relational database. |
| **Cache** | [Redis](https://redis.io/) | Session storage and caching layer. |
| **PDF Engine** | [Gotenberg](https://gotenberg.dev/) | Microservice for converting documents to PDF. |
| **Secrets** | [Doppler](https://www.doppler.com/) | Centralized secrets management (Zero-trust). |
| **Infra** | [Podman](https://podman.io/) / Docker | Container runtime and orchestration. |

---

## 🛠️ Prerequisites

To contribute to this project, you must have the following tools installed on your local machine:

1.  **[Git](https://git-scm.com/)** - Version control.
2.  **[Go](https://go.dev/dl/)** - Version 1.24+.
3.  **[Node.js](https://nodejs.org/)** - LTS Version (v20+).
4.  **[Podman Desktop](https://podman-desktop.io/)** (or Docker Desktop) - Container runtime.
5.  **[Doppler CLI](https://docs.doppler.com/docs/install-cli)** - Required for injecting environment variables.
6.  **SQL Tools** (Install via Go):
    ```bash
    # Type-safe SQL generator
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
    # Database migration tool
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    ```

---

## 🚀 Quick Start (Fast-Track Onboarding)

Follow these steps to get your development environment running in **under 5 minutes**.

### 1. Setup Project & Secrets
```bash
# Clone the repository
git clone <repository-url>
cd perjadin-protokol-kemnaker-ri/DEVELOPMENT/perjadin-kemnaker-ri

# Login to Doppler and select the 'dev' config
doppler login
doppler setup
# Select Project: perjadin-kemnaker-ri
# Select Config: dev (or dev_local)
```

### 2. Start Infrastructure (Redis Only)
Since PostgreSQL runs on **Baremetal (Host Machine)**, we only need to start the Redis container for caching/session management.

```bash
# Start Redis in the background
podman-compose up -d redis
# Ensure your local PostgreSQL service is running on port 5432
```

### 3. Initialize Database
Apply the database migrations to your local Postgres instance.
```bash
cd backend
# Run migrations (assuming default local credentials)
migrate -path db/migrations -database "postgresql://postgres:postgres@localhost:5432/perjadin?sslmode=disable" up
```

### 4. Run Backend (Native Mode)
Open a new terminal. Run the Go backend locally with secrets injected.
```bash
cd backend
doppler run -- go run cmd/api/main.go
# Server starts at http://localhost:8081
```

### 5. Run Frontend (Native Mode)
Open another terminal. Install dependencies and start the SvelteKit dev server.
```bash
cd frontend
npm install
doppler run -- npm run dev
# App starts at http://localhost:3000
```

---

## 💻 Development Workflow

### Database Changes
We use a **Schema-First** approach. Do not modify Go structs manually for DB tables.

1.  **Create Migration:**
    ```bash
    migrate create -ext sql -dir db/migrations -seq <name_of_change>
    ```
2.  **Edit SQL:** Modify the generated `.up.sql` and `.down.sql` files in `backend/db/migrations/`.
3.  **Apply Migration:**
    ```bash
    migrate -path db/migrations -database "postgresql://postgres:postgres@localhost:5432/perjadin?sslmode=disable" up
    ```
4.  **Update Queries:** If you changed queries, edit `.sql` files in `backend/db/queries/`.
5.  **Generate Go Code:**
    ```bash
    cd backend
    sqlc generate
    ```

### Running Tests
```bash
cd backend
doppler run -- go test -v ./...
```

### Full Stack Simulation
To test the production build locally (including multi-stage Docker builds):
```bash
# From the project root
doppler run -- podman-compose up --build
```

---

## 📂 Project Structure

```
├── .github/            # CI/CD Workflows
├── backend/            # Go Backend
│   ├── cmd/            # Entry points (api, seeder, etc.)
│   ├── db/             # Migrations and SQL queries
│   ├── internal/       # Private application logic (handlers, services)
│   ├── pkg/            # Public libraries
│   └── sqlc.yaml       # SQLC Configuration
├── frontend/           # SvelteKit Frontend
│   ├── src/lib/        # Shared components and utilities
│   ├── src/routes/     # App pages and API routes
│   └── static/         # Public assets
├── docker-compose.yml  # Container orchestration
└── README.md           # You are here
```

## ⚠️ Troubleshooting

**Q: Connection refused to Database?**
A: Ensure your **local Postgres service** is running and port 5432 is available. Check your Doppler secrets for `DB_HOST` (should be `localhost` when running native Go, or your Host IP when running inside Docker).

**Q: `PieChart is not defined` error?**
A: This was a known issue in the dashboard. Ensure you have pulled the latest changes where the import was fixed.

**Q: Permission denied on `go install`?**
A: Check your `$GOPATH` and `$PATH`. Ensure `Go/bin` is in your system PATH.

---
**Maintained by Application Development Team**
