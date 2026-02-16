# Program Flow & Architecture

## Request flow (high level)

1. **Browser** loads the SPA from the Go server (or a reverse proxy). All API calls go to `/api/*`.
2. **CORS** middleware allows configured origins (or `*`).
3. **Logging** middleware logs the request.
4. **Routes**:
   - Public: `/api/auth/login`, `/api/auth/refresh`, `/api/auth/logout`, `/api/health`.
   - Protected: everything else under `/api` (requires valid JWT from cookie or `Authorization: Bearer`).
5. **Auth middleware** reads the token from the `Authorization` header or the `access_token` cookie, validates it, and puts the user into the request context.
6. **Handler** reads/writes DB (GORM) and returns JSON (or file for uploads).

## Authentication flow

- **Login**: POST `/api/auth/login` with email/password → server validates, creates access + refresh tokens, sets httpOnly cookies for both, returns user + access token in body. Frontend stores the access token in memory and uses it in the `Authorization` header for subsequent requests.
- **Protected request**: Client sends cookie (and optionally `Authorization: Bearer <token>`). If the token is missing or expired (401), the frontend can call POST `/api/auth/refresh` with the refresh cookie to get new tokens and retry.
- **Logout**: POST `/api/auth/logout` clears cookies; frontend clears in-memory token.

New users (seed admin and admin-created users) get default weight unit, currency, and language from server config (env: `DEFAULT_WEIGHT_UNIT`, `DEFAULT_CURRENCY`, `DEFAULT_LANGUAGE`). When a user’s settings are empty, the API normalizes them using these same defaults.

## Data flow (typical)

- **Pets**: List (GET), create (POST), get one (GET), update (PUT), delete (DELETE). Pet has many vaccinations, weight entries, documents, photos. Ownership is enforced by `user_id` on the pet.
- **Vaccinations / Weights / Documents / Photos**: All scoped by `pet_id`; create/list/update/delete with ownership checked via the pet’s `user_id`.
- **Settings**: Per-user; GET/PUT for current user; admins can GET/PUT another user’s settings.
- **Files**: Photos and documents are uploaded with multipart/form-data; files are stored under `UPLOAD_DIR` and metadata (and file path) in the database. Serving is via a dedicated handler under `/api/uploads/`.

## Frontend flow

- **AuthContext**: On load, calls GET `/api/auth/me`; on 401, calls refresh then retries. Exposes `user`, `login`, `logout`, `refreshUser`.
- **Routes**: Login page (public), then a protected layout with nested routes: Dashboard, Pet detail/edit, Users (admin), Admin options, Settings.
- **PWA**: Install prompt and Settings “Install app” section use `PWAInstallContext` (standalone/mobile/deferred prompt). Service worker is registered in `main.tsx` for installability and caching.

## Startup (backend)

1. Load config from env.
2. Initialize i18n and debug logging.
3. Connect to PostgreSQL (with retry).
4. Run GORM AutoMigrate for all models.
5. Seed default admin (if no users) and default dropdown options.
6. Register routes and start HTTP server.
7. Static files (embedded SPA) and upload serving are part of the same server.
