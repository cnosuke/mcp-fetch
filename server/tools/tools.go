package tools

import (
	"github.com/cnosuke/mcp-fetch/config"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterAllTools - Register all tools with the server
func RegisterAllTools(mcpServer *server.MCPServer, fetcher MultiFetcher, maxURLs int, cfg *config.Config) error {
	// Register fetch tool
	if err := RegisterFetchTool(mcpServer, fetcher, cfg); err != nil {
		return err
	}

	// Register fetch_multiple tool
	if err := RegisterFetchMultipleTool(mcpServer, fetcher, maxURLs, cfg); err != nil {
		return err
	}

	return nil
}