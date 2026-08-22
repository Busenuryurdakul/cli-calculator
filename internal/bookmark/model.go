package bookmark

import "time"

const (
	MaxTitleLen = 200
	MaxURLLen   = 2048
	MaxTags     = 10
	MaxTagLen   = 32
)

// Bookmark is the public resource representation.
type Bookmark struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateRequest is the JSON body for POST /bookmarks.
type CreateRequest struct {
	Title string   `json:"title"`
	URL   string   `json:"url"`
	Tags  []string `json:"tags"`
}

// UpdateRequest is the JSON body for PUT /bookmarks/{id}.
// Omitted tags replace the stored list with an empty list.
type UpdateRequest struct {
	Title string   `json:"title"`
	URL   string   `json:"url"`
	Tags  []string `json:"tags"`
}

// PatchRequest is the JSON body for PATCH /bookmarks/{id}.
// Nil pointer fields are left unchanged. Tags pointing at an empty
// slice clears tags.
type PatchRequest struct {
	Title *string   `json:"title"`
	URL   *string   `json:"url"`
	Tags  *[]string `json:"tags"`
}
