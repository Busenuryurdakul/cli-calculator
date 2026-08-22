package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Busenuryurdakul/cli-calculator/internal/bookmark"
	"github.com/Busenuryurdakul/cli-calculator/internal/httpx"
)

func TestHealth(t *testing.T) {
	rec := do(t, New(bookmark.NewMemoryStore()), http.MethodGet, "/health", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestBookmarksCRUD(t *testing.T) {
	h := New(bookmark.NewMemoryStore())

	empty := do(t, h, http.MethodGet, "/bookmarks", "")
	if empty.Code != http.StatusOK {
		t.Fatalf("empty list status=%d", empty.Code)
	}
	var listed []bookmark.Bookmark
	decode(t, empty.Body, &listed)
	if listed == nil || len(listed) != 0 {
		t.Fatalf("listed=%#v", listed)
	}

	created := do(t, h, http.MethodPost, "/bookmarks", `{"title":"Go","url":"https://go.dev","tags":["lang"]}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	var b bookmark.Bookmark
	decode(t, created.Body, &b)
	if b.ID == "" || b.Title != "Go" || len(b.Tags) != 1 {
		t.Fatalf("bookmark=%+v", b)
	}
	if created.Header().Get("Location") != "/bookmarks/"+b.ID {
		t.Fatalf("location=%q", created.Header().Get("Location"))
	}

	got := do(t, h, http.MethodGet, "/bookmarks/"+b.ID, "")
	if got.Code != http.StatusOK {
		t.Fatalf("get status=%d", got.Code)
	}

	put := do(t, h, http.MethodPut, "/bookmarks/"+b.ID, `{"title":"Tour","url":"https://go.dev/tour"}`)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}
	var replaced bookmark.Bookmark
	decode(t, put.Body, &replaced)
	if replaced.Title != "Tour" || replaced.URL != "https://go.dev/tour" || len(replaced.Tags) != 0 {
		t.Fatalf("put replacement=%+v", replaced)
	}

	putMissing := do(t, h, http.MethodPut, "/bookmarks/"+b.ID, `{"url":"https://go.dev"}`)
	assertProblem(t, putMissing, http.StatusBadRequest, "validation_error")

	charset := doCT(t, h, http.MethodPatch, "/bookmarks/"+b.ID, `{"title":"Blog"}`, "application/json; charset=utf-8")
	if charset.Code != http.StatusOK {
		t.Fatalf("patch charset status=%d body=%s", charset.Code, charset.Body.String())
	}
	var patched bookmark.Bookmark
	decode(t, charset.Body, &patched)
	if patched.Title != "Blog" || patched.URL != "https://go.dev/tour" {
		t.Fatalf("patched=%+v", patched)
	}

	emptyPatch := do(t, h, http.MethodPatch, "/bookmarks/"+b.ID, `{}`)
	assertProblem(t, emptyPatch, http.StatusBadRequest, "validation_error")

	seedTags := do(t, h, http.MethodPatch, "/bookmarks/"+b.ID, `{"tags":["keep"]}`)
	if seedTags.Code != http.StatusOK {
		t.Fatalf("seed tags status=%d", seedTags.Code)
	}
	clearTags := do(t, h, http.MethodPatch, "/bookmarks/"+b.ID, `{"tags":[]}`)
	if clearTags.Code != http.StatusOK {
		t.Fatalf("clear tags status=%d body=%s", clearTags.Code, clearTags.Body.String())
	}
	var cleared bookmark.Bookmark
	decode(t, clearTags.Body, &cleared)
	if len(cleared.Tags) != 0 {
		t.Fatalf("cleared=%+v", cleared)
	}

	del := do(t, h, http.MethodDelete, "/bookmarks/"+b.ID, "")
	if del.Code != http.StatusNoContent || del.Body.Len() != 0 {
		t.Fatalf("delete status=%d body=%q", del.Code, del.Body.String())
	}
	repeat := do(t, h, http.MethodDelete, "/bookmarks/"+b.ID, "")
	assertProblem(t, repeat, http.StatusNotFound, "not_found")

	missing := do(t, h, http.MethodGet, "/bookmarks/"+b.ID, "")
	assertProblem(t, missing, http.StatusNotFound, "not_found")
}

func TestDuplicateURLConflict(t *testing.T) {
	h := New(bookmark.NewMemoryStore())
	first := do(t, h, http.MethodPost, "/bookmarks", `{"title":"A","url":"https://example.com/x"}`)
	if first.Code != http.StatusCreated {
		t.Fatalf("first=%d %s", first.Code, first.Body.String())
	}
	dup := do(t, h, http.MethodPost, "/bookmarks", `{"title":"B","url":"  https://example.com/x  "}`)
	assertProblem(t, dup, http.StatusConflict, "conflict")

	var a bookmark.Bookmark
	decode(t, first.Body, &a)
	second := do(t, h, http.MethodPost, "/bookmarks", `{"title":"C","url":"https://example.com/y"}`)
	if second.Code != http.StatusCreated {
		t.Fatalf("second=%d %s", second.Code, second.Body.String())
	}
	var c bookmark.Bookmark
	decode(t, second.Body, &c)

	clash := do(t, h, http.MethodPut, "/bookmarks/"+c.ID, `{"title":"C","url":"https://example.com/x"}`)
	assertProblem(t, clash, http.StatusConflict, "conflict")

	own := do(t, h, http.MethodPut, "/bookmarks/"+a.ID, `{"title":"A2","url":"https://example.com/x"}`)
	if own.Code != http.StatusOK {
		t.Fatalf("own url update=%d %s", own.Code, own.Body.String())
	}
}

func TestUnhappyPaths(t *testing.T) {
	h := New(bookmark.NewMemoryStore())
	created := do(t, h, http.MethodPost, "/bookmarks", `{"title":"Go","url":"https://go.dev"}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("seed status=%d", created.Code)
	}
	var b bookmark.Bookmark
	decode(t, created.Body, &b)

	tooMany := make([]string, bookmark.MaxTags+1)
	for i := range tooMany {
		tooMany[i] = "t"
	}
	tooManyJSON := `{"title":"x","url":"https://go.dev/extra","tags":["` + strings.Join(tooMany, `","`) + `"]}`

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		ct         string
		wantStatus int
		wantCode   string
		wantAllow  string
	}{
		{name: "empty body", method: http.MethodPost, path: "/bookmarks", body: "", ct: "application/json", wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "malformed json", method: http.MethodPost, path: "/bookmarks", body: `{"title":`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "unknown field", method: http.MethodPost, path: "/bookmarks", body: `{"title":"x","url":"https://go.dev","nope":1}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "trailing json", method: http.MethodPost, path: "/bookmarks", body: `{"title":"x","url":"https://go.dev"}{}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "empty title", method: http.MethodPost, path: "/bookmarks", body: `{"title":"  ","url":"https://go.dev"}`, wantStatus: http.StatusBadRequest, wantCode: "validation_error"},
		{name: "long title", method: http.MethodPost, path: "/bookmarks", body: `{"title":"` + strings.Repeat("a", bookmark.MaxTitleLen+1) + `","url":"https://go.dev"}`, wantStatus: http.StatusBadRequest, wantCode: "validation_error"},
		{name: "bad url", method: http.MethodPost, path: "/bookmarks", body: `{"title":"x","url":"not-a-url"}`, wantStatus: http.StatusBadRequest, wantCode: "validation_error"},
		{name: "long url", method: http.MethodPost, path: "/bookmarks", body: `{"title":"x","url":"https://go.dev/` + strings.Repeat("a", bookmark.MaxURLLen) + `"}`, wantStatus: http.StatusBadRequest, wantCode: "validation_error"},
		{name: "empty tag", method: http.MethodPost, path: "/bookmarks", body: `{"title":"x","url":"https://go.dev/t","tags":[" "]}`, wantStatus: http.StatusBadRequest, wantCode: "validation_error"},
		{name: "too many tags", method: http.MethodPost, path: "/bookmarks", body: tooManyJSON, wantStatus: http.StatusBadRequest, wantCode: "validation_error"},
		{name: "long tag", method: http.MethodPost, path: "/bookmarks", body: `{"title":"x","url":"https://go.dev/tag","tags":["` + strings.Repeat("a", bookmark.MaxTagLen+1) + `"]}`, wantStatus: http.StatusBadRequest, wantCode: "validation_error"},
		{name: "invalid id", method: http.MethodGet, path: "/bookmarks/abc", wantStatus: http.StatusBadRequest, wantCode: "validation_error"},
		{name: "missing id", method: http.MethodGet, path: "/bookmarks/22222222-2222-4222-8222-222222222222", wantStatus: http.StatusNotFound, wantCode: "not_found"},
		{name: "wrong collection method", method: http.MethodDelete, path: "/bookmarks", wantStatus: http.StatusMethodNotAllowed, wantCode: "method_not_allowed", wantAllow: "GET, POST"},
		{name: "wrong item method", method: http.MethodPost, path: "/bookmarks/" + b.ID, wantStatus: http.StatusMethodNotAllowed, wantCode: "method_not_allowed", wantAllow: "GET, PUT, PATCH, DELETE"},
		{name: "unknown route", method: http.MethodGet, path: "/nope", wantStatus: http.StatusNotFound, wantCode: "not_found"},
		{name: "patch no fields", method: http.MethodPatch, path: "/bookmarks/" + b.ID, body: `{}`, wantStatus: http.StatusBadRequest, wantCode: "validation_error"},
		{name: "unsupported content type", method: http.MethodPost, path: "/bookmarks", body: `{"title":"x","url":"https://go.dev"}`, ct: "text/plain", wantStatus: http.StatusUnsupportedMediaType, wantCode: "unsupported_media_type"},
		{name: "missing content type", method: http.MethodPost, path: "/bookmarks", body: `{"title":"x","url":"https://go.dev"}`, ct: "omit", wantStatus: http.StatusUnsupportedMediaType, wantCode: "unsupported_media_type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rec *httptest.ResponseRecorder
			switch tt.ct {
			case "":
				rec = do(t, h, tt.method, tt.path, tt.body)
			case "omit":
				rec = doCT(t, h, tt.method, tt.path, tt.body, "")
			default:
				rec = doCT(t, h, tt.method, tt.path, tt.body, tt.ct)
			}
			assertProblem(t, rec, tt.wantStatus, tt.wantCode)
			if tt.wantAllow != "" && rec.Header().Get("Allow") != tt.wantAllow {
				t.Fatalf("Allow=%q want %q", rec.Header().Get("Allow"), tt.wantAllow)
			}
		})
	}
}

