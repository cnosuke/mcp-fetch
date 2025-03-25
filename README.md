# MCP Fetch Server

MCP Fetch Server is a Go-based MCP server implementation that provides URL fetching functionality, allowing MCP clients (e.g., Claude Desktop) to fetch content from URLs with automatic format conversion.

## Features

* MCP Compliance: Provides a JSON‐RPC based interface for tool execution according to the MCP specification.
* URL Fetching: Fetch content from URLs with automatic format conversion.
* Markdown Conversion: Automatically converts HTML content to Markdown for better readability.
* Content Type Detection: Returns appropriate content based on the Content-Type header.

## Requirements

* Go 1.24 or later

## Configuration

The server is configured via a YAML file (default: config.yml). For example:

```yaml
fetch:
  timeout: 30
  user_agent: "mcp-fetch/1.0"
```

Note: Configuration parameters can also be injected via environment variables:

* `FETCH_TIMEOUT`: Override the fetch timeout in seconds
* `FETCH_USER_AGENT`: Override the user agent string

## Logging

Adjust logging behavior using the following command-line flags:

* `--no-logs`: Suppress non-critical logs.
* `--log`: Specify a file path to write logs.

Important: When using the MCP server with a stdio transport, logging must not be directed to standard output because it would interfere with the MCP protocol communication. Therefore, you should always use `--no-logs` along with `--log` to ensure that all logs are written exclusively to a log file.

## MCP Server Usage

MCP clients interact with the server by sending JSON‐RPC requests to execute various tools. The following MCP tools are supported:

* `fetch`: Fetches content from a URL, with automatic format conversion based on content type.

### Using with Claude Desktop

To integrate with Claude Desktop, add an entry to your `claude_desktop_config.json` file. Because MCP uses stdio for communication, you must redirect logs away from stdio by using the `--no-logs` and `--log` flags:

```json
{
  "mcpServers": {
    "fetch": {
      "command": "./bin/mcp-fetch",
      "args": ["server", "--no-logs", "--log", "mcp-fetch.log"],
      "env": {
        "FETCH_TIMEOUT": "30",
        "FETCH_USER_AGENT": "mcp-fetch/1.0"
      }
    }
  }
}
```

This configuration registers the MCP Fetch Server with Claude Desktop, ensuring that all logs are directed to the specified log file rather than interfering with the MCP protocol messages transmitted over stdio.

## Contributing

Contributions are welcome! Please fork the repository and submit pull requests for improvements or bug fixes. For major changes, open an issue first to discuss your ideas.

## License

This project is licensed under the MIT License.

Author: cnosuke ( x.com/cnosuke )
