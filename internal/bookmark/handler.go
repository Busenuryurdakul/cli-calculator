package bookmark

import (
	"errors"
	"net/http"

	"github.com/Busenuryurdakul/cli-calculator/internal/httpx"
)

// Handler serves /bookmarks using a Store. It decodes JSON and maps
// domain errors; validation and persistence live in the store.
type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		httpx.WriteMethodNotAllowed(w, "GET, POST")
	}
}

func (h *Handler) Item(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodPut:
		h.update(w, r)
	case http.MethodPatch:
		h.patch(w, r)
	case http.MethodDelete:
		h.delete(w, r)
	default:
		httpx.WriteMethodNotAllowed(w, "GET, PUT, PATCH, DELETE")
	}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if !readJSON(w, r, &req) {
		return
	}
	b, err := h.store.Create(req.Title, req.URL, req.Tags)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Location", "/bookmarks/"+b.ID)
	httpx.WriteJSON(w, http.StatusCreated, b)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.store.List())
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	b, err := h.store.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, b)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var req UpdateRequest
	if !readJSON(w, r, &req) {
		return
	}
	b, err := h.store.Update(r.PathValue("id"), req.Title, req.URL, req.Tags)
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, b)
}

func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	var req PatchRequest
	if !readJSON(w, r, &req) {
		return
	}
	b, err := h.store.Patch(r.PathValue("id"), req.Title, req.URL, req.Tags)
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, b)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	err := httpx.DecodeJSONRequest(r, dst)
	if err == nil {
		return true
	}
	switch {
	case errors.Is(err, httpx.ErrUnsupportedMediaType):
		httpx.WriteProblem(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "content type must be application/json")
	case errors.Is(err, httpx.ErrEmptyBody):
		httpx.WriteProblem(w, http.StatusBadRequest, "invalid_json", "empty body")
	case errors.Is(err, httpx.ErrTrailingJSON):
		httpx.WriteProblem(w, http.StatusBadRequest, "invalid_json", "trailing json")
	default:
		httpx.WriteProblem(w, http.StatusBadRequest, "invalid_json", "invalid json")
	}
	return false
}

func writeError(w http.ResponseWriter, err error) {
	var v *ValidationError
	switch {
	case errors.As(err, &v):
		httpx.WriteProblem(w, http.StatusBadRequest, "validation_error", v.Msg)
	case errors.Is(err, ErrInvalidID):
		httpx.WriteProblem(w, http.StatusBadRequest, "validation_error", "invalid id")
	case errors.Is(err, ErrNotFound):
		httpx.WriteProblem(w, http.StatusNotFound, "not_found", "bookmark not found")
	case errors.Is(err, ErrConflict):
		httpx.WriteProblem(w, http.StatusConflict, "conflict", "bookmark with this url already exists")
	default:
		httpx.WriteProblem(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
