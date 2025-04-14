# MCP Fetch Server

MCP Fetch Server is a Go-based MCP server implementation that provides URL fetching functionality, allowing MCP clients (e.g., Claude Desktop) to fetch content from URLs with automatic format conversion.

MCP Fetch Server optimizes content for AI models by extracting only the meaningful content from web pages and converting it to token-efficient Markdown format, significantly reducing token usage while preserving the essential information.

## Features

- MCP Compliance: Provides a JSON‐RPC based interface for tool execution according to the MCP specification.
- URL Fetching: Fetch content from URLs with automatic format conversion.
- Token Efficiency: Extracts only the main content from web pages and removes clutter like ads, navigation, and irrelevant elements, significantly reducing token usage with AI models.
- Markdown Conversion: Automatically converts HTML content to Markdown for better readability and further token optimization.
- Readability Enhancement: Uses go-readability to preserve titles and important content while removing clutter.
- Smart Content Processing: Includes title and author information in the converted output, with fallback processing if the primary conversion fails.
- Content Type Detection: Returns appropriate content based on the Content-Type header.
- Content Control: Supports content length limitation, offset, and raw content retrieval.

## Requirements

- Docker (recommended)

For local development:

- Go 1.24 or later

## Using with Docker (Recommended)

```bash
docker pull cnosuke/mcp-fetch:latest

docker run -i --rm cnosuke/mcp-fetch:latest
```

### Using with Claude Desktop (Docker)

To integrate with Claude Desktop using Docker, add an entry to your `claude_desktop_config.json` file:

```json
{
  "mcpServers": {
    "fetch": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "cnosuke/mcp-fetch:latest"]
    }
  }
}
```

## Building and Running (Go Binary)

Alternatively, you can build and run the Go binary directly:

```bash
# Build the server
make bin/mcp-fetch

# Run the server
./bin/mcp-fetch server --config=config.yml
```

### Using with Claude Desktop (Go Binary)

To integrate with Claude Desktop using the Go binary, add an entry to your `claude_desktop_config.json` file:

```json
{
  "mcpServers": {
    "fetch": {
      "command": "./bin/mcp-fetch",
      "args": ["server"],
      "env": {
        "LOG_PATH": "mcp-fetch.log",
        "DEBUG": "false",
        "FETCH_TIMEOUT": "10",
        "FETCH_USER_AGENT": "mcp-fetch/1.0",
        "FETCH_MAX_URLS": "20",
        "FETCH_MAX_WORKERS": "20",
        "FETCH_DEFAULT_MAX_LENGTH": "5000"
      }
    }
  }
}
```

## Configuration

The server is configured via a YAML file (default: config.yml). For example:

```yaml
log: 'path/to/mcp-fetch.log' # Log file path, if empty no log will be produced
debug: false # Enable debug mode for verbose logging

fetch:
  timeout: 10
  user_agent: 'mcp-fetch/1.0'
  max_urls: 20
  max_workers: 20
  default_max_length: 5000
```

Note: Configuration parameters can also be injected via environment variables:

- `LOG_PATH`: Path to log file
- `DEBUG`: Enable debug mode (true/false)
- `FETCH_TIMEOUT`: Override the fetch timeout in seconds
- `FETCH_USER_AGENT`: Override the user agent string
- `FETCH_MAX_URLS`: Override the maximum number of URLs that can be processed in a single request
- `FETCH_MAX_WORKERS`: Override the maximum number of worker goroutines for parallel processing
- `FETCH_DEFAULT_MAX_LENGTH`: Override the default maximum length for content fetching (default: 5000)

## Logging

Logging behavior is controlled through configuration:

- If `log` is set in the config file, logs will be written to the specified file
- If `log` is empty, no logs will be produced
- Set `debug: true` for more verbose logging

## MCP Server Usage

MCP clients interact with the server by sending JSON‐RPC requests to execute various tools. The following MCP tools are supported:

### fetch

Fetches a URL from the internet and extracts its contents as markdown.

Parameters:

- `url` (string, required): URL to fetch
- `max_length` (integer, optional): Maximum number of characters to return (default: 5000)
- `start_index` (integer, optional): Start content from this character index (default: 0)
- `raw` (boolean, optional): Get raw content without markdown conversion (default: false)

### fetch_multiple

Fetches content from multiple URLs in parallel (up to the configured limit), with automatic format conversion.

Parameters:

- `urls` (array of strings, required): URLs to fetch (maximum depends on config)
- `max_length` (integer, optional): Maximum number of characters to return, initially distributed equally among all URLs with unused allocation redistributed (default: 5000)
- `raw` (boolean, optional): Get raw content without markdown conversion (default: false)

## Command-Line Parameters

When starting the server, you can specify various settings:

```bash
./bin/mcp-fetch server [options]
```

Options:

- `--config`, `-c`: Path to the configuration file (default: "config.yml").

## Examples

### Single URL Options:

```json
{
  "url": "https://example.com",
  "max_length": 1000,
  "start_index": 500,
  "raw": false
}
```

### Multiple URLs with Balanced Distribution:

```json
{
  "urls": [
    "https://1.example.com",
    "https://2.example.net",
    "https://3.example.org"
  ],
  "max_length": 9000,
  "raw": false
}
```

In this example, each URL is initially allocated 3000 characters. If example1.com only uses 1500 characters, the remaining 1500 characters will be redistributed to the other URLs.

## Dependencies

This project makes use of several excellent open source libraries:

- [github.com/mackee/go-readability](https://github.com/mackee/go-readability) - A Go implementation of Mozilla's Readability library for extracting main content from web pages and converting to Markdown
- [github.com/mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - Go implementation of the MCP (Message Control Protocol) specification
- [go.uber.org/zap](https://github.com/uber-go/zap) - Blazing fast, structured, leveled logging in Go

## Contributing

Contributions are welcome! Please fork the repository and submit pull requests for improvements or bug fixes. For major changes, open an issue first to discuss your ideas.

## License

This project is licensed under the MIT License.

Author: cnosuke ( x.com/cnosuke )
