# Notes API

In-memory HTTP service. Start it with:

```bash
go run . serve
# default http://127.0.0.1:8080
go run . serve 127.0.0.1:8080
```

`Authorization: Bearer demo` is required on **all `/notes` routes**. It is an educational placeholder, not real authentication. Health, inspect, and users do not use it.

| Method | Path | Status | Description |
|--------|------|--------|-------------|
| GET | `/health` | 200 | liveness JSON |
| GET | `/inspect` | 200 | plain-text request metadata (Authorization/Cookie redacted) |
| GET | `/users/{id}` | 200 | demo user (`1` or `2`) |
| GET | `/notes` | 200 | list notes (auth required) |
| POST | `/notes` | 201 | create note; `title` required (auth required) |
| GET | `/notes/{id}` | 200 | get one note (auth required) |
| DELETE | `/notes/{id}` | 204 | delete one note, empty body (auth required) |

Missing note → **404**. Invalid note ID → **400**. Wrong method → **405** with `Allow`. Unknown route → **404**. Missing/invalid notes auth → **401**.

Errors always use `{"code":"...","message":"..."}`.

```bash
curl.exe -sS http://127.0.0.1:8080/health
curl.exe -sS http://127.0.0.1:8080/inspect
curl.exe -sS http://127.0.0.1:8080/users/1
curl.exe -sS http://127.0.0.1:8080/notes
# 401 without Authorization
curl.exe -sS -X POST http://127.0.0.1:8080/notes \
  -H "Authorization: Bearer demo" \
  -H "Content-Type: application/json" \
  --data-binary "{\"title\":\"Buy milk\",\"body\":\"2L\"}"
curl.exe -sS http://127.0.0.1:8080/notes -H "Authorization: Bearer demo"
curl.exe -sS http://127.0.0.1:8080/notes/1 -H "Authorization: Bearer demo"
curl.exe -sS -X DELETE http://127.0.0.1:8080/notes/1 -H "Authorization: Bearer demo"
```
