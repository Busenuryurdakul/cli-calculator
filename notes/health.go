package notes

import "net/http"

// HealthResponse is returned by GET /health.
type HealthResponse struct {
	Status string `json:"status"`
}

// HandleHealth is a liveness probe.
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}
