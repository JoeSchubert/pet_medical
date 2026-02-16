# Pet Medical

A digital pet health portfolio: track pets, vaccinations, weight, documents, and photos. Built with a Go backend, PostgreSQL, and a React PWA frontend.

## Inspiration

This application was inspired by [Robipet](https://github.com/Anghios/robipet), so the UI looks similar. However, the code is original. This project was started due to issues with translations in Robipet and features that weren't present yet that I wanted. Rather than fixing the issues in that codebase, I started my own project inspired by it.

## Features

- **Authentication**: JWT access + refresh tokens (httpOnly cookies). Log in with **email** and password. Default admin: `admin@example.com` / `admin123` — change after first login.
- **Pets**: Add, edit, delete pets with name, species, breed, DOB, gender, color, microchip, notes, and profile photo.
- **Vaccinations**: Per-pet vaccination records with name, date administered, next due, cost, and optional expiry hints.
- **Weight**: Per-pet weight history with date and optional “approximate” flag; dashboard and detail views support lbs/kg.
- **Documents**: Upload and store pet documents with editable names; list and delete. Text is extracted from PDFs, DOCX, RTF, and (if [Tesseract](https://github.com/tesseract-ocr/tesseract) is installed) from images; you can **search by name or document content** in the Documents tab.
- **Photos**: Upload pet photos (file picker or camera on mobile), set one as profile picture.
- **PWA**: Installable on mobile and desktop (Add to Home screen / Install app); works offline for cached assets; responsive layout with mobile nav.
- **Settings**: Per-user weight unit (lbs/kg), currency, and language (en, es, fr, de). Defaults are configurable via environment variables.

## Quick start with Docker

1. **Build and run** (from repo root):
   ```powershell
   .\deploy.ps1
   ```
   Or:
   ```cmd
   deploy.bat
   ```
   On subsequent runs, the script stops existing containers and can optionally remove the database volume.

2. Open **http://localhost:8080** and sign in with **Email** `admin@example.com` and **Password** `admin123`.

## Scripts

| Script | Purpose |
|--------|--------|
| `build.ps1` / `build.bat` | Build the Docker image (no run). |
| `deploy.ps1` / `deploy.bat` | Stop containers, optionally remove DB volume, then build and run with Docker Compose. |

## Configuration

Environment variables can be set in `docker-compose.yml`, a `.env` file, or the host. See **docker-compose.sample.yml** for a full list with comments.

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:postgres@db:5432/pet_medical?sslmode=disable` (Compose) |
| `JWT_SECRET` | Secret for signing JWTs | (required in production; use e.g. `openssl rand -base64 32`) |
| `JWT_ACCESS_TTL_MIN` | Access token lifetime (minutes) | `30` |
| `JWT_REFRESH_TTL_DAYS` | Refresh token lifetime (days) | `7` |
| `CORS_ORIGINS` | Allowed origins (comma-separated or `*`) | `*` |
| `ENABLE_DEBUG_LOGGING` | Enable debug logs | `false` |
| `SYSTEM_LANGUAGE` | Backend log message language | `en` |
| **`DEFAULT_WEIGHT_UNIT`** | Default for new users: `lbs` or `kg` | `lbs` |
| **`DEFAULT_CURRENCY`** | Default for new users (e.g. USD, EUR) | `USD` |
| **`DEFAULT_LANGUAGE`** | Default for new users (e.g. en, es, fr, de) | `en` |
| `UPLOAD_DIR` | Directory for uploaded photos and documents | `./uploads` (or `/app/uploads` in Docker) |
| `GOOGLE_CLIENT_ID` | Google OAuth2 client ID (optional; e.g. for oauth2-proxy) | — |
| `GOOGLE_CLIENT_SECRET` | Google OAuth2 client secret (optional) | — |
| `GOOGLE_REDIRECT_URI` | Google OAuth2 redirect URI (optional) | — |
| **`TRUSTED_PROXIES`** | Comma-separated proxy IPs or CIDRs; when request is from one of these, forwarded auth headers (e.g. `X-Forwarded-Email`) are trusted and used to log in by email | — |
| **`FORWARDED_EMAIL_HEADER`** | Header name for email from proxy | `X-Forwarded-Email` |
| **`FORWARDED_USER_HEADER`** | Header name for display name from proxy | `X-Forwarded-User` |
| `DEVELOPMENT` | Set to `true` (or `1`/`yes`) for local dev: relaxes security (cookies not Secure, no HSTS, no JWT/CORS warnings). Default `false`. | `false` |
| `SECURE_COOKIES` | When **not** in development: set to `true` for HTTPS so auth cookies use the Secure flag. Defaults to `true` when `DEVELOPMENT` is false. Ignored when `DEVELOPMENT=true`. | `true` (when not dev) |
| **`SAME_SITE_COOKIE`** | Cookie SameSite: `lax`, `strict`, or `none`. Use `none` behind a reverse proxy (e.g. Kubernetes Ingress + oauth2-proxy) if you get 401s after login when navigating; requires HTTPS (Secure cookies). Default `lax`. | `lax` |

When running behind **oauth2-proxy** (or similar), set `TRUSTED_PROXIES` to the proxy’s IP or CIDR so the app trusts `X-Forwarded-Email` (and optional `X-Forwarded-User` for display name). Users are matched by email to existing accounts or auto-created with default role/settings. For Google-based proxy login, set these in the proxy; the app does not require them to run (they are optional).

**Production**: Do **not** set `DEVELOPMENT=true`. Set a strong `JWT_SECRET` (e.g. `openssl rand -base64 32`), restrict `CORS_ORIGINS` to your frontend origin(s), use a dedicated database and backup strategy, and change the default admin password after first login. With `DEVELOPMENT` unset or false, cookies default to Secure and the app logs warnings at startup when JWT_SECRET or CORS_ORIGINS use default/permissive values.

**Development**: Set `DEVELOPMENT=true` (or `1`/`yes`) for local development to allow HTTP cookies, disable HSTS, and silence strict security warnings.

**Kubernetes / reverse proxy**: If you see 401s after logging in when navigating (e.g. `/auth/me` or `/auth/refresh` failing), set `SAME_SITE_COOKIE=none` so the browser sends auth cookies on requests behind the proxy. Use `/health` (or `/api/health`) for liveness and readiness probes.

## Local development

- **Backend** (from `backend/`): Go 1.21+ and PostgreSQL. Run with `go run ./cmd/api` (set `DATABASE_URL` if needed).
- **Frontend** (from `frontend/`): `npm install` then `npm run dev`. Vite proxies `/api` to `http://localhost:8080`.

## Tech stack

- **Backend**: Go 1.21+, Gorilla Mux, GORM, PostgreSQL, JWT + refresh tokens.
- **Frontend**: React 19, TypeScript, Vite, React Router, Vite PWA plugin, Recharts.
- **Deploy**: Docker and Docker Compose (single app image with embedded frontend + Postgres).

## Documentation

Detailed documentation lives in the **[documentation/](documentation/)** folder:

- [Tech stack](documentation/tech-stack.md) — Backend and frontend technologies and structure.
- [Program flow & architecture](documentation/program-flow.md) — Request flow, auth, and data flow.
- [Flowcharts](documentation/flowcharts.md) — Mermaid diagrams for auth, pet management, and deployment.
