# AI Reading Assistant

An AI agent that discovers and calls tools from federated MCP servers through Agent Gateway. Designed to demonstrate the `ai-mcp-federation` and `ai-llm-routing` ClusterTraits on OpenChoreo.

## How It Works

The agent exposes a `POST /chat` endpoint. On startup, it connects to the MCP gateway (`MCP_GATEWAY_URL`) to discover available tools via `tools/list`. When a user sends a chat message, the agent:

1. Sends the message + available tools to an LLM via `OPENAI_BASE_URL`
2. If the LLM returns tool calls, executes them via the MCP gateway
3. Feeds tool results back to the LLM
4. Repeats until the LLM returns a text response

Both `OPENAI_BASE_URL` and `MCP_GATEWAY_URL` are injected by the `ai-llm-routing` and `ai-mcp-federation` traits respectively.

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/chat` | POST | Send a message, get an AI response with tool usage |
| `/refresh-tools` | POST | Re-discover tools from MCP gateway |
| `/healthz` | GET | Health probe |
| `/readyz` | GET | Readiness probe |

## Deploy in OpenChoreo

### 1. Create a Component
- Set up OpenChoreo following the instructions at https://openchoreo.dev
- Open the Backstage UI and navigate to **Create**
- Select **Component Type: Service**

### 2. Build and Deploy
- Once the build completes, go to the **Deploy** tab
- Click **Deploy** to deploy the service

## Configuration

| Environment Variable | Required | Default | Description |
|---------------------|----------|---------|-------------|
| `MCP_GATEWAY_URL` | Yes | - | MCP federation endpoint (injected by trait) |
| `OPENAI_BASE_URL` | No | `https://api.openai.com` | LLM endpoint (injected by trait) |
| `OPENAI_MODEL` | No | `gpt-4o-mini` | Default model to use |
| `OPENAI_API_KEY` | No | - | API key (not needed when using trait) |
| `X_OPENCHOREO_COMPONENT` | No | - | Component identity (injected by trait) |
| `LOG_LEVEL` | No | `info` | Log level (debug, info, warn, error) |

## Project Structure

```
service-go-ai-reading-assistant/
├── main.go              # Agent with MCP client + LLM tool-calling loop
├── go.mod               # Go module definition
├── Dockerfile           # Container build configuration
├── workload.yaml        # OpenChoreo workload descriptor
└── README.md            # This file
```

## Local Development

### Prerequisites

- Go 1.24 or later
- A running MCP server (e.g., `service-go-reading-notes-mcp`)
- An LLM API key

### Run Locally

```bash
export MCP_GATEWAY_URL="http://localhost:8081"
export OPENAI_API_KEY="sk-..."
go run . --port 8080
```

### Test Locally

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Add a note for Dune: incredible worldbuilding"}'
```
