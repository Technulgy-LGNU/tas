# T.A.S. (Technulgy Admin Software)

For managing teams and the website.

## Local FusionAuth setup

Keep these application settings:

- **Client authentication: Required** and **PKCE: Required**. The Go backend supplies the client secret and an S256 PKCE verifier.
- **Authorized redirect URL:** `http://localhost:2005/auth/callback` (exact match).
- **Grants:** Authorization Code and Refresh Token.
- **Require registration:** enabled.
- **Security → Generate refresh tokens:** enabled. TAS requests `offline_access`.

Disable **self-service registration** for this admin application, and manually register permitted users for the application in FusionAuth. Require registration by itself is not an admin-role check. Optionally create an `admin` application role, assign it to permitted users, and set `required_role = "admin"` in TAS.

Set the application's OAuth **Logout URL** to `http://localhost:2005/#/login`. For Vite development use `http://localhost:5173/#/login` instead. TAS uses this configured URL when ending the FusionAuth SSO session.

Copy `config.example.toml` to `config.toml` if you don't already have one. Fill in `[auth]`:

```toml
[auth]
fusionauth_url = "http://192.168.1.100:9011"
fusionauth_client_id = "YOUR-APPLICATION-ID"
fusionauth_client_secret = "YOUR-CLIENT-SECRET"
fusionauth_tenant_id = "YOUR-TENANT-ID"
oauth_redirect_uri = "http://localhost:2005/auth/callback"
frontend_url = "http://localhost:2005/#/"
required_role = "" # Optional; e.g. "admin"
```

Use the other laptop's reachable LAN address for `fusionauth_url`, not `localhost`. Both the browser and Go backend must reach it. `localhost` in the callback refers to the computer running the browser, so run your browser on the TAS computer with this setup. Keep the client secret only in backend configuration. Existing `config.toml` files using the keys above work unchanged.

## Run

With PostgreSQL running and `[database]` configured:

```sh
cd frontend
npm ci
npm run build
cd ..
go run ./backend
```

Open `http://localhost:2005`. Unauthenticated visitors are sent to `/#/login`; the sign-in button opens FusionAuth's hosted login page. Successful sign-in returns to the requested page.

For frontend hot reload, set `frontend_url = "http://localhost:5173/#/"`, start the backend, and run `npm run dev` from `frontend`. Open `http://localhost:5173`. Vite proxies `/api` and `/auth` to port 2005; the FusionAuth callback remains `http://localhost:2005/auth/callback`. Use `localhost` consistently, because cookies are shared between these localhost ports. The callback and frontend must share their hostname and scheme.

For Docker, set the database host to `postgres` in `config.toml` and run `docker compose up --build`. TAS is exposed on port 2005.

## Working without FusionAuth

For local development when your FusionAuth laptop is unavailable, set this in `config.toml` and restart the backend:

```toml
[auth]
disable_fusionauth = true
```

TAS automatically uses a **Local admin** identity with `admin` and `editor` roles. No FusionAuth connection, credentials, login, or refresh tokens are needed in this mode. The sidebar identifies local development mode and hides sign-out. Existing frontend session checks and backend role middleware continue to work; CSRF checks remain enabled.

When this option is true, the backend binds to `127.0.0.1:2005` only. Use `http://localhost:2005`, or the existing Vite proxy on `http://localhost:5173`. This mode is for running TAS directly on your laptop, not a Docker port mapping or public deployment. PostgreSQL is still required, and image uploads/deletions still need access to Cloudflare. If omitted, local frontend/callback URLs default to port 2005; keep `frontend_url` pointed at port 5173 when using Vite.

Set `disable_fusionauth = false` and restart to restore FusionAuth. The option defaults to false when absent. It does not change the settings in your FusionAuth instance.

## Authentication behavior and extending the API

- `GET /auth/login` starts authorization; `GET /auth/callback` consumes a single-use state and exchanges the code. `/api/v1/auth/login` is also supported.
- `GET /api/v1/auth/me` returns `{ "user": { "id", "email", "username", "roles" } }` for an authenticated user, or HTTP 401.
- `POST /api/v1/auth/logout` destroys the local session and returns a `logout_url`; the frontend navigates there to end FusionAuth SSO. It works even if tokens have expired.
- Register private endpoints on `v1` **after** `v1.Use(a.Auth.RequireAuth)` in `backend/web/initWeb.go`. Read the verified user with `c.Locals("user").(web.User)` (or `User` within package `web`).
- Use `apiFetch` from `frontend/src/lib/auth.ts` for frontend API calls. It adds the CSRF header to mutations and sends users back to login on HTTP 401. Router navigation checks authentication; active pages also recheck on window focus and every minute.
- Mutations require `X-TAS-CSRF: 1`; when an Origin is present it must match the configured frontend or callback origin. Cross-origin API access is disabled; development uses the Vite proxy.

