# MCP Fetch Server

MCP Fetch Server is a Go-based MCP server implementation that provides URL fetching functionality, allowing MCP clients (e.g., Claude Desktop) to fetch content from URLs with automatic format conversion.

## Features

- MCP Compliance: Provides a JSON‐RPC based interface for tool execution according to the MCP specification.
- URL Fetching: Fetch content from URLs with automatic format conversion.
- Markdown Conversion: Automatically converts HTML content to Markdown for better readability.
- Content Type Detection: Returns appropriate content based on the Content-Type header.

## Requirements

- Go 1.24 or later

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
```

Note: Configuration parameters can also be injected via environment variables:

- `LOG_PATH`: Path to log file
- `DEBUG`: Enable debug mode (true/false)
- `FETCH_TIMEOUT`: Override the fetch timeout in seconds
- `FETCH_USER_AGENT`: Override the user agent string
- `FETCH_MAX_URLS`: Override the maximum number of URLs that can be processed in a single request
- `FETCH_MAX_WORKERS`: Override the maximum number of worker goroutines for parallel processing

## Logging

Logging behavior is controlled through configuration:

- If `log` is set in the config file, logs will be written to the specified file
- If `log` is empty, no logs will be produced
- Set `debug: true` for more verbose logging

## MCP Server Usage

MCP clients interact with the server by sending JSON‐RPC requests to execute various tools. The following MCP tools are supported:

- `fetch`: Fetches content from a URL, with automatic format conversion based on content type.
- `fetch_multiple`: Fetches content from multiple URLs in parallel (up to the configured limit), with automatic format conversion.

### Using with Claude Desktop

To integrate with Claude Desktop, add an entry to your `claude_desktop_config.json` file:

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
        "FETCH_MAX_WORKERS": "20"
      }
    }
  }
}
```

This configuration registers the MCP Fetch Server with Claude Desktop, ensuring that all logs are directed to the specified log file.

## Contributing

Contributions are welcome! Please fork the repository and submit pull requests for improvements or bug fixes. For major changes, open an issue first to discuss your ideas.

## License

This project is licensed under the MIT License.

Author: cnosuke ( x.com/cnosuke )
