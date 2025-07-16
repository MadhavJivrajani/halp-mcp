# `halp-mcp`


For more context and info, check out the [accompanying post](https://nonmonotonic.dev/posts/halp-mcp/)!

## MCP Server

The MCP server is implemented at the root of the repo.
Having cloned the repo locally, you can install/build the binary:
```sh
# Or go build -o halp-mcp .
go install .
```

And then modify the [`claude_desktop_config.json`](https://modelcontextprotocol.io/quickstart/user#2-add-the-filesystem-mcp-server) file:
```json
{
  "mcpServers": {
    "halp-mcp": {
      "command": "/path/to/halp-mcp/binary"
    }
  }
}
```

Or, you could directly `go run` as well:
```json
{
  "mcpServers": {
    "halp-mcp": {
      "command": "go",
      "args": [
        "run",
        "/path/to/halp-mcp",
      ]
    }
  }
}
```

## `halp-watch`

Firstly, install [`halp`](https://github.com/MadhavJivrajani/halp).

```sh
go install github.com/MadhavJivrajani/halp@latest
```

To `WATCH` for `ConfigMaps`, you need to have a Kubernetes cluster up and running with your `kubeconfig`
pointed to it. Once you have that, simply run the code in the `halp-watch` directory:
```sh
cd halp-watch
go run .
```
