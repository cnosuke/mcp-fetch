package tools

import (
	"encoding/json"
	"fmt"

	"github.com/cnosuke/mcp-fetch/config"
	"github.com/cockroachdb/errors"
	mcp "github.com/metoro-io/mcp-golang"
	"go.uber.org/zap"
)

// FetchMultipleArgs - Arguments for fetch_multiple tool
type FetchMultipleArgs struct {
	URLs      []string `json:"urls" jsonschema:"description=URLs to fetch (maximum depends on config),maxItems=100"`
	MaxLength int      `json:"max_length,omitempty" jsonschema:"description=Maximum total number of characters to return across all URLs combined"`
	Raw       bool     `json:"raw,omitempty" jsonschema:"description=Get raw content without markdown conversion"`
}

// FetchMultipleTool - Register the fetch_multiple tool
func RegisterFetchMultipleTool(mcpServer *mcp.Server, fetcher MultiFetcher, maxURLs int, cfg *config.Config) error {
	zap.S().Debugw("registering fetch_multiple tool", "max_urls", maxURLs)
	err := mcpServer.RegisterTool("fetch_multiple", fmt.Sprintf("Fetch content from multiple URLs (max %d)", maxURLs),
		func(args FetchMultipleArgs) (*mcp.ToolResponse, error) {
			// Log the request
			zap.S().Debugw("executing fetch_multiple",
				"urls_count", len(args.URLs),
				"max_length", args.MaxLength,
				"raw", args.Raw)

			// Validate URLs count
			if len(args.URLs) == 0 {
				return nil, errors.New("at least one URL is required")
			}

			if len(args.URLs) > maxURLs {
				return nil, errors.Newf("too many URLs: maximum allowed is %d", maxURLs)
			}

			// Set default values
			maxLength := cfg.Fetch.DefaultMaxLength
			if args.MaxLength > 0 {
				maxLength = args.MaxLength
			}

			// Fetch URLs with parameters
			response, err := fetcher.FetchMultipleURLs(args.URLs, maxLength, args.Raw)
			if err != nil {
				zap.S().Errorw("failed to fetch multiple URLs",
					"error", err)
				return nil, errors.Wrap(err, "failed to fetch multiple URLs")
			}

			// Convert response to JSON
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				zap.S().Errorw("failed to marshal response to JSON",
					"error", err)
				return nil, errors.Wrap(err, "failed to marshal response to JSON")
			}

			return mcp.NewToolResponse(mcp.NewTextContent(string(jsonResponse))), nil
		})

	if err != nil {
		zap.S().Errorw("failed to register fetch_multiple tool", "error", err)
		return errors.Wrap(err, "failed to register fetch_multiple tool")
	}

	return nil
}
