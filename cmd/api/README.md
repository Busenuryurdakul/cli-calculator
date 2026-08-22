# Bookmarks REST MVP

Entry point: `go run ./cmd/api` (default `http://127.0.0.1:8081`).

Address can be overridden with `API_ADDR` or a single argument:

```bash
go run ./cmd/api 127.0.0.1:18081
```

Ctrl+C / SIGTERM shuts the process down. `http.ErrServerClosed` is not treated as a startup failure.

## Layout and dependency direction

```
cmd/api          composition root: constructs MemoryStore and wires server.New(store)
internal/server  routing only
internal/bookmark  resource, validation, Store interface, MemoryStore, HTTP handlers
internal/httpx   JSON decode/encode and the flat problem body
```

Dependency direction: `cmd/api` → `server` → `bookmark` → `httpx`.

Handlers depend on the `Store` interface, not `*MemoryStore`. There is no package-level bookmark map; each store instance owns its own `RWMutex` + map.

`notes`, `api`, `calc`, and the existing CLI packages are unchanged. The notes server still lives behind `cli-calculator serve`.

## Endpoints

| Method | Path | Status | Notes |
|--------|------|--------|--------|
| GET | `/health` | 200 | liveness |
| GET | `/bookmarks` | 200 | list (`[]` when empty), ID order |
| POST | `/bookmarks` | 201 | create; `Location: /bookmarks/{id}` |
| GET | `/bookmarks/{id}` | 200 | malformed UUID → 400, missing → 404 |
| PUT | `/bookmarks/{id}` | 200 | full replacement |
| PATCH | `/bookmarks/{id}` | 200 | partial update |
| DELETE | `/bookmarks/{id}` | 204 | empty body |

Wrong method → 405 plus `Allow`. Unknown route → JSON 404.

## PUT vs PATCH

PUT is a full replacement. `title` and `url` are required. If `tags` is omitted, tags become `[]`.

PATCH is partial. Only sent fields change. Pointers distinguish omitted vs empty. `{}` is 400. `"tags":[]` clears tags.

## Validation

- Title required after trim; max 200 runes
- URL required after trim; max 2048 runes; absolute `http`/`https` with a host
- At most 10 tags; each tag max 32 runes after trim
- Empty or whitespace-only tags are rejected (`tag must not be empty`)
- Malformed UUID path → 400

## Duplicate URL

The stored URL is the trimmed value. A second create with the same trimmed URL is **409 Conflict**. Updating another record onto that URL is also 409. Updating a record with its own current URL is allowed.

## Errors

Every error body uses the same flat shape:

```json
{"code":"validation_error","message":"title is required"}
```

| Status | Code | When |
|--------|------|------|
| 400 | `invalid_json` | empty, malformed, unknown-field, or trailing JSON |
| 400 | `validation_error` | field rules or malformed UUID |
| 404 | `not_found` | missing bookmark or unknown route |
| 405 | `method_not_allowed` | plus `Allow` |
| 409 | `conflict` | duplicate normalized URL |
| 415 | `unsupported_media_type` | body requests that are not `application/json` (charset=utf-8 is fine) |
| 500 | `internal_error` | unexpected store failure; the raw error is not returned |

JSON decode uses `json.NewDecoder`, a 1 MiB body limit, and `DisallowUnknownFields`.

## Out of scope

In-memory data is lost on restart. Auth and a database are not part of this MVP.

## Windows curl.exe examples

```bash
curl.exe -sS http://127.0.0.1:8081/health
curl.exe -sS http://127.0.0.1:8081/bookmarks

curl.exe -sS -D - -X POST http://127.0.0.1:8081/bookmarks -H "Content-Type: application/json" --data-binary "{\"title\":\"Go\",\"url\":\"https://go.dev\",\"tags\":[\"lang\"]}"

curl.exe -sS http://127.0.0.1:8081/bookmarks/<id>

curl.exe -sS -X POST http://127.0.0.1:8081/bookmarks -H "Content-Type: application/json" --data-binary "{\"title\":\"Also Go\",\"url\":\"https://go.dev\"}"

curl.exe -sS -X PUT http://127.0.0.1:8081/bookmarks/<id> -H "Content-Type: application/json" --data-binary "{\"title\":\"Tour\",\"url\":\"https://go.dev/tour\"}"

curl.exe -sS -X PATCH http://127.0.0.1:8081/bookmarks/<id> -H "Content-Type: application/json" --data-binary "{\"title\":\"Blog\"}"

curl.exe -sS -X PATCH http://127.0.0.1:8081/bookmarks/<id> -H "Content-Type: application/json" --data-binary "{\"tags\":[]}"

curl.exe -sS -X DELETE http://127.0.0.1:8081/bookmarks/<id>
```

The first POST is 201. The second POST with the same URL is 409. PUT without `tags` clears them. PATCH with only `title` leaves `url` alone.

## Learning notes

What worked: keeping `cmd/api` as the composition root and pointing handlers at a `Store` interface made HTTP tests inject a failing store for 500 without touching persistence. `RWMutex` plus deep-copied `Tags` slices stopped callers from mutating internal state.

What was hard: giving PUT and PATCH different JSON shapes (value vs pointer) while still rejecting empty PATCH and treating omitted PUT tags as a replacement. Mapping one `ErrConflict` sentinel through a single HTTP helper kept 409 consistent without leaking store text.

Next: persistence and auth stay out of this MVP on purpose.
