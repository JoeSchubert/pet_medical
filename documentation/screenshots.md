# Screenshots

Screenshots of the Pet Medical UI for documentation and the README.

| Screen | File | Description |
|--------|------|-------------|
| Login | [login.png](screenshots/login.png) | Sign-in page: email and password. Default credentials: `admin@example.com` / `admin123`. |
| Dashboard | [dashboard.png](screenshots/dashboard.png) | My Pets dashboard with list/cards. On a fresh DB, the app seeds a demo pet (Luna) so the dashboard has sample data. |
| Pet detail | [pet-detail.png](screenshots/pet-detail.png) | Pet profile (Vaccinations tab). |
| Pet detail – Weights | [pet-detail-weights.png](screenshots/pet-detail-weights.png) | Weight history tab. |
| Pet detail – Documents | [pet-detail-documents.png](screenshots/pet-detail-documents.png) | Documents tab. |
| Pet detail – Photos | [pet-detail-photos.png](screenshots/pet-detail-photos.png) | Photos tab. |
| Add Pet | [pet-form-add.png](screenshots/pet-form-add.png) | Add pet form. |
| Edit Pet | [pet-form-edit.png](screenshots/pet-form-edit.png) | Edit pet form. |
| Settings | [settings.png](screenshots/settings.png) | User settings (weight unit, currency, language, change password). |
| Users | [users.png](screenshots/users.png) | User management (admin). |
| Admin options | [admin-options.png](screenshots/admin-options.png) | Default species, breeds, and vaccinations (admin). |

## Demo data

On first run (empty database), the app seeds:

- **Admin** (`admin@example.com` / `admin123`): **5 pets** — Luna (Golden Retriever), Max (Labrador), Whiskers (Cat), Bella (Siamese), Thumper (Rabbit). Each has a **free public-use pet photo** (Unsplash) and sample vaccinations and weight entries.
- **Jane** (`jane@example.com` / `demo123`): **2 pets** — Buddy (Beagle), Mittens (Cat), with photos and one vaccination each.

That gives you a full dashboard and multiple users to try the app with.

## Adding or refreshing screenshots

**Option A – Playwright script (from `frontend/`):**

1. Run the app: `.\deploy.ps1` or `docker compose up -d`, then open **http://localhost:8080**.
2. From `frontend/`: `npm run capture-screenshots` (requires `playwright` and Chromium: `npx playwright install chromium`).
3. Screenshots are written to **documentation/screenshots/** (all pages and pet-detail tabs listed above).

**Option B – Manual:**

1. Run the app and sign in with `admin@example.com` / `admin123`.
2. Capture each view and save into **documentation/screenshots/** (e.g. `dashboard.png`, `pet-detail.png`, `settings.png`).
3. Update the table above and the [README](../README.md#screenshots) if needed.
