# T.A.S. - Technulgy Admin Software

T.A.S. is a management platform for the Technulgy organization. It provides an internal admin UI and API for inventory, order requests, order lists, member approval, roles, and website content that can be consumed by the public website.

The application is built as:

- Backend: Go, Fiber, GORM, PostgreSQL
- Frontend: Vue 3 Composition API, Vite, Tailwind CSS
- Authentication: FusionAuth for identity, local TAS roles and permissions for authorization
- Image uploads: Cloudflare Images
- Deployment: Docker, Docker Compose, manual GitHub Actions build to GHCR

## Features

### Members, approval, and roles

- FusionAuth handles login and user identity.
- TAS creates a local pending member on first successful login.
- Approved members can access the platform according to local TAS roles.
- Default roles and permissions are seeded during startup.
- Member managers can approve/reject members and assign roles.

The seeded permission model uses domain-action permissions:

- `inventory:view`, `inventory:request`, `inventory:edit`, `inventory:manage`
- `orders:view`, `orders:request`, `orders:edit`, `orders:manage`
- `website:view`, `website:edit`, `website:manage`
- `members:view`, `members:manage`

### Inventory

- Category and subcategory management.
- Expandable category tree in the UI.
- Inventory item CRUD.
- Item fields include name, quantity, vendor ID, product URL, website/source URL, notes, category, and confirmation status.
- Reorder workflow creates an order request from an existing inventory item.
- Received order-list items enter inventory as unconfirmed stock.

### Orders

- Users can create order requests with name, quantity, unit price, URL, shop override, and notes.
- Total price is calculated server-side using integer cents.
- Requests are grouped by shop, using the URL domain by default.
- Managers can approve requests into order lists.
- Managers/editors can create direct order-list entries.
- Order lists can be draft, published, or archived.
- Published lists require manager permission plus an explicit force override to add or edit list items.
- Published list entries support ordered and received tracking.
- Order requests and order lists use GORM soft delete.

### Website content API

TAS manages content that the actual public website can fetch from read-only public JSON endpoints.

Managed content includes:

- Teams with summary text and shared prize records.
- Participation history with competitions, teams, and prizes.
- Sponsor categories and sponsors with linked logos and website URLs.
- Homepage article slots with text and image.
- Image metadata stored locally after upload to Cloudflare Images.

Public endpoints are exposed under `/api/public`.

## Repository layout

```text
.
├── cmd/tas-server/          # Go server entrypoint
├── internal/config/         # TOML config loading
├── internal/database/       # GORM models, migrations, seeded roles
├── internal/web/            # Fiber routes, auth, API handlers
├── frontend/                # Vue 3 + Vite frontend
├── Dockerfile               # Multi-stage production image
├── docker-compose.yml       # Local container deployment
├── config.example.toml      # Example runtime config
└── .github/workflows/       # Manual Docker image build workflow
```

## Requirements

For local development:

- Go `1.26.4`
- Node.js `26`
- npm `11`
- PostgreSQL
- FusionAuth instance, unless using development bypass
- Cloudflare Images account for image upload support

For container deployment:

- Docker
- Docker Compose plugin or compatible Compose implementation

## Configuration

TAS reads `config.toml` from the current working directory. Start by copying the example:

```bash
cp config.example.toml config.toml
```

### Database

Only PostgreSQL is supported.

```toml
[database]
host = "localhost"
port = 5432
user = "technulgy_tas"
pass = "password"
database = "technulgy_tas_db"
timezone = "Europe/Berlin"
```

When using `docker-compose.yml`, set:

```toml
host = "postgres"
```

because the app container reaches the included database through the Compose service name.

### Server

```toml
[server]
host = "0.0.0.0"
port = 8000
```

The backend serves:

- API routes under `/api`
- Vue production build from `frontend/dist`

### CORS

```toml
[cors]
origins = [
  "http://localhost:5173",
  "http://localhost:8000",
  "https://tas.technulgy.com",
  "https://technulgy.com"
]
```

Add every frontend or public website origin that needs to call the API.

### FusionAuth

```toml
[auth]
issuer = "https://auth.example.org"
client_id = "your-client-id"
required_role = ""
dev_allow_admin = false
```

FusionAuth is used for login and JWT validation. TAS stores approval state, roles, and permissions locally.

