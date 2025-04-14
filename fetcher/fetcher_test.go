package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cnosuke/mcp-fetch/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTrimContent tests the trimContent helper function.
func TestTrimContent(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		startIndex int
		maxLength  int
		expected   string
	}{
		{
			name:       "no trimming needed",
			content:    "Hello World",
			startIndex: 0,
			maxLength:  0, // No max length limit
			expected:   "Hello World",
		},
		{
			name:       "trim with max length",
			content:    "This is a long sentence.",
			startIndex: 0,
			maxLength:  10,
			expected:   "This is a ",
		},
		{
			name:       "trim with start index",
			content:    "This is a long sentence.",
			startIndex: 5,
			maxLength:  0,
			expected:   "is a long sentence.",
		},
		{
			name:       "trim with start index and max length",
			content:    "This is a long sentence.",
			startIndex: 5,
			maxLength:  6,
			expected:   "is a l",
		},
		{
			name:       "start index out of bounds (positive)",
			content:    "Short",
			startIndex: 10,
			maxLength:  5,
			expected:   "",
		},
		{
			name:       "start index negative (treated as 0)",
			content:    "Negative start",
			startIndex: -5,
			maxLength:  8,
			expected:   "Negative",
		},
		{
			name:       "max length exceeds content length from start index",
			content:    "Content",
			startIndex: 3,
			maxLength:  10,
			expected:   "tent",
		},
		{
			name:       "max length exactly matches remaining content",
			content:    "Exact Match",
			startIndex: 6,
			maxLength:  5,
			expected:   "Match",
		},
		{
			name:       "empty content",
			content:    "",
			startIndex: 0,
			maxLength:  10,
			expected:   "",
		},
		{
			name:       "zero max length",
			content:    "Zero Max",
			startIndex: 0,
			maxLength:  0,
			expected:   "Zero Max",
		},
		{
			name:       "zero max length with start index",
			content:    "Zero Max Start",
			startIndex: 5,
			maxLength:  0,
			expected:   "Max Start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := trimContent(tt.content, tt.startIndex, tt.maxLength)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// --- Mock HTTP Server Setup ---

type mockResponse struct {
	Body        string
	ContentType string
	StatusCode  int
}

func startMockServer(t *testing.T, responses map[string]mockResponse) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Allow matching just the path part for simplicity in tests
		if resp, ok := responses[path]; ok {
			w.Header().Set("Content-Type", resp.ContentType)
			w.WriteHeader(resp.StatusCode)
			_, err := w.Write([]byte(resp.Body))
			require.NoError(t, err, "Failed to write response body in mock server")
		} else {
			// Default response for unexpected paths
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("Not Found"))
			require.NoError(t, err, "Failed to write 404 response body in mock server")
		}
	}))
	t.Cleanup(server.Close) // Ensure server is closed after test
	return server
}

// --- Fetcher Initialization ---

func newTestFetcher(t *testing.T, serverURL string) Fetcher {
	t.Helper()
	cfg := &config.Config{
		Fetch: struct {
			Timeout          int    `yaml:"timeout" default:"10" env:"FETCH_TIMEOUT"`
			UserAgent        string `yaml:"user_agent" default:"mcp-fetch/1.0" env:"FETCH_USER_AGENT"`
			MaxURLs          int    `yaml:"max_urls" default:"20" env:"FETCH_MAX_URLS"`
			MaxWorkers       int    `yaml:"max_workers" default:"20" env:"FETCH_MAX_WORKERS"`
			DefaultMaxLength int    `yaml:"default_max_length" default:"5000" env:"FETCH_DEFAULT_MAX_LENGTH"`
		}{
			Timeout:          5, // Short timeout for tests
			UserAgent:        "test-agent/1.0",
			MaxWorkers:       5,
			DefaultMaxLength: 1000,
		},
		// Other config sections can be default or nil if not needed by NewHTTPFetcher
	}
	fetcher, err := NewHTTPFetcher(cfg)
	require.NoError(t, err, "Failed to create test fetcher")

	// If we need to override the client to point to the test server,
	// we might need to adjust NewHTTPFetcher or the httpFetcher struct access.
	// For now, assume we pass the test server URL directly to Fetch/FetchMultiple.
	// Alternatively, modify the client's Transport for redirection (more complex).

	return fetcher
}

// --- Test Cases for Fetch ---

