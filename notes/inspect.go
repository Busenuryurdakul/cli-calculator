package notes

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// HandleInspect replies with plain text request metadata.
// Authorization and Cookie values are redacted.
func HandleInspect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	var b strings.Builder
	fmt.Fprintf(&b, "method=%s\n", r.Method)
	fmt.Fprintf(&b, "path=%s\n", r.URL.Path)
	fmt.Fprintf(&b, "remote=%s\n", r.RemoteAddr)

	keys := make([]string, 0, len(r.Header))
	for k := range r.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%s\n", k, redactHeader(k, r.Header.Get(k)))
	}
	_, _ = w.Write([]byte(b.String()))
}

func redactHeader(name, value string) string {
	switch strings.ToLower(name) {
	case "authorization", "cookie", "set-cookie":
		return "[redacted]"
	default:
		return value
	}
}