Access and refresh tokens stay in server memory. Browsers receive an opaque HttpOnly, SameSite=Lax cookie, with Secure automatically enabled for HTTPS callbacks. Sessions last at most eight hours; login attempts expire after ten minutes. Each protected request uses FusionAuth token introspection, checking validity, application registration, tenant, and the optional role. Expired access tokens refresh automatically, including refresh-token rotation. This requires FusionAuth availability: provider outages return 503 and preserve the local session for retry.

This testing setup uses a bounded, in-memory session store for one TAS process. Restarting TAS signs everyone out; multiple replicas need a shared session store. Logout discards local tokens and ends SSO; it does not revoke tokens at FusionAuth. Registration/role changes may only appear when FusionAuth issues a new access token. For production, use HTTPS for TAS and FusionAuth, a shared session store if needed, and choose appropriate FusionAuth token lifetimes/revocation policies.

## Verification

```sh
go test -race ./backend/...
cd frontend
npm run build
npx eslint .
npx oxlint .
```

Backend tests use a mock FusionAuth server to cover PKCE, callback state/replay rejection, registration/tenant/role checks, session expiry, token refresh rotation under concurrent requests, CSRF, outages, and logout. To check the live instance, sign in with a registered account, verify an unregistered account cannot enter, sign out, and confirm `/api/v1/auth/me` returns 401.

References: [FusionAuth application settings](https://fusionauth.io/docs/get-started/core-concepts/types/applications), [token introspection](https://fusionauth.io/docs/apis/oauth/introspect), [logout](https://fusionauth.io/docs/apis/oauth/logout).

## Image library

Open **Images** in the sidebar. The sidebar can collapse to icons on desktop and opens as a navigation drawer on mobile.

| Action | Required application role |
| --- | --- |
| Browse/search images | Any authenticated, registered user |
| Upload, rename, edit alt text | `editor` or `admin` |
| Delete an image | `admin` |

These permissions are enforced on the API as well as in the interface. Leave `[auth].required_role` empty if editors and other registered users should be able to sign in; setting it to `admin` restricts the whole application to administrators.

The existing `[cloudflare]` values are used by the backend. `images_account_id` is the Cloudflare account ID, while `images_delivery_url` is the delivery base including the account hash, for example `https://imagedelivery.net/<account_hash>`. Set `images_variant` to an existing public variant, such as `public`. The API token needs **Images Write** permission for that account. Never expose this token to the frontend.

The backend automatically creates/migrates the PostgreSQL `images` table at startup. It stores the image's name, alt text, Cloudflare ID, original filename/type/size, uploader ID, status, and timestamps. Renaming and editing alt text update TAS metadata without changing the delivery URL. Images are uploaded with public delivery URLs for use on the website; the admin library and its API still require login.

Uploads accept one JPEG, PNG, GIF, or WebP at a time, up to 10 MB. The library lists images uploaded through TAS; it does not automatically import existing Cloudflare account images. Search matches image names, and results are paginated in groups of 24.

An image record is created before contacting Cloudflare. If an upload fails or the server stops before finalizing it, the entry remains marked **Upload incomplete**, preserving its Cloudflare ID for cleanup. An administrator can delete it after two minutes (to avoid racing an active upload) and upload it again. Deletion removes the Cloudflare image before the database record; a failed database deletion can be retried safely. Deleting an image also breaks any website links using that image.

API endpoints (all require an authenticated session):

- `GET /api/v1/images?page=1&search=hero`
- `POST /api/v1/images` — multipart fields `file`, `name`, `altText`
- `PATCH /api/v1/images/:id` — JSON `{ "name": "New name", "altText": "Description" }`
- `DELETE /api/v1/images/:id`

Mutation requests require the existing CSRF header; use the frontend's `apiFetch` helper. The shared `RequireRoles("editor", "admin")` middleware can also protect future role-specific endpoints.

Cloudflare references: [upload API and token permissions](https://developers.cloudflare.com/api/resources/images/subresources/v1/methods/create/), [deleting images](https://developers.cloudflare.com/images/storage/manage-images/delete-images/).

## Website management

Select **Website** in the sidebar for Home, Teams, Participation History, Sponsors, Publications, Blog and SSL. Editors/admins can edit and publish; only admins can delete. Content is bilingual, and drafts are private until published. The blog editor supports ordered text, heading, image, gallery and YouTube blocks, with a preview and scheduled publication dates.

Public resources start at `/website/home?lang=en`; use `lang=de` for German. The public website repository can consume them without FusionAuth. Configure its base URL and allowed origins under `[website]`. Contact submissions are emailed using `[website.contact]`; add the SMTP password and restart to enable sending. SSL stays in the website repository.

See [the website API guide](docs/website-api.md) for all endpoints, response shapes, block rendering, SMTP configuration, permissions and integration testing.
