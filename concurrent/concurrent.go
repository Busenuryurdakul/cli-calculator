package concurrent

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Ping exchanges one value on an unbuffered channel (send and receive sync).
func Ping() string {
	ch := make(chan string)
	go func() {
		ch <- "pong"
		close(ch)
	}()
	return <-ch
}

// DrainBuffer fills a buffered channel, closes it, and ranges until empty.
func DrainBuffer(n int) []int {
	ch := make(chan int, n)
	for i := 0; i < n; i++ {
		ch <- i
	}
	close(ch)

	out := make([]int, 0, n)
	for v := range ch {
		out = append(out, v)
	}
	return out
}

// Job is work for the worker pool.
type Job struct {
	N int
}

// Result is a completed job.
type Result struct {
	N      int
	Square int
}

// RunPool fans jobs out to workers and collects results.
// The job producer closes jobCh. resCh closes only after every worker returns.
func RunPool(jobs []Job, workers int) []Result {
	if workers < 1 {
		workers = 1
	}
	jobCh := make(chan Job)
	resCh := make(chan Result, len(jobs))

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for job := range jobCh {
				resCh <- Result{N: job.N, Square: job.N * job.N}
			}
		}()
	}

	go func() {
		for _, job := range jobs {
			jobCh <- job
		}
		close(jobCh)
	}()

	go func() {
		wg.Wait()
		close(resCh)
	}()

	out := make([]Result, 0, len(jobs))
	for r := range resCh {
		out = append(out, r)
	}
	return out
}

// RecvTimeout waits on ch or a timer. ok is false on timeout.
func RecvTimeout(ch <-chan int, d time.Duration) (v int, ok bool) {
	select {
	case v, ok = <-ch:
		return v, ok
	case <-time.After(d):
		return 0, false
	}
}

// TrySend sends v if the buffer has room; otherwise it drops (backpressure).
func TrySend(ch chan int, v int) bool {
	select {
	case ch <- v:
		return true
	default:
		return false
	}
}

// SquareUntilCancel reads jobs until ctx is cancelled.
func SquareUntilCancel(ctx context.Context, jobs <-chan int) []int {
	var out []int
	for {
		select {
		case <-ctx.Done():
			return out
		case n, ok := <-jobs:
			if !ok {
				return out
			}
			out = append(out, n*n)
		}
	}
}

// SafeCounter protects a counter with a mutex.
type SafeCounter struct {
	mu sync.Mutex
	n  int
}

func (c *SafeCounter) Add(delta int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n += delta
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

// SafeMap protects every map access with a mutex.
type SafeMap struct {
	mu sync.Mutex
	m  map[string]int
}

func NewSafeMap() *SafeMap {
	return &SafeMap{m: make(map[string]int)}
}

func (s *SafeMap) Add(key string, delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] += delta
}

func (s *SafeMap) Get(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.m[key]
}

// OnceValue initializes expensive state exactly once.
type OnceValue struct {
	once sync.Once
	v    string
	n    atomic.Int32
}

func (o *OnceValue) Load() string {
	o.once.Do(func() {
		o.n.Add(1)
		o.v = "ready"
	})
	return o.v
}

func (o *OnceValue) Inits() int32 {
	return o.n.Load()
}

// AtomicAdds increments a shared counter from many goroutines.
func AtomicAdds(goroutines int) int64 {
	var n atomic.Int64
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			n.Add(1)
		}()
	}
	wg.Wait()
	return n.Load()
}

// SquarePipeline is generate → square → collect.
// Each producer closes only the channel it writes.
func SquarePipeline(n int) []int {
	gen := make(chan int)
	squared := make(chan int)

	go func() {
		defer close(gen)
		for i := 1; i <= n; i++ {
			gen <- i
		}
	}()
	go func() {
		defer close(squared)
		for v := range gen {
			squared <- v * v
		}
	}()

	out := make([]int, 0, n)
	for v := range squared {
		out = append(out, v)
	}
	return out
}

// CompareSummary explains when channels vs mutexes are clearer.
func CompareSummary() string {
	return `Channels vs mutexes:

- Channels: ownership moves with the value (jobs, results, cancellation)
- Mutex/atomic: many goroutines share one in-memory structure (counters, caches)
- sync.Once: lazy init of clients, config, and templates
- Prefer the model that makes who may write the data obvious`
}

// ConcurrentSummary documents the four concurrency practice blocks.
func ConcurrentSummary() string {
	return `Concurrency patterns in this package:

- Unbuffered channels synchronize send/receive; buffers decouple bursts
- Close + range drains a producer; worker pools use job and result channels
- select waits on many ops; time.After is for RecvTimeout, not HTTP
- HTTP timeouts use context.WithTimeout so the request is cancelled
- Backpressure: block, buffer, or drop with select/default
- Mutex/Once/atomic protect shared memory when a channel would be ceremony
- WaitGroup downloader, three-stage pipeline; run tests with -race in CI`
}
