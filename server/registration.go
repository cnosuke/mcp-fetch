package server

import (
	"github.com/cnosuke/mcp-fetch/config"
	"github.com/cnosuke/mcp-fetch/fetcher"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterAllTools - Register all tools with the server
func RegisterAllTools(mcpServer *server.MCPServer, f fetcher.Fetcher, maxURLs int, cfg *config.Config) error {
	// Register fetch tool
	if err := RegisterFetchTool(mcpServer, f, cfg); err != nil {
		return err
	}

	// Register fetch_multiple tool
	if err := RegisterFetchMultipleTool(mcpServer, f, maxURLs, cfg); err != nil {
		return err
	}

	return nil
}
