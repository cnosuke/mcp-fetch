package server

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	"github.com/cnosuke/mcp-fetch/config"
	"github.com/cnosuke/mcp-fetch/fetcher"
	ierrors "github.com/cnosuke/mcp-fetch/internal/errors"
)

// RunStdio - Execute the MCP server with STDIO transport
func RunStdio(cfg *config.Config, name string, version string, revision string) error {
	zap.S().Infow("starting MCP Fetch Server with STDIO transport")

	mcpServer, err := createMCPServer(cfg, name, version, revision)
	if err != nil {
		return err
	}

	zap.S().Infow("starting MCP server with STDIO")
	err = server.ServeStdio(mcpServer)
	if err != nil {
		zap.S().Errorw("failed to start STDIO server", "error", err)
		return ierrors.Wrap(err, "failed to start STDIO server")
	}

	zap.S().Infow("STDIO server shutting down")
	return nil
}

// createMCPServer - Create MCP server instance with common configuration
func createMCPServer(cfg *config.Config, name string, version string, revision string) (*server.MCPServer, error) {
	// Format version string with revision if available
	versionString := version
	if revision != "" && revision != "xxx" {
		versionString = versionString + " (" + revision + ")"
	}

	// Create Fetcher
	zap.S().Debugw("creating HTTP Fetcher")
	httpFetcher, err := fetcher.NewHTTPFetcher(&fetcher.Config{
		Timeout:          cfg.Fetch.Timeout,
		UserAgent:        cfg.Fetch.UserAgent,
		MaxURLs:          cfg.Fetch.MaxURLs,
		MaxWorkers:       cfg.Fetch.MaxWorkers,
		DefaultMaxLength: cfg.Fetch.DefaultMaxLength,
	})
	if err != nil {
		zap.S().Errorw("failed to create HTTP Fetcher", "error", err)
		return nil, err
	}

	// Create custom hooks for error handling
	hooks := &server.Hooks{}
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		zap.S().Errorw("MCP error occurred",
			"id", id,
			"method", method,
			"error", err,
		)
	})

	// Create MCP server with server name and version
	zap.S().Debugw("creating MCP server",
		"name", name,
		"version", versionString,
	)
	mcpServer := server.NewMCPServer(
		name,
		versionString,
		server.WithHooks(hooks),
	)

	// Register all tools
	zap.S().Debugw("registering tools")
	if err := RegisterAllTools(mcpServer, httpFetcher, cfg.Fetch.MaxURLs, cfg); err != nil {
		zap.S().Errorw("failed to register tools", "error", err)
		return nil, err
	}

	return mcpServer, nil
}
