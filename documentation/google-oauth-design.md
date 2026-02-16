# Google OAuth Login – What’s Involved

This document outlines what’s required to add **Google OAuth** login and **match users by email** to existing accounts.

---

## Goals

1. User can click “Sign in with Google” and complete the Google OAuth flow.
2. If a user with that **email** already exists (e.g. created by an admin), they are logged into that account.
3. Optionally: if no user exists, either **create** a new account or **deny** and require pre-registration.

---

## High-level flow

1. User clicks “Sign in with Google” on the login page.
2. Frontend redirects to **backend** (e.g. `GET /api/auth/google`).
3. Backend redirects to **Google** with `client_id`, `redirect_uri`, `scope=email profile`.
4. User signs in with Google; Google redirects back to **backend** (e.g. `GET /api/auth/google/callback?code=...`).
5. Backend exchanges `code` for tokens, fetches user info (email, name), then:
   - **Find user by email.** If found → issue your app’s JWT + refresh (same as password login), set cookies, redirect to frontend (e.g. `/`).
   - **If not found** → either create a new user (with that email, derived username, no password) and log them in, or redirect to login with an error “No account for this email”.
6. Frontend loads; cookies are already set, so `GET /api/auth/me` or refresh returns the user.

---

## 1. Backend

### 1.1 User model and password

Today every user has a **required** `PasswordHash`. For OAuth-only users you have two options:

- **Option A (recommended):** Allow **empty** `PasswordHash` for OAuth-only users.
  - In the **password login** handler: if `u.PasswordHash == ""`, return “invalid credentials” (don’t try to check password).
  - No DB migration if the column already allows empty string; if it’s `NOT NULL`, add a migration to allow `''` or make the column nullable.
- **Option B:** Add a separate table, e.g. `user_auth_providers (user_id, provider, provider_user_id)`, and keep `PasswordHash` required. OAuth flow finds or creates a user and links the Google id in this table. Slightly more flexible for multiple providers later.

For “match by email” only, Option A is enough.

### 1.2 Config

- **`GOOGLE_CLIENT_ID`** – From Google Cloud Console (OAuth 2.0 Client ID, type “Web application”).
- **`GOOGLE_CLIENT_SECRET`** – Same client.
- **`GOOGLE_REDIRECT_URI`** – Must match the redirect URI configured in Google (e.g. `https://yourdomain.com/api/auth/google/callback` or `http://localhost:8080/api/auth/google/callback` for dev).

Add these to `internal/config` and load from env.

### 1.3 OAuth endpoints

- **`GET /api/auth/google`**  
  - Build Google’s authorization URL (`https://accounts.google.com/o/oauth2/v2/auth?client_id=...&redirect_uri=...&response_type=code&scope=email+profile`).  
  - Redirect the user (HTTP 302) to that URL.

- **`GET /api/auth/google/callback`**  
  - Read `code` from query.  
  - Exchange `code` for access token (POST to `https://oauth2.googleapis.com/token`).  
  - Use the access token to get user info (GET `https://www.googleapis.com/oauth2/v2/userinfo` or userinfo endpoint).  
  - From the response you get **email** (and optionally name, picture).  
  - **Lookup:** `SELECT * FROM users WHERE email = ?` (normalize email: trim, lowercase).  
  - **If found:**  
    - Call the same “issue JWT + refresh + set cookies” logic you use after password login.  
    - Redirect to the frontend (e.g. `https://yourdomain.com/` or `http://localhost:5173/`) so the SPA loads with cookies set.  
  - **If not found:**  
    - **Option 1 (match-only):** Redirect to login with a query param like `?error=no_account` and show “No account for this email. Ask an admin to create one.”  
    - **Option 2 (auto-create):** Create a user with `email = google email`, `display_name = derived from email or name` (e.g. local part of email, or slug from name), `password_hash = ""`, `role = "user"`, and default weight_unit/currency/language from config; then issue JWT + refresh and redirect to `/`.

Implement the “issue tokens and set cookies” part in a shared helper so both password login and Google callback use it (no duplication).

### 1.4 Security

- **State:** Send a random `state` in the authorization URL and verify it in the callback to avoid CSRF. Store `state` in a short-lived cookie or server-side store keyed by session.
- **Redirect URI:** Validate that the redirect URI used in the callback matches your config (and that it’s your own domain). Google will only send the code to the URI you registered.
- **Email verification:** Google’s OAuth email is considered verified; you can trust it for matching.

