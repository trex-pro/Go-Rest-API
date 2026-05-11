# Project API (Go + MySQL)

HTTPS REST API for managing **students**, **teachers**, and **executives (execs)**, with **JWT cookie auth** and a set of security/performance middlewares.

## Tech stack

- **Go**: module `project-api` (see `go.mod`)
- **Database**: MySQL (`github.com/go-sql-driver/mysql`)
- **Auth**: JWT (cookie name: `JWT`)
- **Server**: HTTPS via `ListenAndServeTLS`

## Routes

### Students

- `GET /students`
- `POST /students`
- `PUT /students`
- `PATCH /students`
- `DELETE /students`
- `GET /students/{id}`
- `PUT /students/{id}`
- `PATCH /students/{id}`
- `DELETE /students/{id}`

### Teachers

- `GET /teachers`
- `POST /teachers`
- `PUT /teachers`
- `PATCH /teachers`
- `DELETE /teachers`
- `GET /teachers/{id}`
- `PUT /teachers/{id}`
- `PATCH /teachers/{id}`
- `DELETE /teachers/{id}`
- `GET /teachers/{id}/students`

### Execs (users/auth)

- `GET /execs`
- `POST /execs`
- `PATCH /execs`
- `GET /execs/{id}`
- `PATCH /execs/{id}`
- `DELETE /execs/{id}`
- `POST /execs/{id}/updatepassword`
- `POST /execs/login` (public)
- `POST /execs/logout`
- `POST /execs/forgotpassword` (public)
- `POST /execs/resetpassword/reset/{resetcode}` (public)

## Environment variables

The server reads configuration from environment variables (loaded from an embedded `.env` in `cmd/api/.env` during development).

### MySQL

- **`DB_USER`**: MySQL username
- **`DB_PASS`**: MySQL password
- **`DB_NAME`**: database name
- **`HOST`**: database host (example: `127.0.0.1`)
- **`DB_PORT`**: database port (example: `:3306`)

### Server / TLS

- **`SERVER_PORT`**: server listen address (example: `:3000`)
- **`CERT_FILE`**: path to TLS certificate PEM
- **`KEY_FILE`**: path to TLS private key PEM

### Auth / reset

- **`JWT_SECRET`**: HMAC secret for signing JWTs
- **`JWT_EXPIRY`**: Go duration string (examples: `15m`, `24h`, `3000s`)
- **`RESET_CODE_EXPIRY`**: Go duration string (example: `300s`)

## Run locally (development)

1) Make sure MySQL is running and the target database exists.

2) Update `cmd/api/.env` with your local values (do **not** commit real secrets).

3) Run:

```bash
go run ./cmd/api
```

The server starts on `SERVER_PORT` and serves **HTTPS**.

## Run with Docker

This repo includes a multi-stage `Dockerfile` that builds a small runtime image.

### Build

```bash
docker build -t project-api .
```

### Run

Mount TLS files and pass config via env vars:

```bash
docker run --rm -p 3000:3000 \
  -e SERVER_PORT=":3000" \
  -e CERT_FILE="/run/tls/cert.pem" \
  -e KEY_FILE="/run/tls/key.pem" \
  -e DB_USER="root" \
  -e DB_PASS="your_password" \
  -e DB_NAME="school" \
  -e HOST="host.docker.internal" \
  -e DB_PORT=":3306" \
  -e JWT_SECRET="change_me" \
  -e JWT_EXPIRY="15m" \
  -e RESET_CODE_EXPIRY="300s" \
  -v "$PWD/cmd/api/cert.pem:/run/tls/cert.pem:ro" \
  -v "$PWD/cmd/api/key.pem:/run/tls/key.pem:ro" \
  project-api
```

Notes:

- The server is **HTTPS-only**, so use `https://localhost:3000`.
- With a self-signed cert, clients must trust the cert or skip verification for local testing.

## Auth notes

- Protected routes expect a `JWT` cookie.
- The following routes are intentionally excluded from JWT middleware:
  - `POST /execs/login`
  - `POST /execs/forgotpassword`
  - `POST /execs/resetpassword/reset/{resetcode}`

