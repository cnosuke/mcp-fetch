package fetcher

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	ierrors "github.com/cnosuke/mcp-fetch/internal/errors"
	"github.com/cnosuke/mcp-fetch/types"
	"github.com/mackee/go-readability"
	"go.uber.org/zap"
)

type Config struct {
	Timeout          int
	UserAgent        string
	MaxURLs          int
	MaxWorkers       int
	DefaultMaxLength int
}

// Fetcher defines the interface for fetching and processing URL content.
type Fetcher interface {
	// Fetch fetches and processes content from a single URL.
	// It handles HTTP requests, content extraction (using readability),
	// Markdown conversion, and content trimming based on parameters.
	Fetch(urlStr string, maxLength int, startIndex int, raw bool) (*types.FetchResponse, error)

	// FetchMultiple fetches and processes content from multiple URLs.
	// It handles parallel fetching and content reallocation logic.
	FetchMultiple(urls []string, maxLength int, raw bool) (*types.MultipleFetchResponse, error)
}

// httpFetcher implements the Fetcher interface using HTTP.
type httpFetcher struct {
	client           *http.Client
	userAgent        string
	maxWorkers       int
	defaultMaxLength int
}

// NewHTTPFetcher creates a new httpFetcher.
func NewHTTPFetcher(cfg *Config) (Fetcher, error) {
	zap.S().Infow("creating new HTTP fetcher",
		"timeout", cfg.Timeout,
		"user_agent", cfg.UserAgent,
		"max_workers", cfg.MaxWorkers,
		"default_max_length", cfg.DefaultMaxLength)

	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	return &httpFetcher{
		client:           client,
		userAgent:        cfg.UserAgent,
		maxWorkers:       cfg.MaxWorkers,
		defaultMaxLength: cfg.DefaultMaxLength,
	}, nil
}

type fetchResponse struct {
	url         string
	status      int
	body        string
	contentType string
	originalURL string // Set only if redirect occurred
	err         error
}

func (f *httpFetcher) fetch(urlStr string) *fetchResponse {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return &fetchResponse{err: ierrors.Wrap(err, "failed to create request")}
	}
	req.Header.Set("User-Agent", f.userAgent)

	// Track redirect chain for this request
	var redirectChain []string
	// Save original CheckRedirect to restore later
	origCheckRedirect := f.client.CheckRedirect
	f.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// Record the redirect chain
		if len(via) == 1 {
			redirectChain = append(redirectChain, via[0].URL.String())
		}
		redirectChain = append(redirectChain, req.URL.String())
		// Default policy: allow up to 10 redirects
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}

	resp, err := f.client.Do(req)
	// Restore original CheckRedirect
	f.client.CheckRedirect = origCheckRedirect

	if err != nil {
		return &fetchResponse{err: ierrors.Wrap(err, "failed to execute request")}
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &fetchResponse{err: ierrors.Wrap(err, "failed to read response body")}
	}

	zap.S().Debugw(
		"response received",
		"url", urlStr,
		"status", resp.StatusCode,
		"content-length", resp.ContentLength,
		"bytes", len(bodyBytes),
		"content_type", resp.Header.Get("Content-Type"),
	)

	originalURL := ""
	if len(redirectChain) > 0 && redirectChain[0] != resp.Request.URL.String() {
		originalURL = urlStr
	}

	return &fetchResponse{
		url:         resp.Request.URL.String(),
		status:      resp.StatusCode,
		body:        string(bodyBytes),
		contentType: resp.Header.Get("Content-Type"),
		originalURL: originalURL,
		err:         nil,
	}
}

// Fetch fetches and processes content from a single URL.
func (f *httpFetcher) Fetch(urlStr string, maxLength int, startIndex int, raw bool) (*types.FetchResponse, error) {
	zap.S().Debugw("fetching URL",
		"url", urlStr,
		"max_length", maxLength,
		"start_index", startIndex,
		"raw", raw)

	// Fetch the URL using the internal fetch method
	resp := f.fetch(urlStr)
	if resp.err != nil {
		// Error is already wrapped in f.fetch
		return nil, resp.err
	}
	var processedContent string

	if raw {
		processedContent = resp.body
		zap.S().Debugw("raw mode enabled", "url", urlStr)
	} else if strings.Contains(resp.contentType, "text/html") {
		processedContent = processHTMLContent(resp.body, urlStr)
	} else {
		processedContent = resp.body // Non-HTML content
		zap.S().Debugw("non-HTML content", "url", urlStr, "content_type", resp.contentType)
	}

	// Apply trimming
	trimmedContent := trimContent(processedContent, startIndex, maxLength)
	if len(processedContent) != len(trimmedContent) {
		zap.S().Debugw("content trimmed",
			"original_length", len(processedContent),
			"start_index", startIndex,
			"trimmed_length", len(trimmedContent))
	}

	return &types.FetchResponse{
		URL:         urlStr,
		ContentType: resp.contentType,
		Content:     trimmedContent,
		StatusCode:  resp.status,
		// Set only if redirect occurred
		OriginalURL: resp.originalURL,
	}, nil
}

