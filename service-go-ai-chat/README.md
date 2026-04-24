# AI Chat Service

A Go service that proxies chat requests to any OpenAI-compatible LLM API. Designed to demonstrate the `ai-llm-routing` ClusterTrait on OpenChoreo, which routes LLM traffic through Agent Gateway.

## How It Works

The service exposes a simple `POST /chat` endpoint. It forwards messages to an LLM provider using the OpenAI chat completions API format. When deployed with the `ai-llm-routing` trait, the `OPENAI_BASE_URL` is automatically injected, pointing to Agent Gateway which handles provider routing, API key management, rate limiting, and PII guardrails.

## Deploy in OpenChoreo

Follow these steps to deploy the application in OpenChoreo:

### 1. Create a Component
- Set up OpenChoreo following the instructions at https://openchoreo.dev
- Open the Backstage UI and navigate to **Create**
- Select **Component Type: Service**
- After creation, navigate to the **Workflows** tab to view builds

### 2. Build and Deploy
- Once the build completes successfully, go to the **Deploy** tab
- Click **Deploy** to deploy the service to your environment

### 3. Test the Service
Navigate to the **Test** section in the left menu to test the service endpoints.

## API Endpoints

### POST /chat

Send a message and get an AI-generated reply.

**Request:**
```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What is Kubernetes?"}'
```

**Response:**
```json
{
  "reply": "Kubernetes is an open-source container orchestration platform...",
  "model": "gpt-4o-mini",
  "usage": {
    "prompt_tokens": 25,
    "completion_tokens": 42,
    "total_tokens": 67
  }
}
```

You can optionally specify a model in the request to switch providers on the fly:
```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello", "model": "llama-3.3-70b-versatile"}'
```

### GET /healthz, GET /readyz

Health and readiness probes.

## Configuration

| Environment Variable | Required | Default | Description |
|---------------------|----------|---------|-------------|
| `OPENAI_API_KEY` | No | - | API key (not needed when using `ai-llm-routing` trait) |
| `OPENAI_BASE_URL` | No | `https://api.openai.com` | Base URL (injected by `ai-llm-routing` trait) |
| `OPENAI_MODEL` | No | `gpt-4o-mini` | Default model to use |
| `SYSTEM_PROMPT` | No | Generic helpful assistant | System prompt for the conversation |
| `LOG_LEVEL` | No | `info` | Log level (debug, info, warn, error) |
| `X_OPENCHOREO_COMPONENT` | No | - | Component identity header (injected by trait) |

## Project Structure

```
service-go-ai-chat/
├── main.go              # Main service implementation
├── go.mod               # Go module definition
├── Dockerfile           # Container build configuration
├── workload.yaml        # OpenChoreo workload descriptor
└── README.md            # This file
```

## Local Development

### Prerequisites

- Go 1.24 or later
- An OpenAI API key (or compatible provider)

### Run Locally

```bash
export OPENAI_API_KEY="sk-..."
go run . --port 8080
```

### Test Locally

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, what can you do?"}'
```
