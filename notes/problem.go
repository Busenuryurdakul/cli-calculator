package notes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const maxJSONBody = 1 << 20

// Problem is a consistent JSON error body.
type Problem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func decodeJSON(r io.Reader, dst any) error {
	if r == nil {
		return fmt.Errorf("decode json: empty body")
	}
	limited := io.LimitReader(r, maxJSONBody+1)
	dec := json.NewDecoder(limited)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("decode json: empty body")
		}
		return fmt.Errorf("decode json: %w", err)
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("decode json: trailing data")
		}
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeProblem(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, Problem{Code: code, Message: message})
}

func writeMethodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	writeProblem(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}

func requireMethod(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			writeMethodNotAllowed(w, method)
			return
		}
		next(w, r)
	}
}

func HandleNotFound(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, http.StatusNotFound, "not_found", "not found")
}
