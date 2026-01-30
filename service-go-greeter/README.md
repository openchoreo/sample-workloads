# Go Greeter Service

A simple HTTP greeter service built with Go's standard library. This service demonstrates a basic REST API with graceful shutdown handling.

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
Navigate to the **Test** section in the left menu and use the OpenAPI Console to test the service endpoints.

## API Endpoints

### Greet

Send a personalized greeting.

**Endpoint:** `GET /greeter/greet`

**Query Parameters:**
- `name` (optional) - The name to greet. Defaults to "Stranger" if not provided.

**Example Request:**
```
GET /greeter/greet?name=Alice
```

**Response:**
```
Hello, Alice!
```

**Example Request (without name):**
```
GET /greeter/greet
```

**Response:**
```
Hello, Stranger!
```

## Project Structure

```
service-go-greeter/
├── main.go              # Main service implementation
├── go.mod               # Go module definition
├── Dockerfile           # Container build configuration
├── openapi.yaml         # OpenAPI specification
├── workload.yaml        # OpenChoreo workload descriptor
└── README.md            # This file
```

## Local Development

### Prerequisites

- Go 1.x or later

### Run Locally

```bash
# Navigate to the service directory
cd service-go-greeter

# Run the service
go run main.go
```

The service will start on http://localhost:9090

### Test Locally

```bash
# Test with a name
curl "http://localhost:9090/greeter/greet?name=Alice"

# Test without a name
curl "http://localhost:9090/greeter/greet"
```