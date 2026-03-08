# Perjadin Protokol Kemnaker RI

This repository contains the official frontend and backend code for the Perjadin Protokol system.

## Project Structure

This project follows an enterprise monorepo architecture, strictly separating the frontend and backend environments, orchestrated entirely via Docker.

- **`/frontend`**: SvelteKit web application (Node.js).
- **`/backend`**: REST API backend (Golang/Echo).
- **`docker-compose.yml`**: Root orchestration configuration.

## Prerequisites

- **Docker** and **Docker Compose** installed on your system.
*(You do not need Node.js or Go installed locally on your host machine to run this project.)*

## Getting Started

### 1. Build and Run

From the root directory, simply run Docker Compose to build and start the entire stack:

```bash
doppler run -- docker-compose up --build
```

*(Add `-d` to run it in detached/background mode: `doppler run -- docker-compose up -d --build`)*

This single command will:
1. Spin up a **PostgreSQL** database container.
2. Build and start the **Go Backend** container (connecting it to the database).
3. Build and start the **SvelteKit Frontend** container (connecting it to the backend).

### 2. Access the Application

Once the containers are running, you can access the applications at:

- **Frontend Web App**: `http://localhost:3000`
- **Backend API**: `http://localhost:8081`
- **Postgres Database**: `localhost:5432` (User: `postgres`, Password: `postgres`)

### 3. Stopping the Stack

To stop the containers and network:

```bash
docker-compose down
```

*Note: The database data is persisted in a local Docker volume. Running `down` will not delete your database tables or users.*
