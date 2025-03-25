package server

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown"
	"github.com/cnosuke/mcp-fetch/config"
	"github.com/cnosuke/mcp-fetch/types"
	"github.com/cockroachdb/errors"
	"go.uber.org/zap"
)

// FetchServer - Fetch server structure
type FetchServer struct {
	client    *http.Client
	userAgent string
}

// NewFetchServer - Create a new Fetch server
func NewFetchServer(cfg *config.Config) (*FetchServer, error) {
	zap.S().Infow("creating new Fetch server",
		"timeout", cfg.Fetch.Timeout,
		"user_agent", cfg.Fetch.UserAgent)

	client := &http.Client{
		Timeout: time.Duration(cfg.Fetch.Timeout) * time.Second,
	}

	return &FetchServer{
		client:    client,
		userAgent: cfg.Fetch.UserAgent,
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
