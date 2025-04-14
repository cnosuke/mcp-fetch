package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cnosuke/mcp-fetch/config"
	"github.com/cnosuke/mcp-fetch/fetcher"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// FetchArgs - Arguments for fetch tool (kept for test compatibility)
type FetchArgs struct {
	URL        string `json:"url" jsonschema:"description=URL to fetch,required=true"`
	MaxLength  int    `json:"max_length,omitempty" jsonschema:"description=Maximum number of characters to return"`
	StartIndex int    `json:"start_index,omitempty" jsonschema:"description=Start content from this character index"`
	Raw        bool   `json:"raw,omitempty" jsonschema:"description=Get raw content without markdown conversion"`
}

// RegisterFetchTool - Register the fetch tool
func RegisterFetchTool(mcpServer *server.MCPServer, f fetcher.Fetcher, cfg *config.Config) error {
	zap.S().Debugw("registering fetch tool")

	// Define the tool
	tool := mcp.NewTool("fetch",
		mcp.WithDescription(fmt.Sprintf("Fetches a URL from the internet and extracts its contents as markdown. Default max_length is %d.", cfg.Fetch.DefaultMaxLength)),
		mcp.WithString("url",
			mcp.Description("URL to fetch"),
			mcp.Required(),
		),
		mcp.WithNumber("max_length",
			mcp.Description("Maximum number of characters to return"),
		),
		mcp.WithNumber("start_index",
			mcp.Description("Start content from this character index"),
		),
		mcp.WithBoolean("raw",
			mcp.Description("Get raw content without markdown conversion"),
		),
	)

	// Register the tool handler
	mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		url, _ := request.Params.Arguments["url"].(string)
		
		var maxLength int
		if maxLengthVal, ok := request.Params.Arguments["max_length"].(float64); ok {
			maxLength = int(maxLengthVal)
		}
		
		var startIndex int
		if startIndexVal, ok := request.Params.Arguments["start_index"].(float64); ok {
			startIndex = int(startIndexVal)
		}
		
		var raw bool
		if rawVal, ok := request.Params.Arguments["raw"].(bool); ok {
			raw = rawVal
		}

		zap.S().Infow("executing fetch",
			"url", url,
			"max_length", maxLength,
			"start_index", startIndex,
			"raw", raw)

		// Validate URL
		if url == "" {
			return mcp.NewToolResultError("URL is required"), nil
		}

		// Set default values
		if maxLength <= 0 {
			maxLength = cfg.Fetch.DefaultMaxLength
		}

		// Fetch URL with parameters using the Fetcher interface
		response, err := f.Fetch(url, maxLength, startIndex, raw)
		if err != nil {
			zap.S().Errorw("failed to fetch URL",
				"url", url,
				"error", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to fetch URL: %s", err.Error())), nil
		}

		// Convert response to JSON
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			zap.S().Errorw("failed to marshal response to JSON",
				"error", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response to JSON: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	})

	return nil
}
