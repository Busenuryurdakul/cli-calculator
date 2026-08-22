package notes

import (
	"sort"
	"strconv"
	"sync"
)

// Note is an in-memory note record.
type Note struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Store keeps notes in memory. Every map access uses RWMutex.
type Store struct {
	mu    sync.RWMutex
	seq   int
	notes map[string]Note
}

// NewStore returns an empty note store.
func NewStore() *Store {
	return &Store{notes: make(map[string]Note)}
}

func (s *Store) Create(title, body string) Note {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	n := Note{ID: strconv.Itoa(s.seq), Title: title, Body: body}
	s.notes[n.ID] = n
	return n
}

func (s *Store) Get(id string) (Note, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.notes[id]
	return n, ok
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.notes[id]; !ok {
		return false
	}
	delete(s.notes, id)
	return true
}

func (s *Store) List() []Note {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.notes))
	for id := range s.notes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]Note, 0, len(ids))
	for _, id := range ids {
		out = append(out, s.notes[id])
	}
	return out
}
