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
	URL        string `json:"url" jsonschema:"description=URL to fetch,required=true"`
	MaxLength  int    `json:"max_length,omitempty" jsonschema:"description=Maximum number of characters to return"`
	StartIndex int    `json:"start_index,omitempty" jsonschema:"description=Start content from this character index"`
	Raw        bool   `json:"raw,omitempty" jsonschema:"description=Get raw content without markdown conversion"`
}

// Fetcher defines the interface for single URL fetching
type Fetcher interface {
	FetchURL(url string, maxLength int, startIndex int, raw bool) (*types.FetchResponse, error)
}

// MultiFetcher defines the interface for multiple URL fetching
type MultiFetcher interface {
	Fetcher
	FetchMultipleURLs(urls []string, maxLength int, raw bool) (*types.MultipleFetchResponse, error)
}

// RegisterFetchTool - Register the fetch tool
func RegisterFetchTool(mcpServer *mcp.Server, fetcher Fetcher) error {
	zap.S().Debugw("registering fetch tool")
	err := mcpServer.RegisterTool("fetch", "Fetches a URL from the internet and extracts its contents as markdown",
		func(args FetchArgs) (*mcp.ToolResponse, error) {
			zap.S().Infow("executing fetch",
				"url", args.URL,
				"max_length", args.MaxLength,
				"start_index", args.StartIndex,
				"raw", args.Raw)

			// Validate URL
			if args.URL == "" {
				return nil, errors.New("URL is required")
			}

			// Set default values
			maxLength := 5000
			if args.MaxLength > 0 {
				maxLength = args.MaxLength
			}

			startIndex := 0
			if args.StartIndex > 0 {
				startIndex = args.StartIndex
			}

			// Fetch URL with parameters
			response, err := fetcher.FetchURL(args.URL, maxLength, startIndex, args.Raw)
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
