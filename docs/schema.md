# Database Schema

Managed with [golang-migrate](https://github.com/golang-migrate/migrate).
Migration files live in `internal/db/migrations/`.

## Tables

### `users` — accounts

| column | type | notes |
|---|---|---|
| `id` | `bigint` PK (identity) | sequential, indexed; the short, shareable user id used for search |
| `email` | `text` unique | login identifier |
| `username` | `text` | editable display name shown to others (not unique) |
| `password_hash` | `text` nullable | null for OAuth-only users |
| `created_at` | `timestamptz` | |

The numeric `id` is the permanent account number (stable, orderable, searchable).
`username` is the editable friendly name. All other tables reference `users.id`.

### `user_identities` — external logins

Supports Google OAuth now, and any future provider, without reshaping `users`.

| column | type | notes |
|---|---|---|
| `user_id` | `uuid` → users.id | cascade on delete |
| `provider` | `text` | e.g. `google` |
| `provider_user_id` | `text` | stable ID from the provider |
| `created_at` | `timestamptz` | |

Primary key: `(provider, provider_user_id)`.

### `countries` — reference list

ISO 3166-1 alpha-2 codes; the same codes used by the frontend SVG map.

| column | type | notes |
|---|---|---|
| `code` | `text` PK | e.g. `SK` |
| `name` | `text` | e.g. `Slovakia` |

### `user_countries` — colored countries

The heart of the app. One row per country a user has "cracked."

| column | type | notes |
|---|---|---|
| `user_id` | `uuid` → users.id | cascade on delete |
| `country_code` | `text` → countries.code | |
| `cracks` | `integer` default 1, `CHECK (cracks >= 1)` | people cracked with from this country |
| `created_at` | `timestamptz` | first marked |
| `updated_at` | `timestamptz` | last count change |

Primary key: `(user_id, country_code)` — prevents duplicates.

### `friendships` — friend connections

Request + accept model.

| column | type | notes |
|---|---|---|
| `user_id` | `uuid` → users.id | requester |
| `friend_id` | `uuid` → users.id | recipient |
| `status` | `text` | `pending` or `accepted` |
| `created_at` | `timestamptz` | |

Primary key: `(user_id, friend_id)`. Constraint: `user_id <> friend_id`.

### `user_settings` — per-user feature flags

One row per user; one boolean column per toggleable feature. Add a column (and a
migration) per new feature.

| column | type | notes |
|---|---|---|
| `user_id` | `uuid` PK → users.id | cascade on delete |
| `count_mode` | `boolean` default false | track cracks-per-country when on |
| `created_at` | `timestamptz` | |
| `updated_at` | `timestamptz` | |

## Privacy Rule

A user's map is visible **only to themselves and their accepted friends**. This is
enforced in application authorization on every map-read endpoint (return `403`
otherwise) — it is not a schema-level constraint.

## Auth

Email/password auth is implemented with DB-backed sessions. Google OAuth is
planned (the `user_identities` table already supports external providers).

- `users.email_verified` (`boolean`) — set true after the user clicks the
  verification link.
- `sessions` — `id` (opaque token / cookie value) PK, `user_id` → users.id,
  `created_at`, `expires_at`. The session cookie is `nc_session` (HttpOnly).
- `user_tokens` — one-time tokens: `token` PK, `user_id`, `kind`
  (`verify` | `reset`), `expires_at`, `created_at`.

Passwords are hashed with bcrypt. `users.password_hash` is nullable (OAuth-only
accounts have no password).
