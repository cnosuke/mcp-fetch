# MCP Greeting Server

MCP Greeting Server is a Go-based MCP server implementation that provides basic greeting functionality, allowing MCP clients (e.g., Claude Desktop) to generate greeting messages.

## Features

* MCP Compliance: Provides a JSON‐RPC based interface for tool execution according to the MCP specification.
* Greeting Operations: Supports generating greeting messages, with options for personalization.

## Requirements

* Go 1.24 or later

## Configuration

The server is configured via a YAML file (default: config.yml). For example:

```yaml
greeting:
  default_message: "こんにちは！"
```

Note: The default greeting message can also be injected via an environment variable `GREETING_DEFAULT_MESSAGE`. If this environment variable is set, it will override the value in the configuration file.

## Logging

Adjust logging behavior using the following command-line flags:

* `--no-logs`: Suppress non-critical logs.
* `--log`: Specify a file path to write logs.

Important: When using the MCP server with a stdio transport, logging must not be directed to standard output because it would interfere with the MCP protocol communication. Therefore, you should always use `--no-logs` along with `--log` to ensure that all logs are written exclusively to a log file.

## MCP Server Usage

MCP clients interact with the server by sending JSON‐RPC requests to execute various tools. The following MCP tools are supported:

* `greeting/hello`: Generates a greeting message, with an optional name parameter for personalization.

### Using with Claude Desktop

To integrate with Claude Desktop, add an entry to your `claude_desktop_config.json` file. Because MCP uses stdio for communication, you must redirect logs away from stdio by using the `--no-logs` and `--log` flags:

```json
{
  "mcpServers": {
    "greeting": {
      "command": "./bin/mcp-greeting",
      "args": ["server", "--no-logs", "--log", "mcp-greeting.log"],
      "env": {
        "GREETING_DEFAULT_MESSAGE": "こんにちは"
      }
    }
  }
}
```

This configuration registers the MCP Greeting Server with Claude Desktop, ensuring that all logs are directed to the specified log file rather than interfering with the MCP protocol messages transmitted over stdio.

## Contributing

Contributions are welcome! Please fork the repository and submit pull requests for improvements or bug fixes. For major changes, open an issue first to discuss your ideas.

## License

This project is licensed under the MIT License.

Author: cnosuke ( x.com/cnosuke )
