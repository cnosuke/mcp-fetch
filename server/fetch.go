package server

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/cnosuke/mcp-fetch/config"
	"github.com/cnosuke/mcp-fetch/types"
	"github.com/cockroachdb/errors"
	"github.com/go-shiori/go-readability"
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

	zap.S().Debugw(
		"response received",
		"url", url,
		"status", resp.StatusCode,
		"content-length", resp.ContentLength,
		"bytes", len(body),
		"content_type", resp.Header.Get("Content-Type"),
	)

	contentType := resp.Header.Get("Content-Type")
	content := string(body)

	// Process content based on content type
	if strings.Contains(contentType, "text/html") {
		processedContent, err := s.processHTMLContent(content, url)
		if err != nil {
			zap.S().Warnw("failed to process HTML content with readability, falling back to basic conversion", "error", err)
			// If readability fails, try fallback HTML-to-markdown conversion
			basicMarkdown, fallbackErr := s.convertHTMLToMarkdown(content)
			if fallbackErr != nil {
				zap.S().Warnw("fallback HTML conversion also failed", "error", fallbackErr)
				// If all conversions fail, return the original content
			} else {
				content = basicMarkdown
			}
		} else {
			content = processedContent
		}
	} else {
		// For non-HTML content, just return as is
		content = s.processNonHTMLContent(content, contentType)
	}

	return &types.FetchResponse{
		URL:         url,
		ContentType: contentType,
		Content:     content,
		StatusCode:  resp.StatusCode,
	}, nil
}

// processHTMLContent - Process HTML content using readability and convert to markdown
func (s *FetchServer) processHTMLContent(htmlContent, urlStr string) (string, error) {
	// Parse URL string to *url.URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse URL")
	}

	// Use readability to extract the main content
	article, err := readability.FromReader(strings.NewReader(htmlContent), parsedURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to extract content with readability")
	}

	// Convert the extracted content to Markdown
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(article.Content)
	if err != nil {
		return article.Content, errors.Wrap(err, "failed to convert extracted content to Markdown")
	}

	// Add title as heading if available
	if article.Title != "" {
		markdown = "# " + article.Title + "\n\n" + markdown
	}

	// Add excerpt/description if available
	if article.Excerpt != "" {
		markdown = markdown + "\n\n---\n\n" + article.Excerpt
	}

	zap.S().Debugw(
		"processed HTML content",
		"url", urlStr,
		"title", article.Title,
		"length", len(markdown),
	)

	return markdown, nil
}

// processNonHTMLContent - Process non-HTML content
func (s *FetchServer) processNonHTMLContent(content, contentType string) string {
	// For now, we just return the content as is
	// But we could add more processing for different content types in the future
	zap.S().Debugw(
		"non-HTML content detected, returning as is",
		"content_type", contentType,
		"length", len(content),
	)
	return content
}

// convertHTMLToMarkdown - Convert HTML to Markdown directly
func (s *FetchServer) convertHTMLToMarkdown(htmlContent string) (string, error) {
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(htmlContent)
	if err != nil {
		return htmlContent, errors.Wrap(err, "failed to convert HTML to Markdown")
	}
	return markdown, nil
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
