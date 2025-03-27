package tools

import (
	"github.com/cnosuke/mcp-fetch/config"
	mcp "github.com/metoro-io/mcp-golang"
)

// RegisterAllTools - Register all tools with the server
func RegisterAllTools(mcpServer *mcp.Server, fetcher MultiFetcher, maxURLs int, cfg *config.Config) error {
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
