package tools

import (
	mcp "github.com/metoro-io/mcp-golang"
)

// RegisterAllTools - Register all tools with the server
func RegisterAllTools(mcpServer *mcp.Server, fetcher Fetcher) error {
	// Register fetch tool
	if err := RegisterFetchTool(mcpServer, fetcher); err != nil {
		return err
	}

	return nil
}
