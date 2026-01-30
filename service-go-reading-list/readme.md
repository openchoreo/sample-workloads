# Go Reading List REST API

A REST API service for managing a reading list of books, built with Go and the Fiber web framework. This service demonstrates CRUD operations, OpenAPI documentation with Swagger, and optional initial data loading.

## Features

- **Book Management**: Create, read, update, and delete books
- **Reading Status**: Track books as "to_read", "reading", or "read"
- **OpenAPI/Swagger**: Auto-generated API documentation with Swagger UI
- **Configuration Support**: Environment-based configuration
- **Initial Data Loading**: Optional preloading of sample data
- **Graceful Shutdown**: Proper signal handling for clean shutdowns

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

Base path: `/api/v1/reading-list`

### List All Books

**Endpoint:** `GET /api/v1/reading-list/books`

**Response:**
```json
[
  {
    "id": "fe2594d0-ccea-42a2-97ac-0487458b5642",
    "title": "The Lord of the Rings",
    "author": "J. R. R. Tolkien",
    "status": "to_read"
  },
  {
    "id": "3c17c654-0609-4842-9525-d98cef921a1d",
    "title": "Harry Potter and the Philosopher's Stone",
    "author": "J. K. Rowling",
    "status": "reading"
  }
]
```

### Add a New Book

**Endpoint:** `POST /api/v1/reading-list/books`

**Request Body:**
```json
{
  "title": "The Hobbit",
  "author": "J. R. R. Tolkien",
  "status": "to_read"
}
```

**Response (201):**
```json
{
  "id": "b1b4b3b0-0b1a-4b1a-8b1a-0b1a0b1a0b1a",
  "title": "The Hobbit",
  "author": "J. R. R. Tolkien",
  "status": "to_read"
}
```

### Get a Book by ID

**Endpoint:** `GET /api/v1/reading-list/books/{id}`

**Response:**
```json
{
  "id": "fe2594d0-ccea-42a2-97ac-0487458b5642",
  "title": "The Lord of the Rings",
  "author": "J. R. R. Tolkien",
  "status": "to_read"
}
```

### Update a Book

**Endpoint:** `PUT /api/v1/reading-list/books/{id}`

**Request Body:**
```json
{
  "title": "The Lord of the Rings",
  "author": "J. R. R. Tolkien",
  "status": "read"
}
```

### Delete a Book

**Endpoint:** `DELETE /api/v1/reading-list/books/{id}`

**Response (204):** No content

### Swagger UI

Access the interactive API documentation at:
```
GET /swagger/index.html
```

## Project Structure

```
service-go-reading-list/
├── main.go              # Main service entry point
├── go.mod               # Go module dependencies
├── Dockerfile           # Container build configuration
├── workload.yaml        # OpenChoreo workload descriptor
├── api/                 # API route handlers
├── internal/            # Internal packages
│   ├── config/          # Configuration management
│   └── utils/           # Utility functions
├── docs/                # Generated OpenAPI documentation
│   └── openapi.yaml
└── configs/             # Configuration files
    └── initial_data.json
```

## Local Development

### Prerequisites

- Go 1.x or later

### Run Locally

```bash
# Navigate to the service directory
cd service-go-reading-list

# Run the service
go run main.go
```

The service will start on http://localhost:8080

### Access Swagger UI

Open your browser and navigate to:
```
http://localhost:8080/swagger/index.html
```

### Test with cURL

```bash
# List all books
curl http://localhost:8080/api/v1/reading-list/books

# Add a new book
curl -X POST http://localhost:8080/api/v1/reading-list/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "The Hobbit",
    "author": "J. R. R. Tolkien",
    "status": "to_read"
  }'

# Get a specific book (replace {id} with actual book ID)
curl http://localhost:8080/api/v1/reading-list/books/{id}

# Update a book
curl -X PUT http://localhost:8080/api/v1/reading-list/books/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "title": "The Hobbit",
    "author": "J. R. R. Tolkien",
    "status": "read"
  }'

# Delete a book
curl -X DELETE http://localhost:8080/api/v1/reading-list/books/{id}
```

## Configuration

### Environment Variables

See [config.go](internal/config/config.go) for available configuration options.