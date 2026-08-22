package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Busenuryurdakul/cli-calculator/calc"
)

// CalculateRequest is the public JSON body for POST /calculate.
type CalculateRequest struct {
	A  float64 `json:"a"`
	B  float64 `json:"b"`
	Op string  `json:"op"`
}

// CalculateResponse is the public JSON body returned by POST /calculate.
type CalculateResponse struct {
	Result *float64 `json:"result,omitempty"`
	Error  string   `json:"error,omitempty"`
}

// HealthResponse is the public JSON body returned by GET /health.
type HealthResponse struct {
	Status string `json:"status"`
}

// Handler mounts the exported HTTP API.
func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", HandleHealth)
	mux.HandleFunc("/calculate", HandleCalculate)
	return mux
}

// HandleHealth reports liveness as JSON.
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

// HandleCalculate evaluates a JSON expression through calc.Calculate.
func HandleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req CalculateRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, CalculateResponse{Error: "invalid json"})
		return
	}

	result, err := calc.Calculate(req.A, req.B, req.Op)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, CalculateResponse{Error: publicCalcError(err)})
		return
	}
	writeJSON(w, http.StatusOK, CalculateResponse{Result: floatPtr(result)})
}

func decodeJSON(r io.Reader, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r, 1<<20))
	if err := dec.Decode(dst); err != nil {
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

func publicCalcError(err error) string {
	switch {
	case errors.Is(err, calc.ErrDivisionByZero):
		return "division by zero"
	case errors.Is(err, calc.ErrInvalidOp):
		return "invalid operator"
	default:
		return "calculate failed"
	}
}

func floatPtr(v float64) *float64 {
	return &v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// TestingSummary documents the testing-week patterns in this package.
func TestingSummary() string {
	return `Testing patterns in this repo:

- _test.go files live beside the code and use the testing package
- Table-driven tests plus t.Run name each case, including zeros and errors
- Examples lock public output; benchmarks use ResetTimer and report ns/op
- HTTP handlers are tested with httptest.NewRecorder, not real ports
- t.Helper() keeps fixtures short so failures point at the table row`
}