func TestHTTPFetcher_Fetch_Success_HTML(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/html": {
			Body:        "<html><head><title>Test Page</title></head><body><h1>Main Content</h1><p>Some text.</p></body></html>",
			ContentType: "text/html; charset=utf-8",
			StatusCode:  http.StatusOK,
		},
	}
	server := startMockServer(t, mockResponses)
	fetcher := newTestFetcher(t, server.URL)

	urlToFetch := server.URL + "/html"
	maxLength := 100
	startIndex := 0
	raw := false

	resp, err := fetcher.Fetch(urlToFetch, maxLength, startIndex, raw)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, urlToFetch, resp.URL)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.ContentType, "text/html")
	// We don't test the exact markdown conversion per user request,
	// but check that *some* processing happened (it's not the raw HTML).
	assert.NotEqual(t, mockResponses["/html"].Body, resp.Content)
	assert.Contains(t, resp.Content, "Test Page") // Check if title was extracted
	// assert.Contains(t, resp.Content, "Main Content") // Removed: Avoid testing readability details
	assert.LessOrEqual(t, len(resp.Content), maxLength)
}

func TestHTTPFetcher_Fetch_Success_NonHTML(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/plain": {
			Body:        "This is plain text.",
			ContentType: "text/plain",
			StatusCode:  http.StatusOK,
		},
	}
	server := startMockServer(t, mockResponses)
	fetcher := newTestFetcher(t, server.URL)

	urlToFetch := server.URL + "/plain"
	maxLength := 50
	startIndex := 5
	raw := false

	resp, err := fetcher.Fetch(urlToFetch, maxLength, startIndex, raw)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, urlToFetch, resp.URL)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/plain", resp.ContentType)
	// Non-HTML content should be returned as is (after trimming)
	expectedContent := trimContent(mockResponses["/plain"].Body, startIndex, maxLength)
	assert.Equal(t, expectedContent, resp.Content) // "is plain text." trimmed
	assert.Equal(t, "is plain text.", resp.Content)
}

func TestHTTPFetcher_Fetch_Success_Raw(t *testing.T) {
	htmlBody := "<html><body>Raw HTML</body></html>"
	mockResponses := map[string]mockResponse{
		"/rawhtml": {
			Body:        htmlBody,
			ContentType: "text/html",
			StatusCode:  http.StatusOK,
		},
	}
	server := startMockServer(t, mockResponses)
	fetcher := newTestFetcher(t, server.URL)

	urlToFetch := server.URL + "/rawhtml"
	maxLength := 10 // Shorter than content to test trimming
	startIndex := 7
	raw := true

	resp, err := fetcher.Fetch(urlToFetch, maxLength, startIndex, raw)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, urlToFetch, resp.URL)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/html", resp.ContentType)
	// Raw mode should return the original body (after trimming)
	expectedContent := trimContent(htmlBody, startIndex, maxLength)
	assert.Equal(t, expectedContent, resp.Content) // "<body>Raw"
	assert.Equal(t, "body>Raw H", resp.Content)
}

func TestHTTPFetcher_Fetch_Success_Trimming(t *testing.T) {
	longText := "This is a very long text string for testing the trimming functionality."
	mockResponses := map[string]mockResponse{
		"/long": {
			Body:        longText,
			ContentType: "text/plain",
			StatusCode:  http.StatusOK,
		},
	}
	server := startMockServer(t, mockResponses)
	fetcher := newTestFetcher(t, server.URL)

	urlToFetch := server.URL + "/long"

	// Test case 1: Start index and max length
	startIndex1 := 10
	maxLength1 := 20
	resp1, err1 := fetcher.Fetch(urlToFetch, maxLength1, startIndex1, false)
	require.NoError(t, err1)
	require.NotNil(t, resp1)
	expected1 := trimContent(longText, startIndex1, maxLength1) // "very long text strin"
	assert.Equal(t, expected1, resp1.Content)
	assert.Len(t, resp1.Content, maxLength1)

	// Test case 2: Only max length
	startIndex2 := 0
	maxLength2 := 15
	resp2, err2 := fetcher.Fetch(urlToFetch, maxLength2, startIndex2, false)
	require.NoError(t, err2)
	require.NotNil(t, resp2)
	expected2 := trimContent(longText, startIndex2, maxLength2) // "This is a very "
	assert.Equal(t, expected2, resp2.Content)
	assert.Len(t, resp2.Content, maxLength2)

	// Test case 3: Only start index
	startIndex3 := 50
	maxLength3 := 0 // No limit
	resp3, err3 := fetcher.Fetch(urlToFetch, maxLength3, startIndex3, false)
	require.NoError(t, err3)
	require.NotNil(t, resp3)
	expected3 := trimContent(longText, startIndex3, maxLength3) // " trimming functionality."
	assert.Equal(t, expected3, resp3.Content)
}

