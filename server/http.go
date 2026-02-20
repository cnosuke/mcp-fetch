package server

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/cnosuke/mcp-fetch/config"
	ierrors "github.com/cnosuke/mcp-fetch/internal/errors"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// RunHTTP - Execute the MCP server with Streamable HTTP transport
func RunHTTP(cfg *config.Config, name string, version string, revision string) error {
	zap.S().Infow("starting MCP Fetch Server with Streamable HTTP transport")

	mcpSrv, err := createMCPServer(cfg, name, version, revision)
	if err != nil {
		return err
	}

	var opts []mcpserver.StreamableHTTPOption
	if cfg.HTTP.HeartbeatSeconds > 0 {
		opts = append(opts, mcpserver.WithHeartbeatInterval(time.Duration(cfg.HTTP.HeartbeatSeconds)*time.Second))
	}

	httpHandler := mcpserver.NewStreamableHTTPServer(mcpSrv, opts...)

	var handler http.Handler = httpHandler
	handler = withAuthMiddleware(handler, cfg.HTTP.AuthToken)
	handler = withOriginFilter(handler, cfg.HTTP.AllowedOrigins)

	mux := http.NewServeMux()
	mux.Handle(cfg.HTTP.EndpointPath, handler)
	mux.HandleFunc("/health", handleHealth)

	srv := &http.Server{
		Addr:              cfg.HTTP.Binding,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		zap.S().Infow("starting Streamable HTTP server",
			"binding", cfg.HTTP.Binding,
			"endpoint", cfg.HTTP.EndpointPath,
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- ierrors.Wrap(err, "HTTP server error")
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		zap.S().Infow("shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return ierrors.Wrap(err, "failed to shutdown HTTP server")
		}
	}

	return nil
}
