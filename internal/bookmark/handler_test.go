package bookmark

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerCreateRejectsEmptyTitle(t *testing.T) {
	h := NewHandler(NewMemoryStore())
	req := httptest.NewRequest(http.MethodPost, "/bookmarks", strings.NewReader(`{"title":"","url":"https://go.dev"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Collection(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"validation_error"`) {
		t.Fatalf("body=%s", rec.Body.String())
	}
}

func TestHandlerItemRejectsBadMethod(t *testing.T) {
	h := NewHandler(NewMemoryStore())
	req := httptest.NewRequest(http.MethodPost, "/bookmarks/11111111-1111-4111-8111-111111111111", nil)
	req.SetPathValue("id", "11111111-1111-4111-8111-111111111111")
	rec := httptest.NewRecorder()
	h.Item(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d", rec.Code)
	}
	if rec.Header().Get("Allow") != "GET, PUT, PATCH, DELETE" {
		t.Fatalf("Allow=%q", rec.Header().Get("Allow"))
	}
}
