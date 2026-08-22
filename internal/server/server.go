package server

import (
	"net/http"

	"github.com/Busenuryurdakul/cli-calculator/internal/bookmark"
	"github.com/Busenuryurdakul/cli-calculator/internal/httpx"
)

type healthResponse struct {
	Status string `json:"status"`
}

// New mounts the bookmarks REST MVP. The caller (cmd/api) owns store wiring.
func New(store bookmark.Store) http.Handler {
	h := bookmark.NewHandler(store)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			httpx.WriteMethodNotAllowed(w, http.MethodGet)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, healthResponse{Status: "ok"})
	})
	mux.HandleFunc("/bookmarks", h.Collection)
	mux.HandleFunc("/bookmarks/{id}", h.Item)
	mux.HandleFunc("/{$}", notFound)
	mux.HandleFunc("/{path...}", notFound)
	return mux
}

func notFound(w http.ResponseWriter, r *http.Request) {
	httpx.WriteProblem(w, http.StatusNotFound, "not_found", "not found")
}