For local development without FusionAuth, use:

```toml
dev_allow_admin = true
```

This creates a local development admin context. Keep it disabled in production.

### Cloudflare Images

```toml
[cloudflare]
images_account_id = ""
images_api_token = ""
images_delivery_url = "https://imagedelivery.net/<account_hash>"
images_variant = "public"
```

Image uploads use Cloudflare Images. TAS stores the Cloudflare image ID, filename, content type, delivery URL, and variant metadata.

## Local development

### 1. Install frontend dependencies

```bash
cd frontend
npm install
cd ..
```

### 2. Configure PostgreSQL

Create the database and user matching `config.toml`, or adjust `config.toml` to match an existing PostgreSQL instance.

The backend uses GORM `AutoMigrate` on startup and seeds default permissions and roles.

### 3. Start the backend

```bash
go run ./cmd/tas-server
```

The backend listens on the configured server port, usually:

```text
http://localhost:8000
```

### 4. Start the frontend dev server

In another terminal:

```bash
cd frontend
npm run dev
```

The dev frontend listens on:

```text
http://localhost:5173
```

Vite proxies `/api` to `http://localhost:8000`.

## Build and test

Run backend tests:

```bash
go test ./...
```

Build the frontend:

```bash
cd frontend
npm run build
```

Build the full production Docker image:

```bash
docker build -t tas:local .
```

## Docker deployment

The provided Dockerfile builds both frontend and backend:

1. Node stage installs frontend dependencies and runs `npm run build`.
2. Go stage builds `./cmd/tas-server`.
3. Runtime stage copies the Go binary and `frontend/dist` into `/app`.

The runtime container expects:

```text
/app/config.toml
```

### Docker Compose

`docker-compose.yml` runs:

- `tas`: the application
- `postgres`: PostgreSQL 17 Alpine
- `tas-postgres-data`: persistent database volume

Before starting Compose, make sure `config.toml` exists and uses:

```toml
[database]
host = "postgres"
port = 5432
user = "technulgy_tas"
pass = "password"
database = "technulgy_tas_db"
```

Then run:

```bash
docker compose up --build
```

The app will be available at:

```text
http://localhost:8000
```

The Compose file bind-mounts the local config file:

```yaml
./config.toml:/app/config.toml:ro
```

This means config changes can be made outside the image. Restart the container after changing config.

## GitHub Actions Docker build

The workflow at `.github/workflows/docker-build.yml` is manual-only.

It runs when triggered from the GitHub Actions UI with `workflow_dispatch`.

It always builds and pushes a multi-architecture image to GitHub Container Registry:

- `linux/amd64`
- `linux/arm64`

The Dockerfile uses Buildx target arguments so the Go binary is compiled for the target image architecture.
The GitHub workflow installs QEMU before Buildx so the arm64 runtime image can be built on the amd64 GitHub runner.

Images are tagged as:

```text
ghcr.io/<owner>/<repo>:<image_tag>
ghcr.io/<owner>/<repo>:<commit_sha>
```

The manual input is:

- `image_tag`: tag to publish, default `latest`

The workflow uses `GITHUB_TOKEN` with `packages: write` permission.

## API overview

Protected API routes are under `/api`.

Important protected route groups:

```text
/api/auth/config
/api/me
/api/members
/api/roles
/api/inventory/categories
/api/inventory/items
/api/orders/requests
/api/orders/lists
/api/images
/api/website/teams
/api/website/competitions
/api/website/prizes
/api/website/sponsor-categories
/api/website/sponsors
/api/website/home
```

Public website routes:

```text
/api/public/teams
/api/public/participation-history
/api/public/sponsors
/api/public/home
```

## Operational notes

- `config.toml` should not be committed with production secrets.
- `dev_allow_admin` must be `false` in production.
- Published order lists are protected from normal item edits. Managers must use the explicit force override for item add/edit operations.
- Order requests and order lists are soft-deleted through GORM `DeletedAt`.
- The included Compose database credentials are development defaults; change them for real deployments.
- Cloudflare image uploads require a valid API token with permission to upload images.

## Common commands

```bash
# Backend tests
go test ./...

# Frontend development
cd frontend && npm run dev

# Frontend production build
cd frontend && npm run build

# Full Docker image build
docker build -t tas:local .

# Compose deployment
docker compose up --build
```
