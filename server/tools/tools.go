package tools

import (
	mcp "github.com/metoro-io/mcp-golang"
)

// RegisterAllTools - Register all tools with the server
func RegisterAllTools(mcpServer *mcp.Server, fetcher MultiFetcher, maxURLs int) error {
	// Register fetch tool
	if err := RegisterFetchTool(mcpServer, fetcher); err != nil {
		return err
	}

	// Register fetch_multiple tool
	if err := RegisterFetchMultipleTool(mcpServer, fetcher, maxURLs); err != nil {
		return err
	}

	return nil
}