func TestInjectedStoreFailure(t *testing.T) {
	h := New(failStore{})
	rec := do(t, h, http.MethodPost, "/bookmarks", `{"title":"Go","url":"https://go.dev"}`)
	assertProblem(t, rec, http.StatusInternalServerError, "internal_error")
	if strings.Contains(rec.Body.String(), "disk exploded") {
		t.Fatal("leaked internal error")
	}
}

type failStore struct{}

func (failStore) Create(string, string, []string) (bookmark.Bookmark, error) {
	return bookmark.Bookmark{}, errors.New("disk exploded")
}
func (failStore) List() []bookmark.Bookmark { return nil }
func (failStore) Get(string) (bookmark.Bookmark, error) {
	return bookmark.Bookmark{}, errors.New("disk exploded")
}
func (failStore) Update(string, string, string, []string) (bookmark.Bookmark, error) {
	return bookmark.Bookmark{}, errors.New("disk exploded")
}
func (failStore) Patch(string, *string, *string, *[]string) (bookmark.Bookmark, error) {
	return bookmark.Bookmark{}, errors.New("disk exploded")
}
func (failStore) Delete(string) error { return errors.New("disk exploded") }

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	ct := ""
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		ct = "application/json"
	}
	return doCT(t, h, method, path, body, ct)
}

func doCT(t *testing.T, h http.Handler, method, path, body, ct string) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decode(t *testing.T, r io.Reader, dst any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(dst); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func assertProblem(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("got status %d want %d body=%s", rec.Code, status, rec.Body.String())
	}
	if !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("content-type=%q", rec.Header().Get("Content-Type"))
	}
	var p httpx.Problem
	decode(t, rec.Body, &p)
	if p.Code != code || p.Message == "" {
		t.Fatalf("problem=%+v want code %s", p, code)
	}
}
