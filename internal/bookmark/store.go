package bookmark

import (
	"crypto/rand"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Store is the repository interface handlers depend on.
type Store interface {
	Create(title, rawURL string, tags []string) (Bookmark, error)
	List() []Bookmark
	Get(id string) (Bookmark, error)
	Update(id, title, rawURL string, tags []string) (Bookmark, error)
	Patch(id string, title, rawURL *string, tags *[]string) (Bookmark, error)
	Delete(id string) error
}

// MemoryStore is an in-memory Store guarded by RWMutex.
// Each instance owns its own map; there is no package-level store.
type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]Bookmark
	now   func() time.Time
	newID func() (string, error)
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		items: make(map[string]Bookmark),
		now:   time.Now,
		newID: newUUID,
	}
}

func (s *MemoryStore) Create(title, rawURL string, tags []string) (Bookmark, error) {
	title, err := ValidateTitle(title)
	if err != nil {
		return Bookmark{}, err
	}
	rawURL, err = ValidateURL(rawURL)
	if err != nil {
		return Bookmark{}, err
	}
	tags, err = ValidateTags(tags)
	if err != nil {
		return Bookmark{}, err
	}
	now := s.now().UTC()
	id, err := s.newID()
	if err != nil {
		return Bookmark{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.urlTakenLocked(rawURL, "") {
		return Bookmark{}, ErrConflict
	}
	b := Bookmark{
		ID:        id,
		Title:     title,
		URL:       rawURL,
		Tags:      cloneTags(tags),
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.items[b.ID] = b
	return cloneBookmark(b), nil
}

func (s *MemoryStore) List() []Bookmark {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.items))
	for id := range s.items {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]Bookmark, 0, len(ids))
	for _, id := range ids {
		out = append(out, cloneBookmark(s.items[id]))
	}
	return out
}

func (s *MemoryStore) Get(id string) (Bookmark, error) {
	if _, err := ParseID(id); err != nil {
		return Bookmark{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.items[id]
	if !ok {
		return Bookmark{}, ErrNotFound
	}
	return cloneBookmark(b), nil
}

func (s *MemoryStore) Update(id, title, rawURL string, tags []string) (Bookmark, error) {
	if _, err := ParseID(id); err != nil {
		return Bookmark{}, err
	}
	title, err := ValidateTitle(title)
	if err != nil {
		return Bookmark{}, err
	}
	rawURL, err = ValidateURL(rawURL)
	if err != nil {
		return Bookmark{}, err
	}
	tags, err = ValidateTags(tags)
	if err != nil {
		return Bookmark{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.items[id]
	if !ok {
		return Bookmark{}, ErrNotFound
	}
	if s.urlTakenLocked(rawURL, id) {
		return Bookmark{}, ErrConflict
	}
	cur.Title = title
	cur.URL = rawURL
	cur.Tags = cloneTags(tags)
	cur.UpdatedAt = s.now().UTC()
	s.items[id] = cur
	return cloneBookmark(cur), nil
}

func (s *MemoryStore) Patch(id string, title, rawURL *string, tags *[]string) (Bookmark, error) {
	if _, err := ParseID(id); err != nil {
		return Bookmark{}, err
	}
	if title == nil && rawURL == nil && tags == nil {
		return Bookmark{}, &ValidationError{Msg: "at least one field is required"}
	}

	var nextTitle, nextURL *string
	var nextTags *[]string
	if title != nil {
		v, err := ValidateTitle(*title)
		if err != nil {
			return Bookmark{}, err
		}
		nextTitle = &v
	}
	if rawURL != nil {
		v, err := ValidateURL(*rawURL)
		if err != nil {
			return Bookmark{}, err
		}
		nextURL = &v
	}
	if tags != nil {
		v, err := ValidateTags(*tags)
		if err != nil {
			return Bookmark{}, err
		}
		nextTags = &v
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.items[id]
	if !ok {
		return Bookmark{}, ErrNotFound
	}
	if nextURL != nil && s.urlTakenLocked(*nextURL, id) {
		return Bookmark{}, ErrConflict
	}
	if nextTitle != nil {
		cur.Title = *nextTitle
	}
	if nextURL != nil {
		cur.URL = *nextURL
	}
	if nextTags != nil {
		cur.Tags = cloneTags(*nextTags)
	}
	cur.UpdatedAt = s.now().UTC()
	s.items[id] = cur
	return cloneBookmark(cur), nil
}

func (s *MemoryStore) Delete(id string) error {
	if _, err := ParseID(id); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return ErrNotFound
	}
	delete(s.items, id)
	return nil
}

func (s *MemoryStore) urlTakenLocked(normalized string, exceptID string) bool {
	for id, b := range s.items {
		if id != exceptID && b.URL == normalized {
			return true
		}
	}
	return false
}

func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