func TestHTTPFetcher_Fetch_Error_NotFound(t *testing.T) {
	mockResponses := map[string]mockResponse{
		// No entry for /notfound
	}
	server := startMockServer(t, mockResponses)
	fetcher := newTestFetcher(t, server.URL)

	urlToFetch := server.URL + "/notfound"
	resp, err := fetcher.Fetch(urlToFetch, 100, 0, false)

	// Fetch itself doesn't return an error for 404, it returns the response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, "Not Found", resp.Content) // Content from mock server's 404 handler
}

func TestHTTPFetcher_Fetch_Error_ServerDown(t *testing.T) {
	// No server started
	fetcher := newTestFetcher(t, "http://localhost:9999") // Use a non-existent server address

	urlToFetch := "http://localhost:9999/somepath"
	resp, err := fetcher.Fetch(urlToFetch, 100, 0, false)

	require.Error(t, err) // Expect an error because the connection should fail
	require.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to execute request")   // Check for the wrapped error message
	assert.Contains(t, err.Error(), "connect: connection refused") // Check for the underlying network error
}

// --- Test Cases for FetchMultiple ---

func TestHTTPFetcher_FetchMultiple_Success_Simple(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/page1": {Body: "Content page 1", ContentType: "text/plain", StatusCode: http.StatusOK},
		"/page2": {Body: "Content page 2 longer", ContentType: "text/plain", StatusCode: http.StatusOK},
	}
	server := startMockServer(t, mockResponses)
	fetcher := newTestFetcher(t, server.URL)

	urls := []string{server.URL + "/page1", server.URL + "/page2"}
	maxLength := 100 // Enough for both
	raw := true      // Use raw to simplify content checking

	resp, err := fetcher.FetchMultiple(urls, maxLength, raw)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Errors)
	require.Len(t, resp.Responses, 2)

	r1, ok1 := resp.Responses[urls[0]]
	require.True(t, ok1)
	assert.Equal(t, mockResponses["/page1"].Body, r1.Content)
	assert.Equal(t, http.StatusOK, r1.StatusCode)

	r2, ok2 := resp.Responses[urls[1]]
	require.True(t, ok2)
	assert.Equal(t, mockResponses["/page2"].Body, r2.Content)
	assert.Equal(t, http.StatusOK, r2.StatusCode)

	totalLen := len(r1.Content) + len(r2.Content)
	assert.LessOrEqual(t, totalLen, maxLength)
}

func TestHTTPFetcher_FetchMultiple_Success_Allocation(t *testing.T) {
	// Content lengths: 10, 30, 5
	mockResponses := map[string]mockResponse{
		"/short1": {Body: "1234567890", ContentType: "text/plain", StatusCode: http.StatusOK},                        // len 10
		"/long1":  {Body: "This content is thirty chars long", ContentType: "text/plain", StatusCode: http.StatusOK}, // len 30
		"/short2": {Body: "Short", ContentType: "text/plain", StatusCode: http.StatusOK},                             // len 5
	}
	server := startMockServer(t, mockResponses)
	fetcher := newTestFetcher(t, server.URL)

	urls := []string{server.URL + "/short1", server.URL + "/long1", server.URL + "/short2"}
	maxLength := 30 // Less than total (45), allocation needed
	raw := true

	resp, err := fetcher.FetchMultiple(urls, maxLength, raw)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Errors)
	require.Len(t, resp.Responses, 3)

	// Initial allocation: 30 / 3 = 10 per URL
	// /short1: uses 10 (fits)
	// /long1:  uses 10 (truncated) -> beneficiary
	// /short2: uses 5 (fits)
	// Total used initially: 10 + 10 + 5 = 25
	// Remaining: 30 - 25 = 5
	// Beneficiaries: 1 (/long1)
	// Reallocation: 5 / 1 = 5 extra for /long1
	// Final lengths: /short1=10, /long1=10+5=15, /short2=5
	// Final total: 10 + 15 + 5 = 30

	rShort1, _ := resp.Responses[urls[0]]
	rLong1, _ := resp.Responses[urls[1]]
	rShort2, _ := resp.Responses[urls[2]]

	assert.Equal(t, "1234567890", rShort1.Content, "short1 content mismatch")    // Full content
	assert.Equal(t, "This content is", rLong1.Content, "long1 content mismatch") // Truncated to 15
	assert.Equal(t, "Short", rShort2.Content, "short2 content mismatch")         // Full content

	assert.Len(t, rShort1.Content, 10)
	assert.Len(t, rLong1.Content, 15)
	assert.Len(t, rShort2.Content, 5)

	totalLen := len(rShort1.Content) + len(rLong1.Content) + len(rShort2.Content)
	assert.Equal(t, maxLength, totalLen, "Total length should match maxLength")
}