### 1.5 Display name for new users

If you auto-create users, you need a unique **display name**. Options:

- Use the **local part** of the email (e.g. `jane` from `jane@gmail.com`); if collision, append a number.
- Or use a **slug** from Google’s name (e.g. “Jane Doe” → `jane.doe`) and de-duplicate.
- Ensure uniqueness with a unique index on `display_name` and retry with a suffix on conflict.

---

## 2. Frontend

### 2.1 Login page

- Add a **“Sign in with Google”** button/link.
- It should **navigate the browser** to the backend’s start URL (e.g. `window.location.href = '/api/auth/google'` or full URL if frontend is on another origin).  
  Do **not** open a popup unless you implement the popup flow and postMessage; a full redirect is simpler and works well on mobile.

### 2.2 After callback

- Backend redirects to your frontend root (e.g. `/?logged_in=1`) or to a dedicated “oauth success” route with cookies already set.
- Your app loads; `AuthContext` or equivalent runs and calls `GET /api/auth/me` or refresh. The request includes the cookies, so the backend returns the user and the frontend shows the logged-in state. No extra frontend logic is required for the “session” beyond what you already have.

### 2.3 Error handling

- If backend redirects with `?error=no_account`, show a message like “No account found for this Google email. Please ask an admin to create an account for you,” and keep the user on the login page.

---

## 3. Matching existing users by email

- **Normalize email:** Trim and lowercase before lookup (Google often returns lowercase; your DB might store mixed case).  
  Optionally: use a single canonical form (e.g. Gmail dot-aliasing if you care).
- **Lookup:** `WHERE LOWER(TRIM(email)) = ?` or store a normalized email and match on that.
- **One account per email:** Your schema already has a unique index on `email`, so one user per email. Matching by email is therefore a single row.

No extra “linking” step is needed if you only support Google: the user **is** the row identified by that email.

---

## 4. Optional: “Link Google to my existing account”

If later you want “I already have a password account; link my Google to it,” that’s a different feature:

- Require the user to be **logged in** (password or existing OAuth).
- Add an endpoint, e.g. `POST /api/auth/link-google` with the OAuth `code` (or frontend sends the Google ID token and backend verifies it).
- Backend verifies the token, gets Google email, finds current user from JWT, and either:
  - Stores `google_id` (or similar) on the user row or in `user_auth_providers`, or  
  - Just ensures the Google email matches the current user’s email and then stores the link.

For “match by email only” and optional auto-create, you don’t need this; the above flow is enough.

---

## 5. Checklist summary

| Area | Task |
|------|------|
| **Google Cloud** | Create OAuth 2.0 client (Web), add redirect URI(s), copy client ID and secret. |
| **Backend config** | Add `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URI` (and optional `FRONTEND_URL` for redirect after login). |
| **User model** | Allow empty `PasswordHash` (or nullable); in password login, reject if empty. |
| **Backend auth** | Implement `GET /api/auth/google` (redirect to Google) and `GET /api/auth/google/callback` (exchange code, get email, find/create user, issue JWT + refresh, set cookies, redirect to frontend). Use shared “issue tokens + set cookies” helper. Add `state` for CSRF. |
| **Frontend** | Add “Sign in with Google” that redirects to `GET /api/auth/google`. Handle `?error=no_account` on the login page. |
| **Docs / env** | Document the new env vars in README and `docker-compose.sample.yml`. |

---

## 6. Dependencies

Backend only needs the **HTTP client** (to call Google’s token and userinfo endpoints). Go’s standard library is enough; no need for a special “OAuth library” unless you want one for convenience (e.g. `golang.org/x/oauth2`).

Using `golang.org/x/oauth2`:

- You get a config with `ClientID`, `ClientSecret`, `RedirectURL`, `Scopes: ["email", "profile"]`, `Endpoint: google.Endpoint`.
- `AuthCodeURL(state)` for the redirect.
- `Exchange(ctx, code)` for the token.
- Use the token to call the userinfo URL (or use `oauth2.NewClient` and do a GET to Google’s userinfo endpoint).

This keeps the implementation small and maintainable while supporting “login with Google and match by email” (and optional auto-create).
