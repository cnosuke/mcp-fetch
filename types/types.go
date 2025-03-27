package types

// FetchResponse - Response from fetch operation
type FetchResponse struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
	StatusCode  int    `json:"status_code"`
}

// MultipleFetchResponse - Multiple URLs fetch response
type MultipleFetchResponse struct {
	Responses map[string]*FetchResponse `json:"responses"` // Map of responses with URLs as keys
	Errors    map[string]string         `json:"errors"`    // Map of error messages with failed URLs as keys
}
