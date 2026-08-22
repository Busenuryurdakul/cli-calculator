package notes

import "net/http"

// User is a demo resource for path-parameter routing.
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var demoUsers = map[string]string{
	"1": "Ayse",
	"2": "Mehmet",
}

// HandleGetUser serves GET /users/{id}.
func HandleGetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	name, ok := demoUsers[id]
	if !ok {
		writeProblem(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	writeJSON(w, http.StatusOK, User{ID: id, Name: name})
}
