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

// FetchMultipleArgs - Arguments for fetch_multiple tool (kept for test compatibility)
type FetchMultipleArgs struct {
	URLs      []string `json:"urls" jsonschema:"description=URLs to fetch (maximum depends on config),maxItems=100"`
	MaxLength int      `json:"max_length,omitempty" jsonschema:"description=Maximum total number of characters to return across all URLs combined"`
	Raw       bool     `json:"raw,omitempty" jsonschema:"description=Get raw content without markdown conversion"`
}

// RegisterFetchMultipleTool - Register the fetch_multiple tool
func RegisterFetchMultipleTool(mcpServer *server.MCPServer, f fetcher.Fetcher, maxURLs int, cfg *config.Config) error {
	zap.S().Debugw("registering fetch_multiple tool", "max_urls", maxURLs)

	// Define the tool
	tool := mcp.NewTool("fetch_multiple",
		mcp.WithDescription(fmt.Sprintf("Fetch content from multiple URLs (max %d). Default max_length is %d.", maxURLs, cfg.Fetch.DefaultMaxLength)),
		mcp.WithArray("urls",
			mcp.Description(fmt.Sprintf("URLs to fetch (maximum %d)", maxURLs)),
			mcp.Required(),
			mcp.MaxItems(100),
		),
		mcp.WithNumber("max_length",
			mcp.Description("Maximum total number of characters to return across all URLs combined"),
		),
		mcp.WithBoolean("raw",
			mcp.Description("Get raw content without markdown conversion"),
		),
	)

	// Register the tool handler
	mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		var urls []string
		if urlsArray, ok := request.Params.Arguments["urls"].([]interface{}); ok {
			for _, u := range urlsArray {
				if urlStr, ok := u.(string); ok {
					urls = append(urls, urlStr)
				}
			}
		}

		var maxLength int
		if maxLengthVal, ok := request.Params.Arguments["max_length"].(float64); ok {
			maxLength = int(maxLengthVal)
		}

		var raw bool
		if rawVal, ok := request.Params.Arguments["raw"].(bool); ok {
			raw = rawVal
		}

		// Log the request
		zap.S().Debugw("executing fetch_multiple",
			"urls_count", len(urls),
			"max_length", maxLength,
			"raw", raw)

		// Validate URLs count
		if len(urls) == 0 {
			return mcp.NewToolResultError("at least one URL is required"), nil
		}

		if len(urls) > maxURLs {
			return mcp.NewToolResultError(fmt.Sprintf("too many URLs: maximum allowed is %d", maxURLs)), nil
		}

		// Set default values
		if maxLength <= 0 {
			maxLength = cfg.Fetch.DefaultMaxLength
		}

		// Fetch URLs with parameters using the Fetcher interface
		response, err := f.FetchMultiple(urls, maxLength, raw)
		if err != nil {
			zap.S().Errorw("failed to fetch multiple URLs",
				"error", err)
			return mcp.NewToolResultError(fmt.Sprintf("failed to fetch multiple URLs: %s", err.Error())), nil
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
