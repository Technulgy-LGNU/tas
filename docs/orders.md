# Orders

Open **Orders** in the sidebar. The overview links to Website management and displays open lists, pending requests, closed lists, and part lines still to order. Prices are EUR throughout. Amounts are whole units; totals are amount × unit price. Requests are excluded from list totals until approved.

## Roles

Create and assign the `order_admin` and `order_request` application roles in FusionAuth. The existing `editor` and `admin` roles also apply. These names are fixed in TAS. Leave `[auth].required_role = ""` to let registered users with these roles sign in; setting it to `admin` blocks other users from the whole application. Sign in again after assigning new roles if the current access token still has the previous roles.

| Action | order_request | editor | order_admin / admin |
| --- | --- | --- | --- |
| View lists, approved parts and standard parts | Yes | Yes | Yes |
| Submit a request to an open list | Yes | Yes | Yes |
| View requests | Own only | All | All |
| Edit/withdraw own pending requests | Yes | Yes | Yes |
| Add/edit/remove parts in an open list | No | Yes | Yes |
| Review, adjust and accept/reject pending requests | No | Yes | Yes |
| Create, rename, delete, close or reopen lists | No | No | Yes |
| Manage list categories and standard parts | No | No | Yes |
| Change a closed list or mark parts ordered | No | No | Yes |

Permissions are enforced by the API as well as the interface. Users without one of these roles cannot access Orders. Local development mode's `admin` identity has full access.

## Workflow

1. An order admin creates a named list and optional categories for teams, projects or shops. Each part can have one category, or remain uncategorized.
2. Editors and order admins add parts directly. Teams submit requests. Each entry has a name, amount, unit price, shop, optional product link and optional category. The standard-parts picker copies saved values into either form for adjustment.
3. Editors/order admins review requests. Acceptance can adjust the submitted details and add a review note; it atomically creates a list part linked to the request. Rejected and withdrawn requests stay in the history. Requesters can only change their own pending requests.
4. Resolve all pending requests before closing the list. Closed lists are grouped by shop, with per-shop and overall totals. Order admins/admins mark each line ordered; TAS records who marked it and when. Checkmarks track manual purchasing; TAS does not place orders with shops.
5. Editing an ordered part clears its checkmark. Reopening preserves existing checkmarks and allows new requests and editor changes again. Categories still used by parts or pending requests cannot be deleted.

Standard parts are shared templates managed by order admins/admins. Changing or deleting a template leaves previously copied order parts unchanged. There is no Excel import in this version.

## Storage and deployment

No new TOML settings are needed. Startup automatically migrates PostgreSQL tables `order_lists` and `standard_parts`. Each list stores its categories, parts and request history together in JSONB. List changes use a database row lock and version check so competing edits cannot silently overwrite one another and a request cannot be accepted twice. A stale version returns HTTP 409; refresh the list before retrying.

Prices are integer cents (`unitPriceCents`), never floating-point monetary values. Amounts are limited to 1–10,000 units, unit prices to €0–€1,000,000, and each list to 100 categories, 1,000 part lines and 1,000 requests including history. Lists and standard parts are currently loaded in full; search/filtering happens in the browser.

Publish and deploy a fresh image with the existing GitHub workflow and Compose setup. The startup migrations require the same database permissions as the existing website/image migrations.

## Private API

All routes use `/api/v1/orders`, require an authenticated session with an order role, and require `X-TAS-CSRF: 1` for mutations. Use the frontend's `apiFetch` helper.

| Method/path | Result/input |
| --- | --- |
| `GET /stats` | `openLists`, `closedLists`, `pendingRequests`, `remainingParts`, `openTotalCents`, `remainingTotalCents`, `currency` |
| `GET /lists` | `{ "lists": [...] }` summaries with counts and cent totals |
| `POST /lists` | `{ "name": "German Open 2027" }` → `{ "list": ... }` |
| `GET /lists/:id` | `{ "list": ... }`, including `content.categories`, `content.parts`, `content.requests` |
| `DELETE /lists/:id` | `{ "version": 3 }` → 204 |
| `POST /lists/:id/actions` | Command below → updated `{ "list": ... }` |
| `GET /standard-parts` | `{ "parts": [...] }` |
| `POST /standard-parts` | Part fields → `{ "part": ... }` |
| `PUT /standard-parts/:id` | Part fields plus current `version` → `{ "part": ... }` |
| `DELETE /standard-parts/:id` | `{ "version": 1 }` → 204 |

`pendingRequests`/`pendingCount` and returned request arrays include only the caller's requests for users with only `order_request`. Approved parts are visible to everyone with order access. `remainingParts` counts unmarked part lines in closed lists, not individual units.

Example action (a unit price of €1.99):

```json
{
  "version": 3,
  "action": "add_part",
  "part": {
    "name": "Motor",
    "amount": 3,
    "unitPriceCents": 199,
    "shop": "Robot Shop",
    "link": "https://example.org/motor",
    "categoryId": ""
  }
}
```

Actions and additional fields:

- `rename`: `name`; `close` and `reopen`: no additional fields.
- `add_category`: `name`; `rename_category`: `targetId`, `name`; `delete_category`: `targetId`.
- `add_part`, `request_part`: `part`; `edit_part`, `edit_request`: `targetId`, `part`.
- `delete_part`, `withdraw_request`: `targetId`.
- `accept_request`: `targetId`, complete approved `part`, optional `note`; `reject_request`: `targetId`, optional `note`.
- `set_ordered`: `targetId`, `ordered` boolean; available on closed lists only.

Every action requires the current list `version`. Standard-part fields match the example's `part` except categories are specific to lists and are ignored for templates. Actor IDs, review status, ordering timestamps, and currency are set by the server.

## Verification

```sh
TAS_TEST_CONFIG=/path/to/test-config.toml go test -race ./backend/...
cd frontend
npm run build
npx eslint .
npx oxlint .
```

The PostgreSQL tests use isolated, rollback-only schemas and cover request approval, privacy, role enforcement, category dependencies, closed-list changes, exact cent totals, stale versions, template copying and CSRF. Without `TAS_TEST_CONFIG`, database integration tests are skipped. The GitHub workflow supplies a disposable database.
