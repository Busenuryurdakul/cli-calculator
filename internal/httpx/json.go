package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
)

const maxJSONBody = 1 << 20

var (
	ErrEmptyBody            = errors.New("empty body")
	ErrInvalidJSON          = errors.New("invalid json")
	ErrTrailingJSON         = errors.New("trailing json")
	ErrUnsupportedMediaType = errors.New("unsupported media type")
)

// Problem is the stable JSON error body for the REST MVP.
type Problem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func IsJSONContentType(ct string) bool {
	media, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return false
	}
	return strings.EqualFold(media, "application/json")
}

func DecodeJSON(r io.Reader, dst any) error {
	if r == nil {
		return ErrEmptyBody
	}
	dec := json.NewDecoder(io.LimitReader(r, maxJSONBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return ErrEmptyBody
		}
		return ErrInvalidJSON
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return ErrTrailingJSON
	}
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteProblem(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, Problem{Code: code, Message: message})
}

func WriteMethodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	WriteProblem(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}

func DecodeJSONRequest(r *http.Request, dst any) error {
	if !IsJSONContentType(r.Header.Get("Content-Type")) {
		return ErrUnsupportedMediaType
	}
	return DecodeJSON(r.Body, dst)
}