// trimContent helper function to trim content based on startIndex and maxLength
func trimContent(content string, startIndex int, maxLength int) string {
	contentLength := len(content)
	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex >= contentLength {
		return ""
	}
	endIndex := contentLength
	if maxLength > 0 {
		potentialEndIndex := startIndex + maxLength
		if potentialEndIndex < endIndex {
			endIndex = potentialEndIndex
		}
	}
	return content[startIndex:endIndex]
}

// processHTMLContent extracts content from HTML using readability and converts it to Markdown.
// It falls back to the raw body string if readability fails.
func processHTMLContent(body string, urlStr string) string {
	opts := readability.DefaultOptions()
	article, readErr := readability.Extract(body, opts)
	if readErr != nil {
		zap.S().Warnw("readability extraction failed, falling back to raw body", "url", urlStr, "error", readErr)
		return body
	}

	// Convert extracted content to Markdown
	markdownContent := readability.ToMarkdown(article.Root)

	finalMarkdown := ""
	if article.Title != "" {
		finalMarkdown += "# " + article.Title + "\n\n" // Add title as H1
	}
	finalMarkdown += markdownContent
	// Add byline if available
	if article.Byline != "" {
		finalMarkdown += "\n\n---\n\nAuthor: " + article.Byline // Add byline
	}

	zap.S().Debugw("processed HTML with readability to Markdown",
		"url", urlStr,
		"title", article.Title,
		"byline", article.Byline,
		"markdown_length", len(finalMarkdown))

	return finalMarkdown
}

