package types

// FetchResponse - Response from fetch operation
type FetchResponse struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
	StatusCode  int    `json:"status_code"`
}
