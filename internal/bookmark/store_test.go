package bookmark

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMemoryStoreCRUDAndCopies(t *testing.T) {
	s := NewMemoryStore()
	created, err := s.Create("Docs", "https://go.dev", []string{"go", "docs"})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Title != "Docs" || len(created.Tags) != 2 {
		t.Fatalf("created=%+v", created)
	}

	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	got.Title = "mutated"
	got.Tags[0] = "hacked"
	again, err := s.Get(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if again.Title != "Docs" || again.Tags[0] != "go" {
		t.Fatal("Get leaked internal state")
	}

	listed := s.List()
	listed[0].Title = "mutated-list"
	listed[0].Tags[0] = "list-hack"
	listed2 := s.List()
	if listed2[0].Title != "Docs" || listed2[0].Tags[0] != "go" {
		t.Fatal("List leaked internal state")
	}

	inputTags := []string{"tour"}
	updated, err := s.Update(created.ID, "Tour", "https://go.dev/tour", inputTags)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Tour" || updated.Tags[0] != "tour" {
		t.Fatalf("updated=%+v", updated)
	}
	inputTags[0] = "mutated-input"
	afterInput, err := s.Get(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if afterInput.Tags[0] != "tour" {
		t.Fatal("Update input slice leaked into store")
	}

	title := "Blog"
	patched, err := s.Patch(created.ID, &title, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if patched.Title != "Blog" || patched.URL != "https://go.dev/tour" || patched.Tags[0] != "tour" {
		t.Fatalf("patched=%+v", patched)
	}

	empty := []string{}
	cleared, err := s.Patch(created.ID, nil, nil, &empty)
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.Tags) != 0 {
		t.Fatalf("cleared tags=%v", cleared.Tags)
	}

	if err := s.Delete(created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(created.ID); err != ErrNotFound {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestMemoryStoreValidation(t *testing.T) {
	s := NewMemoryStore()
	if _, err := s.Create("  ", "https://go.dev", nil); err == nil {
		t.Fatal("expected empty title error")
	}
	if _, err := s.Create("ok", "ftp://x", nil); err == nil {
		t.Fatal("expected url error")
	}
	if _, err := s.Create("ok", "https://go.dev", []string{" "}); err == nil {
		t.Fatal("expected empty tag error")
	}
	if _, err := s.Get("not-a-uuid"); err != ErrInvalidID {
		t.Fatalf("got %v, want ErrInvalidID", err)
	}
}

func TestMemoryStoreDuplicateURL(t *testing.T) {
	s := NewMemoryStore()
	first, err := s.Create("A", "https://example.com/x", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create("B", "  https://example.com/x  ", nil); err != ErrConflict {
		t.Fatalf("got %v, want ErrConflict", err)
	}
	second, err := s.Create("C", "https://example.com/y", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(second.ID, "C", first.URL, nil); err != ErrConflict {
		t.Fatalf("update conflict got %v", err)
	}
	own, err := s.Update(first.ID, "A2", first.URL, []string{"keep"})
	if err != nil {
		t.Fatalf("own url update: %v", err)
	}
	if own.URL != first.URL || own.Title != "A2" {
		t.Fatalf("own=%+v", own)
	}
}

func TestStoresAreIndependent(t *testing.T) {
	a := NewMemoryStore()
	b := NewMemoryStore()
	if _, err := a.Create("A", "https://a.example", nil); err != nil {
		t.Fatal(err)
	}
	if got := b.List(); len(got) != 0 {
		t.Fatalf("stores share state: %#v", got)
	}
}

func TestListOrderDeterministic(t *testing.T) {
	s := NewMemoryStore()
	s.newID = func() (string, error) { return "bbbbbbbb-1111-4111-8111-111111111111", nil }
	if _, err := s.Create("B", "https://b.example", nil); err != nil {
		t.Fatal(err)
	}
	s.newID = func() (string, error) { return "aaaaaaaa-1111-4111-8111-111111111111", nil }
	if _, err := s.Create("A", "https://a.example", nil); err != nil {
		t.Fatal(err)
	}
	listed := s.List()
	if len(listed) != 2 || listed[0].ID > listed[1].ID {
		t.Fatalf("order=%v", []string{listed[0].ID, listed[1].ID})
	}
}

func TestConcurrentCreatesUniqueIDs(t *testing.T) {
	s := NewMemoryStore()
	const n = 40
	ids := make(chan string, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			b, err := s.Create("t", fmt.Sprintf("https://example.com/%d", i), nil)
			if err != nil {
				t.Errorf("%v", err)
				return
			}
			ids <- b.ID
		}()
	}
	wg.Wait()
	close(ids)
	seen := map[string]bool{}
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate id %s", id)
		}
		seen[id] = true
	}
	if len(seen) != n {
		t.Fatalf("got %d ids, want %d", len(seen), n)
	}
}

func TestConcurrentCreateReadUpdate(t *testing.T) {
	s := NewMemoryStore()
	seed, err := s.Create("seed", "https://example.com/seed", []string{"go"})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 25; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			if _, err := s.Get(seed.ID); err != nil {
				t.Errorf("get: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			_ = s.List()
		}()
		i := i
		go func() {
			defer wg.Done()
			_, err := s.Update(seed.ID, fmt.Sprintf("t%d", i), "https://example.com/seed", []string{"go"})
			if err != nil {
				t.Errorf("update: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestDeterministicIDInjection(t *testing.T) {
	s := NewMemoryStore()
	s.now = func() time.Time { return time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC) }
	s.newID = func() (string, error) { return "11111111-1111-4111-8111-111111111111", nil }
	b, err := s.Create("A", "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	if b.ID != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("id=%s", b.ID)
	}
}
