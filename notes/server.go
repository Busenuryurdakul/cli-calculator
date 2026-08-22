package notes

import (
	"fmt"
	"io"
	"net/http"
)

// NewHandler mounts routes. Middleware order is
// Request ID → Logging → Recovery → Auth → Handler.
func NewHandler(store *Store, logs io.Writer) http.Handler {
	if store == nil {
		store = NewStore()
	}
	notes := &NoteHandlers{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", requireMethod(http.MethodGet, HandleHealth))
	mux.HandleFunc("/inspect", requireMethod(http.MethodGet, HandleInspect))
	mux.HandleFunc("/users/{id}", requireMethod(http.MethodGet, HandleGetUser))
	mux.HandleFunc("/notes", notes.collection)
	mux.HandleFunc("/notes/{id}", notes.item)
	mux.HandleFunc("/{$}", HandleNotFound)
	mux.HandleFunc("/{path...}", HandleNotFound)

	return Chain(mux,
		RequestID,
		RequestLog(logs),
		Recover,
		NotesAuth,
	)
}

// ListenAndServe starts the standard library HTTP server.
func ListenAndServe(addr string, h http.Handler) error {
	if err := http.ListenAndServe(addr, h); err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	return nil
}

// ServerSummary documents the HTTP practice block.
func ServerSummary() string {
	return `HTTP server patterns in this package:

- ListenAndServe plus ServeMux routes (/health, /users/{id}, /notes)
- Plain-text inspect handler redacts Authorization and Cookie
- JSON decode with size limit, DisallowUnknownFields, and problem details
- Middleware: request ID → logging → recovery → notes auth placeholder
- Per-handler in-memory store protected by RWMutex`
}
