# Reading Notes MCP Server

A native MCP (Model Context Protocol) server that provides reading note management tools. Designed to demonstrate MCP server federation with the `ai-mcp-federation` ClusterTrait on OpenChoreo.

## How It Works

The server implements the MCP protocol (JSON-RPC 2.0 over Streamable HTTP) and exposes three tools for managing reading notes. When registered as an `mcpTarget` in the `ai-mcp-federation` trait, Agent Gateway federates these tools behind a single `MCP_GATEWAY_URL` endpoint.

## MCP Tools

| Tool           | Description                                       |
|----------------|---------------------------------------------------|
| `add_note`     | Add a reading note for a book (title + text)      |
| `list_notes`   | List all notes, optionally filtered by book title |
| `search_notes` | Search notes by keyword across titles and text    |

## Deploy in OpenChoreo

### 1. Create a Component
- Set up OpenChoreo following the instructions at https://openchoreo.dev
- Open the Backstage UI and navigate to **Create**
- Select **Component Type: Service**

### 2. Build and Deploy
- Once the build completes, go to the **Deploy** tab
- Click **Deploy** to deploy the service

## Local Development

### Prerequisites

- Go 1.24 or later

### Run Locally

```bash
go run . --port 8081
```

### Test Locally

```bash
# Initialize MCP session
curl -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}},"id":1}'

# List available tools
curl -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":2}'

# Add a note
curl -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"add_note","arguments":{"book_title":"Dune","text":"Incredible worldbuilding"}},"id":3}'
```

## Project Structure

```text
service-go-reading-notes-mcp/
├── main.go              # MCP server implementation
├── go.mod               # Go module definition
├── Dockerfile           # Container build configuration
├── workload.yaml        # OpenChoreo workload descriptor
└── README.md            # This file
```