package notes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestHealth(t *testing.T) {
	rec := serve(t, NewHandler(NewStore(), io.Discard), newRequest(t, http.MethodGet, "/health", "", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	assertJSONContentType(t, rec)
	var hr HealthResponse
	decodeJSONBody(t, rec.Body, &hr)
	if hr.Status != "ok" {
		t.Fatalf("status=%q", hr.Status)
	}
}

func TestInspectRedactsSecrets(t *testing.T) {
	rec := serve(t, NewHandler(NewStore(), io.Discard), newRequest(t, http.MethodGet, "/inspect", "", map[string]string{
		"Authorization": "Bearer super-secret",
		"Cookie":        "session=abc",
		"User-Agent":    "notes-test",
	}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "method=GET") || !strings.Contains(body, "path=/inspect") {
		t.Fatalf("body=%q", body)
	}
	if strings.Contains(body, "super-secret") || strings.Contains(body, "session=abc") {
		t.Fatalf("secret leaked: %q", body)
	}
	if !strings.Contains(body, "Authorization=[redacted]") || !strings.Contains(body, "Cookie=[redacted]") {
		t.Fatalf("missing redaction: %q", body)
	}
}

func TestUsersPathParam(t *testing.T) {
	h := NewHandler(NewStore(), io.Discard)
	ok := serve(t, h, newRequest(t, http.MethodGet, "/users/1", "", nil))
	if ok.Code != http.StatusOK {
		t.Fatalf("status=%d", ok.Code)
	}
	var user User
	decodeJSONBody(t, ok.Body, &user)
	if user.ID != "1" || user.Name != "Ayse" {
		t.Fatalf("user=%+v", user)
	}
	assertProblem(t, serve(t, h, newRequest(t, http.MethodGet, "/users/99", "", nil)), http.StatusNotFound, "not_found")
}

func TestEmptyNoteList(t *testing.T) {
	rec := serve(t, NewHandler(NewStore(), io.Discard), notesRequest(t, http.MethodGet, "/notes", "", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	var notes []Note
	decodeJSONBody(t, rec.Body, &notes)
	if notes == nil || len(notes) != 0 {
		t.Fatalf("list=%#v", notes)
	}
}

func TestNotesCRUD(t *testing.T) {
	h := NewHandler(NewStore(), io.Discard)

	created := serve(t, h, notesRequest(t, http.MethodPost, "/notes", `{"title":"Buy milk","body":"2L"}`, jsonHeaders()))
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	assertJSONContentType(t, created)
	if created.Header().Get("Location") != "/notes/1" {
		t.Fatalf("location=%q", created.Header().Get("Location"))
	}
	var note Note
	decodeJSONBody(t, created.Body, &note)
	if note.ID != "1" || note.Title != "Buy milk" {
		t.Fatalf("note=%+v", note)
	}

	listed := serve(t, h, notesRequest(t, http.MethodGet, "/notes", "", nil))
	var notes []Note
	decodeJSONBody(t, listed.Body, &notes)
	if listed.Code != http.StatusOK || len(notes) != 1 {
		t.Fatalf("list status=%d notes=%+v", listed.Code, notes)
	}

	got := serve(t, h, notesRequest(t, http.MethodGet, "/notes/1", "", nil))
	if got.Code != http.StatusOK {
		t.Fatalf("get status=%d", got.Code)
	}

	del := serve(t, h, notesRequest(t, http.MethodDelete, "/notes/1", "", nil))
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d", del.Code)
	}
	if del.Body.Len() != 0 {
		t.Fatalf("delete body=%q", del.Body.String())
	}

	assertProblem(t, serve(t, h, notesRequest(t, http.MethodGet, "/notes/1", "", nil)), http.StatusNotFound, "not_found")
}

func TestNotesErrorPaths(t *testing.T) {
	h := NewHandler(NewStore(), io.Discard)
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		headers    map[string]string
		wantStatus int
		wantCode   string
		wantAllow  string
	}{
		{name: "empty body", method: http.MethodPost, path: "/notes", body: "", wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "malformed json", method: http.MethodPost, path: "/notes", body: `{"title":`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "unknown field", method: http.MethodPost, path: "/notes", body: `{"title":"x","nope":true}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "trailing json", method: http.MethodPost, path: "/notes", body: `{"title":"x"}{}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "missing title", method: http.MethodPost, path: "/notes", body: `{"title":"  ","body":"x"}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_request"},
		{name: "missing note", method: http.MethodDelete, path: "/notes/9", wantStatus: http.StatusNotFound, wantCode: "not_found"},
		{name: "invalid id", method: http.MethodDelete, path: "/notes/abc", wantStatus: http.StatusBadRequest, wantCode: "invalid_id"},
		{name: "wrong collection method", method: http.MethodPut, path: "/notes", wantStatus: http.StatusMethodNotAllowed, wantCode: "method_not_allowed", wantAllow: "GET, POST"},
		{name: "wrong item method", method: http.MethodPost, path: "/notes/1", wantStatus: http.StatusMethodNotAllowed, wantCode: "method_not_allowed", wantAllow: "GET, DELETE"},
		{name: "unknown route", method: http.MethodGet, path: "/missing", wantStatus: http.StatusNotFound, wantCode: "not_found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := jsonHeaders()
			for k, v := range tt.headers {
				headers[k] = v
			}
			rec := serve(t, h, notesRequest(t, tt.method, tt.path, tt.body, headers))
			assertProblem(t, rec, tt.wantStatus, tt.wantCode)
			if tt.wantAllow != "" && rec.Header().Get("Allow") != tt.wantAllow {
				t.Fatalf("Allow=%q, want %q", rec.Header().Get("Allow"), tt.wantAllow)
			}
		})
	}
}

func TestAuthRequiredOnNotes(t *testing.T) {
	h := NewHandler(NewStore(), io.Discard)

	missing := serve(t, h, newRequest(t, http.MethodGet, "/notes", "", nil))
	assertProblem(t, missing, http.StatusUnauthorized, "unauthorized")

	bad := serve(t, h, newRequest(t, http.MethodGet, "/notes", "", map[string]string{
		"Authorization": "Bearer nope",
	}))
	assertProblem(t, bad, http.StatusUnauthorized, "unauthorized")

	ok := serve(t, h, notesRequest(t, http.MethodGet, "/notes", "", nil))
	if ok.Code != http.StatusOK {
		t.Fatalf("demo token status=%d", ok.Code)
	}

	health := serve(t, h, newRequest(t, http.MethodGet, "/health", "", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health should not require notes auth, got %d", health.Code)
	}
}

func TestRequestID(t *testing.T) {
	var logs bytes.Buffer
	h := NewHandler(NewStore(), &logs)
	rec := serve(t, h, newRequest(t, http.MethodGet, "/health", "", map[string]string{
		"X-Request-ID": "fixed-id",
	}))
	if rec.Header().Get("X-Request-ID") != "fixed-id" {
		t.Fatalf("request id=%q", rec.Header().Get("X-Request-ID"))
	}
	if !strings.Contains(logs.String(), "GET /health 200") || !strings.Contains(logs.String(), "req=fixed-id") {
		t.Fatalf("log=%q", logs.String())
	}
}

func TestRecoverDoesNotLeakAndLogs500(t *testing.T) {
	var logs bytes.Buffer
	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("do-not-leak")
	}), RequestID, RequestLog(&logs), Recover, NotesAuth)

	req := newRequest(t, http.MethodGet, "/boom", "", map[string]string{"X-Request-ID": "panic-id"})
	rec := serve(t, h, req)
	assertProblem(t, rec, http.StatusInternalServerError, "internal_error")
	if strings.Contains(rec.Body.String(), "do-not-leak") {
		t.Fatalf("panic leaked: %s", rec.Body.String())
	}
	if rec.Header().Get("X-Request-ID") != "panic-id" {
		t.Fatalf("request id=%q", rec.Header().Get("X-Request-ID"))
	}
	if !strings.Contains(logs.String(), "GET /boom 500") {
		t.Fatalf("logger missed 500: %q", logs.String())
	}
}

func TestConcurrentCreatesUniqueIDs(t *testing.T) {
	h := NewHandler(NewStore(), io.Discard)
	const n = 40
	ids := make(chan string, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			body := `{"title":"n` + strconv.Itoa(i) + `"}`
			rec := serve(t, h, notesRequest(t, http.MethodPost, "/notes", body, jsonHeaders()))
			if rec.Code != http.StatusCreated {
				t.Errorf("status=%d body=%s", rec.Code, rec.Body.String())
				return
			}
			var note Note
			decodeJSONBody(t, rec.Body, &note)
			ids <- note.ID
		}(i)
	}
	wg.Wait()
	close(ids)

	seen := map[string]bool{}
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate id %q", id)
		}
		seen[id] = true
	}
	if len(seen) != n {
		t.Fatalf("got %d unique ids, want %d", len(seen), n)
	}
}

func TestHealthWrongMethodAllow(t *testing.T) {
	rec := serve(t, NewHandler(NewStore(), io.Discard), newRequest(t, http.MethodPost, "/health", "", nil))
	assertProblem(t, rec, http.StatusMethodNotAllowed, "method_not_allowed")
	if rec.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("Allow=%q", rec.Header().Get("Allow"))
	}
}

func newRequest(t *testing.T, method, path, body string, headers map[string]string) *http.Request {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req
}

func notesRequest(t *testing.T, method, path, body string, headers map[string]string) *http.Request {
	t.Helper()
	if headers == nil {
		headers = map[string]string{}
	}
	headers["Authorization"] = "Bearer demo"
	return newRequest(t, method, path, body, headers)
}

func serve(t *testing.T, h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func jsonHeaders() map[string]string {
	return map[string]string{"Content-Type": "application/json"}
}

func decodeJSONBody(t *testing.T, r io.Reader, dst any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(dst); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func assertJSONContentType(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	got := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(got, "application/json") {
		t.Fatalf("Content-Type=%q, want application/json", got)
	}
}

func assertProblem(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("got status %d, want %d body=%s", rec.Code, status, rec.Body.String())
	}
	assertJSONContentType(t, rec)
	var p Problem
	decodeJSONBody(t, rec.Body, &p)
	if p.Code != code {
		t.Fatalf("got code %q, want %q", p.Code, code)
	}
	if p.Message == "" {
		t.Fatal("expected problem message")
	}
}
