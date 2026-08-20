package concurrent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// FetchResult is one concurrent download outcome.
type FetchResult struct {
	URL    string
	Status int
	Body   string
	Err    error
}

// FetchAll downloads URLs concurrently with the injected client and WaitGroup.
// Results keep input order. A single URL error does not drop the other slots.
func FetchAll(ctx context.Context, client *http.Client, urls []string) []FetchResult {
	if client == nil {
		client = http.DefaultClient
	}
	out := make([]FetchResult, len(urls))
	var wg sync.WaitGroup
	wg.Add(len(urls))
	for i, raw := range urls {
		go func(i int, raw string) {
			defer wg.Done()
			out[i] = fetchOne(ctx, client, raw)
		}(i, raw)
	}
	wg.Wait()
	return out
}

// FetchWithTimeout cancels the HTTP request with context.WithTimeout.
// The call runs on this goroutine so a timeout cannot leave work running behind the caller.
func FetchWithTimeout(client *http.Client, url string, d time.Duration) FetchResult {
	if client == nil {
		client = http.DefaultClient
	}
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return fetchOne(ctx, client, url)
}

func fetchOne(ctx context.Context, client *http.Client, raw string) FetchResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return FetchResult{URL: raw, Err: fmt.Errorf("fetch %s: %w", raw, err)}
	}
	resp, err := client.Do(req)
	if err != nil {
		return FetchResult{URL: raw, Err: fmt.Errorf("fetch %s: %w", raw, err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return FetchResult{URL: raw, Status: resp.StatusCode, Err: fmt.Errorf("fetch %s: %w", raw, err)}
	}
	return FetchResult{URL: raw, Status: resp.StatusCode, Body: string(body)}
}
