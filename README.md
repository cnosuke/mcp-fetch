# MCP Fetch Server

MCP Fetch Server is a Go-based MCP server implementation that provides URL fetching functionality, allowing MCP clients (e.g., Claude Desktop) to fetch content from URLs with automatic format conversion.

## Features

- MCP Compliance: Provides a JSON‐RPC based interface for tool execution according to the MCP specification.
- URL Fetching: Fetch content from URLs with automatic format conversion.
- Markdown Conversion: Automatically converts HTML content to Markdown for better readability.
- Readability Enhancement: Uses go-readability to extract and clean up the main content from HTML pages, preserving titles and important content while removing clutter.
- Smart Content Processing: Includes title and excerpt information in the converted output, with fallback processing if the primary conversion fails.
- Content Type Detection: Returns appropriate content based on the Content-Type header.
- Content Control: Supports content length limitation, offset, and raw content retrieval.

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
- `max_length` (integer, optional): Maximum number of characters to return, distributed equally among all URLs (default: 5000)
- `raw` (boolean, optional): Get raw content without markdown conversion (default: false)

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

## Examples

### 単一URLへの指定オプション:

```json
{
  "url": "https://example.com",
  "max_length": 1000,
  "start_index": 500,
  "raw": false
}
```

### 複数URLへの均等配分:

```json
{
  "urls": ["https://example1.com", "https://example2.com", "https://example3.com"],
  "max_length": 9000,
  "raw": false
}
```

この例では、各URLに3000文字ずつ割り当てられます。

## Contributing

Contributions are welcome! Please fork the repository and submit pull requests for improvements or bug fixes. For major changes, open an issue first to discuss your ideas.

## License

This project is licensed under the MIT License.

Author: cnosuke ( x.com/cnosuke )
