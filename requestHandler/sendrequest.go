package requesthandler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Result struct {
	StatusCode int
	Err        error
}

type URLResult struct {
	URL    string
	Result Result
}

type HTTPGetter interface {
	Do(req *http.Request) (*http.Response, error)
}

type SpyGetter struct {
	mu    sync.Mutex
	Calls int
}

func (s *SpyGetter) Do(req *http.Request) (*http.Response, error) {
	// this lock and unlock is to prevent data race
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Calls++
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

// CheckURLs checks the status of multiple URLs concurrently using a worker pool
func CheckURLs(ctx context.Context, urls []string, getter HTTPGetter) map[string]Result {
	result := make(map[string]Result)

	// Buffer channels to prevent blocking
	jobs := make(chan string, len(urls))
	results := make(chan URLResult, len(urls))

	const workerCount = 3
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ctx, jobs, results, getter, &wg)
	}

	// Feed jobs
	go func() {
		for _, u := range urls {
			if len(u) == 0 || !isValidURL(u) {
				results <- URLResult{
					URL:    u,
					Result: Result{StatusCode: 0, Err: fmt.Errorf("invalid url: %s", u)},
				}
				continue
			}

			select {
			case jobs <- u:
			case <-ctx.Done():
				close(jobs)
				return
			}
		}
		close(jobs)
	}()

	// Close results when all workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for r := range results {
		result[r.URL] = r.Result
	}

	return result
}

func worker(
	ctx context.Context,
	jobs <-chan string,
	results chan<- URLResult,
	getter HTTPGetter,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for url := range jobs {
		select {
		case <-ctx.Done():
			return
		case results <- URLResult{URL: url, Result: checkURL(ctx, url, getter)}:
		}
	}
}

func checkURL(ctx context.Context, urlStr string, getter HTTPGetter) Result {
	if err := ctx.Err(); err != nil {
		return Result{Err: err}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return Result{Err: err}
	}

	resp, err := getter.Do(req)
	if err != nil {
		return Result{Err: err}
	}
	defer resp.Body.Close()

	return Result{StatusCode: resp.StatusCode}
}

func isValidURL(urlStr string) bool {
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme == "https" || u.Scheme == "http"
}
