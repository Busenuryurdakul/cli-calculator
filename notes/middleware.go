package notes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type contextKey int

const requestIDKey contextKey = 1

var requestSeq atomic.Uint64

// Chain wraps h with middleware. The first wrapper is outermost.
func Chain(h http.Handler, wrappers ...func(http.Handler) http.Handler) http.Handler {
	for i := len(wrappers) - 1; i >= 0; i-- {
		h = wrappers[i](h)
	}
	return h
}

func requestIDFrom(r *http.Request) string {
	if v, ok := r.Context().Value(requestIDKey).(string); ok {
		return v
	}
	return strings.TrimSpace(r.Header.Get("X-Request-ID"))
}

// RequestID assigns X-Request-ID before later middleware runs.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if id == "" {
			id = fmt.Sprintf("req-%d", requestSeq.Add(1))
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Recover turns handler panics into a 500 JSON problem without leaking details.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				writeProblem(w, http.StatusInternalServerError, "internal_error", "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// NotesAuth requires Authorization: Bearer demo on /notes routes only.
// This is an educational placeholder, not real authentication.
func NotesAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/notes") {
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer demo" {
			writeProblem(w, http.StatusUnauthorized, "unauthorized", "invalid authorization")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequestLog records method, path, status, duration, remote address, and request ID.
func RequestLog(out io.Writer) func(http.Handler) http.Handler {
	if out == nil {
		out = io.Discard
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(sw, r)
			fmt.Fprintf(out, "%s %s %d %s remote=%s req=%s\n",
				r.Method, r.URL.Path, sw.status, time.Since(start), r.RemoteAddr, requestIDFrom(r))
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(p)
}