func TestHTTPFetcher_FetchMultiple_Success_Allocation_MoreRemaining(t *testing.T) {
	// Content lengths: 5, 5, 5
	mockResponses := map[string]mockResponse{
		"/s1": {Body: "12345", ContentType: "text/plain", StatusCode: http.StatusOK},
		"/s2": {Body: "abcde", ContentType: "text/plain", StatusCode: http.StatusOK},
		"/s3": {Body: "fghij", ContentType: "text/plain", StatusCode: http.StatusOK},
	}
	server := startMockServer(t, mockResponses)
	fetcher := newTestFetcher(t, server.URL)

	urls := []string{server.URL + "/s1", server.URL + "/s2", server.URL + "/s3"}
	maxLength := 50 // Much more than total (15)
	raw := true

	resp, err := fetcher.FetchMultiple(urls, maxLength, raw)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Errors)
	require.Len(t, resp.Responses, 3)

	// Initial allocation: 50 / 3 = 16 per URL
	// /s1: uses 5 (fits)
	// /s2: uses 5 (fits)
	// /s3: uses 5 (fits)
	// Total used initially: 5 + 5 + 5 = 15
	// Remaining: 50 - 15 = 35
	// Beneficiaries: 0 (no one was truncated)
	// No reallocation needed.

	r1, _ := resp.Responses[urls[0]]
	r2, _ := resp.Responses[urls[1]]
	r3, _ := resp.Responses[urls[2]]

	assert.Equal(t, "12345", r1.Content)
	assert.Equal(t, "abcde", r2.Content)
	assert.Equal(t, "fghij", r3.Content)

	totalLen := len(r1.Content) + len(r2.Content) + len(r3.Content)
	assert.Equal(t, 15, totalLen) // Total is less than maxLength
}

func TestHTTPFetcher_FetchMultiple_PartialFailure(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/ok":    {Body: "This is okay", ContentType: "text/plain", StatusCode: http.StatusOK},
		"/error": {Body: "Server Error", ContentType: "text/plain", StatusCode: http.StatusInternalServerError},
		// /notfound will implicitly 404
	}
	server := startMockServer(t, mockResponses)
	fetcher := newTestFetcher(t, server.URL)

	urls := []string{
		server.URL + "/ok",
		server.URL + "/error", // This will fetch but have a non-200 status
		server.URL + "/notfound",
		"http://localhost:9998/down", // This will cause a connection error
	}
	maxLength := 100
	raw := true

	resp, err := fetcher.FetchMultiple(urls, maxLength, raw)

	require.NoError(t, err) // FetchMultiple itself shouldn't error on partial failures
	require.NotNil(t, resp)

	// Check successful response
	require.Contains(t, resp.Responses, urls[0])
	rOk, _ := resp.Responses[urls[0]]
	assert.Equal(t, "This is okay", rOk.Content)
	assert.Equal(t, http.StatusOK, rOk.StatusCode)

	// Check response with server error status
	require.Contains(t, resp.Responses, urls[1])
	rErrStatus, _ := resp.Responses[urls[1]]
	assert.Equal(t, "Server Error", rErrStatus.Content) // Content is still fetched
	assert.Equal(t, http.StatusInternalServerError, rErrStatus.StatusCode)

	// Check response for 404
	require.Contains(t, resp.Responses, urls[2])
	rNotFound, _ := resp.Responses[urls[2]]
	assert.Equal(t, "Not Found", rNotFound.Content) // Content from mock 404 handler
	assert.Equal(t, http.StatusNotFound, rNotFound.StatusCode)

	// Ensure only the successful/fetched ones are in Responses
	assert.Len(t, resp.Responses, 3)
	// Ensure only the connection error is in Errors
	assert.Len(t, resp.Errors, 1)
}

