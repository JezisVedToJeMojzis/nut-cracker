# API

Base URL (dev): `http://localhost:8080`

## Authentication (temporary)

Real auth (Google OAuth + email/password) is not built yet. As a stand-in, the
acting user is identified by the **`X-User-ID`** header containing their numeric
user id. This will be replaced by proper sessions/OAuth without changing endpoint
shapes.

## Users / Profile

### `GET /users/{id}`
The caller's full profile (includes email). Self only (`403` otherwise).

```json
{ "id": 1, "username": "alice", "email": "a@b.com", "created_at": "..." }
```

### `PATCH /users/{id}`
Update the caller's username. Self only. Body `{ "username": "newname" }`.
`400` invalid (must be 2-30 chars). Usernames need not be unique.

### `GET /users/{id}/card`
Public card (id + username) for friend search — lets the caller see who an id
belongs to before adding. `404` if no such user.

```json
{ "id": 2, "username": "bob" }
```

## Endpoints

### `GET /health`
Liveness + database check. `200 {"status":"ok"}` or `503` if DB is unreachable.

### `GET /users/{id}/map`
Returns the user's cracked countries. **Visible only to the user themselves or
an accepted friend** (else `403`).

```json
{ "countries": [ { "country_code": "SK", "cracks": 2 } ] }
```

### `POST /users/{id}/countries/{code}/increment`
Adds one crack for `{code}`, creating the entry at 1 if new. Self only.

```json
{ "country_code": "SK", "cracks": 2 }
```

### `POST /users/{id}/countries/{code}/decrement`
Removes one crack. When the count reaches zero the country is removed from the
map (`removed: true`, `cracks: 0`). Self only. `404` if the country is not on the
map.

```json
{ "country_code": "SK", "cracks": 1, "removed": false }
```

### `DELETE /users/{id}/countries/{code}`
Removes a country from the map entirely. Self only. `204` on success, `404` if
not present.

## Settings (feature flags)

### `GET /users/{id}/settings`
Returns the caller's feature flags (defaults if never saved). Self only (`403`
otherwise).

```json
{ "count_mode": false }
```

### `PUT /users/{id}/settings`
Upserts the caller's settings. Self only. Body mirrors the GET shape.

```json
{ "count_mode": true }
```

## Friends

### `POST /friends/requests`
Send a friend request. Body: `{ "to": "<userId>" }`. If the target already sent
you a pending request, it is auto-accepted.

```json
{ "status": "pending" }   // or "accepted" on auto-accept
```
`400` self-request · `404` user not found · `409` already friends / already pending.

### `POST /friends/requests/{requesterId}/accept`
Accept a pending request addressed to the caller. `200 {"status":"accepted"}`,
`404` if no such pending request.

### `POST /friends/requests/{requesterId}/decline`
Decline a pending request addressed to the caller. `200 {"status":"declined"}`,
`404` if no such pending request.

### `GET /friends`
Accepted friends of the caller.

```json
{ "friends": [ { "id": "...", "username": "bob" } ] }
```

### `GET /friends/requests/incoming`
Pending requests received by the caller. `{ "requests": [ ... ] }`

### `GET /friends/requests/outgoing`
Pending requests sent by the caller. `{ "requests": [ ... ] }`

### `DELETE /friends/{otherId}`
Removes any relationship (unfriend, cancel a sent request, or decline a received
one). `204` on success, `404` if none exists.

## Status codes

| Code | Meaning |
|---|---|
| `401` | Missing `X-User-ID` |
| `403` | Acting on / viewing a map you're not allowed to |
| `404` | Country not on the user's map |
