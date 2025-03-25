package server

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown"
	"github.com/cnosuke/mcp-fetch/config"
	"github.com/cnosuke/mcp-fetch/types"
	"github.com/cockroachdb/errors"
	"go.uber.org/zap"
)

// FetchServer - Fetch server structure
type FetchServer struct {
	client     *http.Client
	userAgent  string
	maxWorkers int
}

// NewFetchServer - Create a new Fetch server
func NewFetchServer(cfg *config.Config) (*FetchServer, error) {
	zap.S().Infow("creating new Fetch server",
		"timeout", cfg.Fetch.Timeout,
		"user_agent", cfg.Fetch.UserAgent,
		"max_workers", cfg.Fetch.MaxWorkers)

	client := &http.Client{
		Timeout: time.Duration(cfg.Fetch.Timeout) * time.Second,
	}

	return &FetchServer{
		client:     client,
		userAgent:  cfg.Fetch.UserAgent,
		maxWorkers: cfg.Fetch.MaxWorkers,
	}, nil
}

// FetchURL - Fetch content from a URL
func (s *FetchServer) FetchURL(url string) (*types.FetchResponse, error) {
	zap.S().Debugw("fetching URL", "url", url)

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	// Set headers
	req.Header.Set("User-Agent", s.userAgent)

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	contentType := resp.Header.Get("Content-Type")
	content := string(body)

	// Process content based on content type
	if strings.Contains(contentType, "text/html") {
		// Convert HTML to Markdown
		converter := md.NewConverter("", true, nil)
		md, err := converter.ConvertString(content)
		if err != nil {
			zap.S().Warnw("failed to convert HTML to Markdown", "error", err)
			// Return original content if conversion fails
		} else {
			content = md
		}
	}

	return &types.FetchResponse{
		URL:         url,
		ContentType: contentType,
		Content:     content,
		StatusCode:  resp.StatusCode,
	}, nil
}

// FetchMultipleURLs - Fetch content from multiple URLs in parallel
func (s *FetchServer) FetchMultipleURLs(urls []string) (*types.MultipleFetchResponse, error) {
	zap.S().Debugw("fetching multiple URLs", "count", len(urls), "workers", s.maxWorkers)

	// Create response struct with maps
	response := &types.MultipleFetchResponse{
		Responses: make(map[string]*types.FetchResponse),
		Errors:    make(map[string]string),
	}

	// Create a wait group to wait for all workers to finish
	wg := &sync.WaitGroup{}

	// Create mutex to protect concurrent map access
	mu := &sync.Mutex{}

	// Create a worker pool with channels
	jobs := make(chan string, len(urls))

	// Determine number of workers (use maxWorkers, but not more than URLs)
	nWorkers := s.maxWorkers
	if nWorkers > len(urls) {
		nWorkers = len(urls)
	}

	// Start workers
	for w := 1; w <= nWorkers; w++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			zap.S().Debugw("starting worker", "worker_id", workerId)

			for url := range jobs {
				// Process URL
				res, err := s.FetchURL(url)

				// Lock for map access
				mu.Lock()
				if err != nil {
					// Store error
					response.Errors[url] = err.Error()
					zap.S().Debugw("fetch failed", "worker_id", workerId, "url", url, "error", err)
				} else {
					// Store successful response
					response.Responses[url] = res
					zap.S().Debugw("fetch successful", "worker_id", workerId, "url", url, "status", res.StatusCode)
				}
				mu.Unlock()
			}
		}(w)
	}

	// Send URLs to the worker pool
	for _, url := range urls {
		jobs <- url
	}
	
	// Close the jobs channel to signal workers that no more jobs are coming
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()
	
	zap.S().Infow("completed fetching multiple URLs", 
		"total", len(urls),
		"success", len(response.Responses),
		"errors", len(response.Errors))

	return response, nil
}
