package notes

import (
	"net/http"
	"strconv"
	"strings"
)

// CreateNoteRequest is the JSON body for POST /notes.
type CreateNoteRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// NoteHandlers serves the notes resource from an injected in-memory store.
type NoteHandlers struct {
	store *Store
}

func (h *NoteHandlers) collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.List(w, r)
	case http.MethodPost:
		h.Create(w, r)
	default:
		writeMethodNotAllowed(w, "GET, POST")
	}
}

func (h *NoteHandlers) item(w http.ResponseWriter, r *http.Request) {
	if !validNoteID(r.PathValue("id")) {
		writeProblem(w, http.StatusBadRequest, "invalid_id", "invalid id")
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.Get(w, r)
	case http.MethodDelete:
		h.Delete(w, r)
	default:
		writeMethodNotAllowed(w, "GET, DELETE")
	}
}

func (h *NoteHandlers) List(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.List())
}

func (h *NoteHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateNoteRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeProblem(w, http.StatusBadRequest, "invalid_request", "title is required")
		return
	}
	note := h.store.Create(title, req.Body)
	w.Header().Set("Location", "/notes/"+note.ID)
	writeJSON(w, http.StatusCreated, note)
}

func (h *NoteHandlers) Get(w http.ResponseWriter, r *http.Request) {
	note, ok := h.store.Get(r.PathValue("id"))
	if !ok {
		writeProblem(w, http.StatusNotFound, "not_found", "note not found")
		return
	}
	writeJSON(w, http.StatusOK, note)
}

func (h *NoteHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	if !h.store.Delete(r.PathValue("id")) {
		writeProblem(w, http.StatusNotFound, "not_found", "note not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validNoteID(raw string) bool {
	n, err := strconv.Atoi(raw)
	return err == nil && n >= 1 && strconv.Itoa(n) == raw
}
