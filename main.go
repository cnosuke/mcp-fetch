package main

import (
	"fmt"
	"os"

	"github.com/cnosuke/mcp-fetch/config"
	ierrors "github.com/cnosuke/mcp-fetch/internal/errors"
	"github.com/cnosuke/mcp-fetch/logger"
	"github.com/cnosuke/mcp-fetch/server"
	"github.com/urfave/cli/v2"
)

var (
	// Version and Revision are replaced when building.
	// To set specific version, edit Makefile.
	Version  = "0.0.1"
	Revision = "xxx"

	Name  = "mcp-fetch"
	Usage = "A simple MCP server implementation for fetch URLs"
)

func main() {
	app := cli.NewApp()
	app.Version = fmt.Sprintf("%s (%s)", Version, Revision)
	app.Name = Name
	app.Usage = Usage

	app.Commands = []*cli.Command{
		{
			Name:    "stdioserver",
			Aliases: []string{"stdio", "s"},
			Usage:   "Run MCP server with STDIO transport",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "config",
					Aliases: []string{"c"},
					Value:   "config.yml",
					Usage:   "path to the configuration file",
				},
			},
			Action: func(c *cli.Context) error {
				configPath := c.String("config")

				// Read the configuration file
				cfg, err := config.LoadConfig(configPath)
				if err != nil {
					return ierrors.Wrap(err, "failed to load configuration file")
				}

				// Initialize logger
				if err := logger.InitLogger(cfg.Debug, cfg.Log); err != nil {
					return ierrors.Wrap(err, "failed to initialize logger")
				}
				defer logger.Sync()

				return server.RunStdio(cfg, Name, Version, Revision)
			},
		},
		{
			Name:    "httpserver",
			Aliases: []string{"http"},
			Usage:   "Run MCP server with Streamable HTTP transport",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "config",
					Aliases: []string{"c"},
					Value:   "config.yml",
					Usage:   "path to the configuration file",
				},
			},
			Action: func(c *cli.Context) error {
				configPath := c.String("config")

				// Read the configuration file
				cfg, err := config.LoadConfig(configPath)
				if err != nil {
					return ierrors.Wrap(err, "failed to load configuration file")
				}

				// Initialize logger
				if err := logger.InitLogger(cfg.Debug, cfg.Log); err != nil {
					return ierrors.Wrap(err, "failed to initialize logger")
				}
				defer logger.Sync()

				return server.RunHTTP(cfg, Name, Version, Revision)
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