// FetchMultiple fetches content from multiple URLs in parallel and allocates content length.
func (f *httpFetcher) FetchMultiple(urls []string, maxLength int, raw bool) (*types.MultipleFetchResponse, error) {
	zap.S().Debugw("fetching multiple URLs",
		"count", len(urls),
		"max_length", maxLength,
		"raw", raw,
		"workers", f.maxWorkers)

	// Default value if maxLength is not specified
	if maxLength <= 0 {
		maxLength = f.defaultMaxLength
	}

	// Slice to store initial fetch results
	results := make([]*fetchResponse, len(urls))

	wg := &sync.WaitGroup{}

	// Fetch URLs in parallel
	for i, url := range urls {
		wg.Add(1)
		go func(index int, urlStr string) {
			defer wg.Done()

			zap.S().Debugw("initiating fetch for URL", "url", urlStr)

			resp := f.fetch(urlStr)
			results[index] = resp
		}(i, url)
	}

	// Wait for all initial fetches to complete
	wg.Wait()

	// --- Content Processing and Allocation ---

	// Structure to hold processed content before final trimming
	type processedResult struct {
		URL                 string // The original URL
		FullContent         string // Content after readability/markdown, before any trimming
		ContentType         string
		StatusCode          int
		OriginalIndex       int    // To maintain order if needed
		FinalTrimmedContent string // Content after allocation and trimming
		WasTruncated        bool   // Flag if initial allocation truncated content
	}

	processedResults := []*processedResult{}
	finalErrors := make(map[string]string)

	// Process successful fetches results
	for i, res := range results {
		if res == nil {
			zap.S().Errorw("encountered nil initial result", "index", i, "url", urls[i])
			finalErrors[urls[i]] = "internal error: nil initial result"
			continue
		}
		if res.err != nil {
			finalErrors[res.url] = res.err.Error()
			continue
		}

		var processedContent string
		if raw {
			processedContent = res.body
			zap.S().Debugw("raw mode enabled for multiple fetch", "url", res.url)
		} else if strings.Contains(res.contentType, "text/html") {
			processedContent = processHTMLContent(res.body, res.url)
		} else {
			processedContent = string(res.body) // Non-HTML content
			zap.S().Debugw("non-HTML content for multiple fetch", "url", res.url, "content_type", res.contentType)
		}
		// Append the processed content result
		processedResults = append(processedResults, &processedResult{
			URL:           res.url,
			FullContent:   processedContent,
			ContentType:   res.contentType,
			StatusCode:    res.status,
			OriginalIndex: i,
		})
	} // End of for loop processing results

	// --- Allocation Logic ---
	numSuccessful := len(processedResults)
	var initialAllocation int
	if numSuccessful > 0 {
		initialAllocation = maxLength / numSuccessful
	} else {
		initialAllocation = 0 // Avoid division by zero if all fetches failed
	}

	zap.S().Debugw("calculated initial allocation",
		"num_successful", numSuccessful,
		"total_max_length", maxLength,
		"initial_allocation_per_url", initialAllocation)

	totalUsed := 0
	beneficiaries := []*processedResult{} // URLs that were truncated by initial allocation

	// Initial trimming and identify beneficiaries
	for _, res := range processedResults {
		trimmed := trimContent(res.FullContent, 0, initialAllocation)
		res.FinalTrimmedContent = trimmed
		usedChars := len(trimmed)
		totalUsed += usedChars

		// Check if content was actually longer than the allocation allowed
		if len(res.FullContent) > initialAllocation {
			res.WasTruncated = true
			beneficiaries = append(beneficiaries, res)
			zap.S().Debugw("URL identified as beneficiary due to truncation",
				"url", res.URL,
				"full_length", len(res.FullContent),
				"allocated", initialAllocation,
				"trimmed_length", usedChars)
		} else {
			zap.S().Debugw("URL content fits initial allocation or is shorter",
				"url", res.URL,
				"full_length", len(res.FullContent),
				"used", usedChars)
		}
	}

	// Reallocation
	remainingChars := maxLength - totalUsed
	numBeneficiaries := len(beneficiaries)

	zap.S().Debugw("reallocation status",
		"total_used_after_initial", totalUsed,
		"remaining_chars", remainingChars,
		"num_beneficiaries", numBeneficiaries)

	if remainingChars > 0 && numBeneficiaries > 0 {
		perURLReallocation := remainingChars / numBeneficiaries
		zap.S().Debugw("performing reallocation",
			"per_url_reallocation", perURLReallocation)

		extraCharsDistributed := 0
		for _, b := range beneficiaries {
			// Calculate the target length for this beneficiary
			targetLength := initialAllocation + perURLReallocation
			// Re-trim the *full* content with the new target length
			finalTrimmed := trimContent(b.FullContent, 0, targetLength)

			// Calculate how many characters were actually added in this step
			additionalCharsAdded := len(finalTrimmed) - len(b.FinalTrimmedContent)
			if additionalCharsAdded < 0 {
				additionalCharsAdded = 0
			} // Should not happen, but safety

			// Update the total used characters and the beneficiary's content
			totalUsed += additionalCharsAdded
			extraCharsDistributed += additionalCharsAdded
			b.FinalTrimmedContent = finalTrimmed

			zap.S().Debugw("reallocated content to beneficiary",
				"url", b.URL,
				"previous_length", len(b.FinalTrimmedContent)-additionalCharsAdded,
				"additional_chars_added", additionalCharsAdded,
				"new_length", len(b.FinalTrimmedContent),
				"target_length", targetLength)
		}
		zap.S().Debugw("reallocation complete", "extra_chars_distributed", extraCharsDistributed, "final_total_used", totalUsed)
		// Note: A small amount of remainingChars might be left unused due to integer division or if beneficiaries' full content was shorter than targetLength.
	}

	// Build final response
	finalResponse := &types.MultipleFetchResponse{
		Responses: make(map[string]*types.FetchResponse),
		Errors:    finalErrors,
	}

	for _, res := range processedResults {
		finalResponse.Responses[res.URL] = &types.FetchResponse{
			URL:         res.URL,
			ContentType: res.ContentType,
			Content:     res.FinalTrimmedContent,
			StatusCode:  res.StatusCode,
		}
	}

	// Log completion
	finalTotalContentLength := 0
	for _, r := range finalResponse.Responses {
		finalTotalContentLength += len(r.Content)
	}
	zap.S().Infow("completed fetching multiple URLs",
		"total_urls_requested", len(urls),
		"successful_fetches", numSuccessful,
		"error_count", len(finalResponse.Errors),
		"final_total_content_length", finalTotalContentLength,
		"max_length_limit", maxLength)

	return finalResponse, nil
}
