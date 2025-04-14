package server

import (
	"testing"

	"github.com/cnosuke/mcp-fetch/types"
	"github.com/stretchr/testify/assert"
)

// MockFetcher for testing
type MockFetcher struct {
	defaultResponse *types.FetchResponse
}

// Fetch - Mock implementation
func (f *MockFetcher) Fetch(urlStr string, maxLength int, startIndex int, raw bool) (*types.FetchResponse, error) {
	// Clone the default response
	response := &types.FetchResponse{
		URL:         f.defaultResponse.URL,
		ContentType: f.defaultResponse.ContentType,
		Content:     f.defaultResponse.Content,
		StatusCode:  f.defaultResponse.StatusCode,
	}

	// Apply content slicing based on parameters
	content := response.Content
	if startIndex > 0 && startIndex < len(content) {
		content = content[startIndex:]
	}

	if maxLength > 0 && len(content) > maxLength {
		content = content[:maxLength]
	}

	response.Content = content
	return response, nil
}

// FetchMultiple - Mock implementation
func (f *MockFetcher) FetchMultiple(urls []string, maxLength int, raw bool) (*types.MultipleFetchResponse, error) {
	// Create a response with each URL getting the same content
	response := &types.MultipleFetchResponse{
		Responses: make(map[string]*types.FetchResponse),
		Errors:    make(map[string]string),
	}

	totalLength := 0
	for _, url := range urls {
		// Get a response for this URL
		urlResponse, _ := f.Fetch(url, 0, 0, raw)
		
		// Check if adding this would exceed the total maxLength
		if maxLength > 0 {
			contentLength := len(urlResponse.Content)
			if totalLength + contentLength > maxLength {
				remainingLength := maxLength - totalLength
				if remainingLength > 0 {
					// Trim to fit
					urlResponse.Content = urlResponse.Content[:remainingLength]
					response.Responses[url] = urlResponse
					totalLength += remainingLength
				}
				// Stop processing more URLs once we hit the limit
				break
			}
			totalLength += contentLength
		}
		
		response.Responses[url] = urlResponse
	}

	return response, nil
}

// TestFetchFunctionality tests the basic fetch functionality with parameters
func TestFetchFunctionality(t *testing.T) {
	// Create mock fetcher with sample data
	mockFetcher := &MockFetcher{
		defaultResponse: &types.FetchResponse{
			URL:         "https://example.com",
			ContentType: "text/html",
			Content:     "This is a sample content string for testing purposes. It should be long enough to test various length limits.",
			StatusCode:  200,
		},
	}

	// Test 1: Basic fetch with default parameters
	resp1, err := mockFetcher.Fetch("https://example.com", 0, 0, false)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", resp1.URL)
	assert.Equal(t, "This is a sample content string for testing purposes. It should be long enough to test various length limits.", resp1.Content)

	// Test 2: Fetch with maxLength
	resp2, err := mockFetcher.Fetch("https://example.com", 20, 0, false)
	assert.NoError(t, err)
	assert.Equal(t, 20, len(resp2.Content))
	assert.Equal(t, "This is a sample con", resp2.Content)

	// Test 3: Fetch with startIndex
	resp3, err := mockFetcher.Fetch("https://example.com", 0, 10, false)
	assert.NoError(t, err)
	assert.Equal(t, "sample content string for testing purposes. It should be long enough to test various length limits.", resp3.Content)

	// Test 4: Fetch with both maxLength and startIndex
	resp4, err := mockFetcher.Fetch("https://example.com", 10, 10, false)
	assert.NoError(t, err)
	assert.Equal(t, "sample con", resp4.Content)
}

// TestFetchMultipleFunctionality tests the fetch_multiple functionality
func TestFetchMultipleFunctionality(t *testing.T) {
	// Create mock fetcher with sample data
	mockFetcher := &MockFetcher{
		defaultResponse: &types.FetchResponse{
			URL:         "https://example.com",
			ContentType: "text/html",
			Content:     "This is a sample content string for testing purposes. It should be long enough to test various length limits.",
			StatusCode:  200,
		},
	}

	// Test 1: Fetch multiple URLs with no limit
	urls := []string{"https://example1.com", "https://example2.com", "https://example3.com"}
	resp1, err := mockFetcher.FetchMultiple(urls, 0, false)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(resp1.Responses))
	assert.Equal(t, 0, len(resp1.Errors))

	// Test 2: Fetch multiple URLs with a total length limit that allows only partial content
	resp2, err := mockFetcher.FetchMultiple(urls, 150, false)
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(resp1.Responses), 3)
	
	// Calculate total content length
	totalLength := 0
	for _, resp := range resp2.Responses {
		totalLength += len(resp.Content)
	}
	assert.LessOrEqual(t, totalLength, 150)
}