func TestHTTPFetcher_FetchMultiple_AllFailures(t *testing.T) {
	// No server started or mock responses defined that match
	fetcher := newTestFetcher(t, "http://localhost:9997")

	urls := []string{
		"http://localhost:9997/fail1",
		"http://localhost:9997/fail2",
	}
	maxLength := 100
	raw := true

	resp, err := fetcher.FetchMultiple(urls, maxLength, raw)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Responses) // No successful responses
}

func TestHTTPFetcher_FetchMultiple_DefaultMaxLength(t *testing.T) {
	// Use a long body to test default limit
	longBody := string(make([]byte, 2000)) // 2000 bytes
	mockResponses := map[string]mockResponse{
		"/long1": {Body: longBody, ContentType: "text/plain", StatusCode: http.StatusOK},
		"/long2": {Body: longBody, ContentType: "text/plain", StatusCode: http.StatusOK},
	}
	server := startMockServer(t, mockResponses)

	// Create fetcher with a specific default max length
	cfg := &config.Config{
		Fetch: struct {
			Timeout          int    `yaml:"timeout" default:"10" env:"FETCH_TIMEOUT"`
			UserAgent        string `yaml:"user_agent" default:"mcp-fetch/1.0" env:"FETCH_USER_AGENT"`
			MaxURLs          int    `yaml:"max_urls" default:"20" env:"FETCH_MAX_URLS"`
			MaxWorkers       int    `yaml:"max_workers" default:"20" env:"FETCH_MAX_WORKERS"`
			DefaultMaxLength int    `yaml:"default_max_length" default:"5000" env:"FETCH_DEFAULT_MAX_LENGTH"`
		}{
			Timeout:          5,
			UserAgent:        "test-agent/1.0",
			MaxWorkers:       5,
			DefaultMaxLength: 1500, // Set a default
		},
	}
	fetcher, err := NewHTTPFetcher(cfg)
	require.NoError(t, err)

	urls := []string{server.URL + "/long1", server.URL + "/long2"}
	maxLength := 0 // Trigger default max length usage
	raw := true

	resp, err := fetcher.FetchMultiple(urls, maxLength, raw)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Errors)
	require.Len(t, resp.Responses, 2)

	// Default MaxLength = 1500
	// Initial allocation: 1500 / 2 = 750 per URL
	// Both /long1 and /long2 have 2000 chars, so both are truncated to 750.
	// Both are beneficiaries.
	// Total used initially: 750 + 750 = 1500
	// Remaining: 1500 - 1500 = 0
	// No reallocation possible.
	// Final lengths: 750, 750

	r1, _ := resp.Responses[urls[0]]
	r2, _ := resp.Responses[urls[1]]

	assert.Len(t, r1.Content, 750)
	assert.Len(t, r2.Content, 750)

	totalLen := len(r1.Content) + len(r2.Content)
	assert.Equal(t, cfg.Fetch.DefaultMaxLength, totalLen, "Total length should match default maxLength")
}

func TestHTTPFetcher_FetchMultiple_HTMLProcessing(t *testing.T) {
	mockResponses := map[string]mockResponse{
		"/html1": {Body: "<html><title>T1</title><body>C1</body></html>", ContentType: "text/html", StatusCode: http.StatusOK},
		"/html2": {Body: "<html><title>T2</title><body>C2</body></html>", ContentType: "text/html", StatusCode: http.StatusOK},
	}
	server := startMockServer(t, mockResponses)
	fetcher := newTestFetcher(t, server.URL)

	urls := []string{server.URL + "/html1", server.URL + "/html2"}
	maxLength := 200
	raw := false // Enable HTML processing

	resp, err := fetcher.FetchMultiple(urls, maxLength, raw)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Errors)
	require.Len(t, resp.Responses, 2)

	r1, ok1 := resp.Responses[urls[0]]
	require.True(t, ok1)
	assert.NotEqual(t, mockResponses["/html1"].Body, r1.Content) // Check it's processed
	assert.Contains(t, r1.Content, "T1")                         // Check title
	// assert.Contains(t, r1.Content, "C1")                      // Removed: Avoid testing readability details

	r2, ok2 := resp.Responses[urls[1]]
	require.True(t, ok2)
	assert.NotEqual(t, mockResponses["/html2"].Body, r2.Content) // Check it's processed
	assert.Contains(t, r2.Content, "T2")                         // Check title
	// assert.Contains(t, r2.Content, "C2")                      // Removed: Avoid testing readability details

	totalLen := len(r1.Content) + len(r2.Content)
	assert.LessOrEqual(t, totalLen, maxLength)
}
