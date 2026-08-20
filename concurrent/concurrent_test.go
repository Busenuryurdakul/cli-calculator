package concurrent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPingUnbuffered(t *testing.T) {
	if got := Ping(); got != "pong" {
		t.Fatalf("Ping()=%q, want pong", got)
	}
}

func TestDrainBuffer(t *testing.T) {
	got := DrainBuffer(3)
	if len(got) != 3 || got[0] != 0 || got[2] != 2 {
		t.Fatalf("DrainBuffer=%v, want 0..2", got)
	}
}

func TestRunPool(t *testing.T) {
	jobs := []Job{{N: 2}, {N: 3}, {N: 4}}
	got := RunPool(jobs, 2)
	if len(got) != 3 {
		t.Fatalf("got %d results, want 3", len(got))
	}
	seen := map[int]int{}
	for _, r := range got {
		seen[r.N] = r.Square
	}
	if seen[2] != 4 || seen[3] != 9 || seen[4] != 16 {
		t.Fatalf("unexpected results: %#v", seen)
	}
}

func TestRecvTimeout(t *testing.T) {
	ch := make(chan int)
	if _, ok := RecvTimeout(ch, 20*time.Millisecond); ok {
		t.Fatal("expected timeout")
	}

	ready := make(chan int, 1)
	ready <- 7
	v, ok := RecvTimeout(ready, time.Second)
	if !ok || v != 7 {
		t.Fatalf("got (%d, %v), want (7, true)", v, ok)
	}
}

func TestTrySendDropsWhenFull(t *testing.T) {
	ch := make(chan int, 1)
	if !TrySend(ch, 1) {
		t.Fatal("first send should fit in buffer")
	}
	if TrySend(ch, 2) {
		t.Fatal("second send should drop (backpressure)")
	}
}

func TestSquareUntilCancel(t *testing.T) {
	jobs := make(chan int)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	var got []int
	go func() {
		defer wg.Done()
		got = SquareUntilCancel(ctx, jobs)
	}()

	jobs <- 3
	jobs <- 4
	cancel()
	wg.Wait()

	if len(got) == 0 {
		t.Fatal("expected some squared values before cancel")
	}
}

func TestSafeCounter(t *testing.T) {
	var c SafeCounter
	var wg sync.WaitGroup
	wg.Add(50)
	for i := 0; i < 50; i++ {
		go func() {
			defer wg.Done()
			c.Add(1)
		}()
	}
	wg.Wait()
	if c.Value() != 50 {
		t.Fatalf("counter=%d, want 50", c.Value())
	}
}

func TestSafeMap(t *testing.T) {
	m := NewSafeMap()
	var wg sync.WaitGroup
	wg.Add(40)
	for i := 0; i < 40; i++ {
		go func() {
			defer wg.Done()
			m.Add("hits", 1)
		}()
	}
	wg.Wait()
	if m.Get("hits") != 40 {
		t.Fatalf("hits=%d, want 40", m.Get("hits"))
	}
}

func TestOnceValue(t *testing.T) {
	var o OnceValue
	var wg sync.WaitGroup
	wg.Add(20)
	for i := 0; i < 20; i++ {
		go func() {
			defer wg.Done()
			_ = o.Load()
		}()
	}
	wg.Wait()
	if o.Load() != "ready" {
		t.Fatalf("Load()=%q, want ready", o.Load())
	}
	if o.Inits() != 1 {
		t.Fatalf("Once initialized %d times, want 1", o.Inits())
	}
}

func TestAtomicAdds(t *testing.T) {
	if got := AtomicAdds(40); got != 40 {
		t.Fatalf("AtomicAdds=%d, want 40", got)
	}
}

func TestFetchAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer srv.Close()

	urls := []string{srv.URL + "/a", srv.URL + "/b"}
	got := FetchAll(context.Background(), srv.Client(), urls)
	if len(got) != 2 {
		t.Fatalf("got %d results", len(got))
	}
	for i, r := range got {
		if r.Err != nil || r.Status != http.StatusOK || r.URL != urls[i] {
			t.Fatalf("unexpected fetch[%d]: %#v", i, r)
		}
	}
	if got[0].Body != "/a" || got[1].Body != "/b" {
		t.Fatalf("results out of order: %q %q", got[0].Body, got[1].Body)
	}
}

func TestFetchAllKeepsOtherResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer srv.Close()

	urls := []string{srv.URL + "/ok", "http://[", srv.URL + "/still"}
	got := FetchAll(context.Background(), srv.Client(), urls)
	if len(got) != 3 {
		t.Fatalf("got %d results, want 3", len(got))
	}
	if got[0].Err != nil || got[0].Body != "/ok" {
		t.Fatalf("first result lost: %#v", got[0])
	}
	if got[1].Err == nil {
		t.Fatal("middle URL should fail")
	}
	if got[2].Err != nil || got[2].Body != "/still" {
		t.Fatalf("third result lost: %#v", got[2])
	}
}

func TestFetchWithTimeout(t *testing.T) {
	var cancelled atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/slow") {
			select {
			case <-r.Context().Done():
				cancelled.Store(true)
				return
			case <-time.After(2 * time.Second):
			}
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	fast := FetchWithTimeout(srv.Client(), srv.URL+"/ok", time.Second)
	if fast.Err != nil || fast.Status != http.StatusOK {
		t.Fatalf("fast fetch: %#v", fast)
	}

	slow := FetchWithTimeout(srv.Client(), srv.URL+"/slow", 30*time.Millisecond)
	if !errors.Is(slow.Err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline, got %#v", slow)
	}

	deadline := time.Now().Add(time.Second)
	for !cancelled.Load() && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if !cancelled.Load() {
		t.Fatal("httptest handler did not observe r.Context().Done()")
	}
}

func TestSquarePipeline(t *testing.T) {
	got := SquarePipeline(4)
	want := []int{1, 4, 9, 16}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
