# API

Base URL (dev): `http://localhost:8080`

## Authentication (temporary)

Real auth (Google OAuth + email/password) is not built yet. As a stand-in, the
acting user is identified by the **`X-User-ID`** header containing their user
UUID. This will be replaced by proper sessions/OAuth without changing endpoint
shapes.

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
