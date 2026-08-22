package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	rec := serve(t, HandleHealth, newRequest(t, http.MethodGet, "/health", "", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
	}
	assertJSONContentType(t, rec)
	got := decodeHealth(t, rec.Body)
	if got.Status != "ok" {
		t.Fatalf("got status %q, want ok", got.Status)
	}
}

func TestHandlers(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		headers    map[string]string
		wantStatus int
		wantResult *float64
		wantErr    string
	}{
		{
			name:       "add",
			method:     http.MethodPost,
			path:       "/calculate",
			body:       `{"a":10,"op":"+","b":5}`,
			headers:    map[string]string{"Content-Type": "application/json"},
			wantStatus: http.StatusOK,
			wantResult: floatPtr(15),
		},
		{
			name:       "zero result",
			method:     http.MethodPost,
			path:       "/calculate",
			body:       `{"a":0,"op":"+","b":0}`,
			wantStatus: http.StatusOK,
			wantResult: floatPtr(0),
		},
		{
			name:       "empty body",
			method:     http.MethodPost,
			path:       "/calculate",
			body:       "",
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid json",
		},
		{
			name:       "malformed json",
			method:     http.MethodPost,
			path:       "/calculate",
			body:       `{"a":1`,
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid json",
		},
		{
			name:       "trailing json",
			method:     http.MethodPost,
			path:       "/calculate",
			body:       `{"a":1,"op":"+","b":2}{}`,
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid json",
		},
		{
			name:       "empty operator",
			method:     http.MethodPost,
			path:       "/calculate",
			body:       `{"a":1,"op":"","b":2}`,
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid operator",
		},
		{
			name:       "invalid operator",
			method:     http.MethodPost,
			path:       "/calculate",
			body:       `{"a":1,"op":"%","b":2}`,
			wantStatus: http.StatusBadRequest,
			wantErr:    "invalid operator",
		},
		{
			name:       "division by zero",
			method:     http.MethodPost,
			path:       "/calculate",
			body:       `{"a":10,"op":"/","b":0}`,
			wantStatus: http.StatusBadRequest,
			wantErr:    "division by zero",
		},
		{
			name:       "method not allowed",
			method:     http.MethodGet,
			path:       "/calculate",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "health post rejected",
			method:     http.MethodPost,
			path:       "/health",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "unknown path",
			method:     http.MethodGet,
			path:       "/missing",
			wantStatus: http.StatusNotFound,
		},
	}

	h := Handler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newRequest(t, tt.method, tt.path, tt.body, tt.headers)
			rec := serveHandler(t, h, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("got status %d, want %d", rec.Code, tt.wantStatus)
			}
			if rec.Code == http.StatusMethodNotAllowed || rec.Code == http.StatusNotFound {
				return
			}
			assertJSONContentType(t, rec)
			got := decodeResponse(t, rec.Body)
			if tt.wantErr != "" && got.Error != tt.wantErr {
				t.Fatalf("got error %q, want %q", got.Error, tt.wantErr)
			}
			if tt.wantResult == nil {
				return
			}
			if got.Result == nil {
				t.Fatalf("got result nil, want %v", *tt.wantResult)
			}
			if *got.Result != *tt.wantResult {
				t.Fatalf("got result %v, want %v", *got.Result, *tt.wantResult)
			}
		})
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

func serve(t *testing.T, h http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func serveHandler(t *testing.T, h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeResponse(t *testing.T, r io.Reader) CalculateResponse {
	t.Helper()
	var got CalculateResponse
	if err := json.NewDecoder(r).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return got
}

func decodeHealth(t *testing.T, r io.Reader) HealthResponse {
	t.Helper()
	var got HealthResponse
	if err := json.NewDecoder(r).Decode(&got); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	return got
}

func assertJSONContentType(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	got := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(got, "application/json") {
		t.Fatalf("Content-Type=%q, want application/json", got)
	}
}
