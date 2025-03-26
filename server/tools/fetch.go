package tools

import (
	"encoding/json"

	"github.com/cnosuke/mcp-fetch/types"
	"github.com/cockroachdb/errors"
	mcp "github.com/metoro-io/mcp-golang"
	"go.uber.org/zap"
)

// FetchArgs - Arguments for fetch tool
type FetchArgs struct {
	URL string `json:"url" jsonschema:"description=URL to fetch,required=true"`
}

// Fetcher defines the interface for single URL fetching
type Fetcher interface {
	FetchURL(url string) (*types.FetchResponse, error)
}

// MultiFetcher defines the interface for multiple URL fetching
type MultiFetcher interface {
	Fetcher
	FetchMultipleURLs(urls []string) (*types.MultipleFetchResponse, error)
}

// RegisterFetchTool - Register the fetch tool
func RegisterFetchTool(mcpServer *mcp.Server, fetcher Fetcher) error {
	zap.S().Debugw("registering fetch tool")
	err := mcpServer.RegisterTool("fetch", "Fetch content from a URL with automatic format conversion",
		func(args FetchArgs) (*mcp.ToolResponse, error) {
			zap.S().Infow("executing fetch",
				"url", args.URL)

			// Validate URL
			if args.URL == "" {
				return nil, errors.New("URL is required")
			}

			// Fetch URL
			response, err := fetcher.FetchURL(args.URL)
			if err != nil {
				zap.S().Errorw("failed to fetch URL",
					"url", args.URL,
					"error", err)
				return nil, errors.Wrap(err, "failed to fetch URL")
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
		zap.S().Errorw("failed to register fetch tool", "error", err)
		return errors.Wrap(err, "failed to register fetch tool")
	}

	return nil
}
