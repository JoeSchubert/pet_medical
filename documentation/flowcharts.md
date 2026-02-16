# Flowcharts

Diagrams are in [Mermaid](https://mermaid.js.org/) format. You can view them in any Markdown viewer that supports Mermaid (e.g. GitHub, GitLab, or VS Code with a Mermaid extension).

---

## Authentication flow

```mermaid
sequenceDiagram
  participant User
  participant Browser
  participant API

  User->>Browser: Enter email / password
  Browser->>API: POST /api/auth/login
  API->>API: Validate credentials
  API->>API: Create access + refresh tokens
  API->>Browser: Set-Cookie (access, refresh) + JSON { user, access_token }
  Browser->>Browser: Store access_token in memory
  Browser->>User: Redirect to dashboard

  Note over Browser,API: Subsequent requests
  Browser->>API: Request with Cookie + Authorization: Bearer <token>
  API->>API: Validate token, set user in context
  API->>Browser: JSON response

  alt Token expired (401)
    Browser->>API: POST /api/auth/refresh (with refresh cookie)
    API->>Browser: New cookies + JSON
    Browser->>API: Retry original request
  end
```

---

## Pet and related data flow

```mermaid
flowchart LR
  subgraph Client
    A[Dashboard]
    B[Pet Detail]
    C[Pet Form]
  end

  subgraph API
    D[Pets API]
    E[Vaccinations API]
    F[Weights API]
    G[Documents API]
    H[Photos API]
  end

  subgraph DB
    I[(Pets)]
    J[(Vaccinations)]
    K[(Weights)]
    L[(Documents)]
    M[(Pet Photos)]
  end

  A --> D
  B --> D
  B --> E
  B --> F
  B --> G
  B --> H
  C --> D

  D --> I
  E --> J
  F --> K
  G --> L
  H --> M

  I --> J
  I --> K
  I --> L
  I --> M
```

---

## Deployment (Docker)

```mermaid
flowchart TB
  subgraph Build
    A[Frontend: npm run build]
    B[Generate PWA icons]
    C[Backend: go build]
    D[Copy frontend into backend/static]
    E[Docker image]
    A --> B
    B --> A
    A --> D
    C --> D
    D --> E
  end

  subgraph Runtime
    E --> F[Container: app]
    G[(Volume: pgdata)]
    H[(Volume: photos)]
    I[(Volume: documents)]
    J[Container: db]
    F --> G
    F --> H
    F --> I
    F --> J
    J --> G
  end

  F --> |:8080| K[User / Browser]
```

---

## Defaults and settings flow

```mermaid
flowchart TD
  subgraph Env
    E1[DEFAULT_WEIGHT_UNIT]
    E2[DEFAULT_CURRENCY]
    E3[DEFAULT_LANGUAGE]
  end

  subgraph Config
    C[config.Load]
  end

  subgraph Backend
    S[Seed admin]
    U[Create user]
    A[Auth: applyUserDefaults]
    H[Settings: get/update]
  end

  subgraph DB
    D[(users)]
  end

  E1 --> C
  E2 --> C
  E3 --> C
  C --> S
  C --> U
  C --> A
  C --> H
  S --> D
  U --> D
  A --> D
  H --> D
```

---

## PWA install flow (simplified)

```mermaid
flowchart TD
  A[User opens site in browser]
  A --> B{Installable?}
  B -->|HTTPS, manifest, SW, PNG icons| C[Chrome may show Install / Add to Home screen]
  B -->|No| D[Custom banner: instructions]
  C --> E[User adds to home screen]
  E --> F[App opens in standalone]
  D --> G[User follows instructions]
  G --> F
```
