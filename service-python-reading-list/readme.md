# Python Reading List REST API

A REST API service for managing a reading list of books, built with Python Flask. This service demonstrates CRUD operations with in-memory storage.

## Features

- **Book Management**: Create, read, update, and delete books
- **Reading Status**: Track books with status (e.g., "to_read", "reading", "read")
- **In-Memory Storage**: Fast, lightweight data storage
- **OpenAPI Specification**: Complete API documentation
- **Unit Tests**: Comprehensive test coverage

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

Base path: `/reading-list`

### List All Books

**Endpoint:** `GET /reading-list/books`

**Response:**
```json
{
  "books": [
    {
      "id": 1,
      "name": "The Lord of the Rings",
      "author": "J. R. R. Tolkien",
      "status": "to_read"
    },
    {
      "id": 2,
      "name": "Harry Potter",
      "author": "J. K. Rowling",
      "status": "reading"
    }
  ]
}
```

### Add a New Book

**Endpoint:** `POST /reading-list/books`

**Request Body:**
```json
{
  "name": "The Hobbit",
  "author": "J. R. R. Tolkien",
  "status": "to_read"
}
```

**Response:**
```json
{
  "id": 3,
  "name": "The Hobbit",
  "author": "J. R. R. Tolkien",
  "status": "to_read"
}
```

### Get a Book by ID

**Endpoint:** `GET /reading-list/books/{id}`

**Response:**
```json
{
  "id": 1,
  "name": "The Lord of the Rings",
  "author": "J. R. R. Tolkien",
  "status": "to_read"
}
```

**Error Response (404):** Empty response with 404 status

### Update a Book Status

**Endpoint:** `PUT /reading-list/books/{id}`

**Request Body:**
```json
{
  "status": "read"
}
```

**Response:**
```json
{
  "id": 1,
  "name": "The Lord of the Rings",
  "author": "J. R. R. Tolkien",
  "status": "read"
}
```

### Delete a Book

**Endpoint:** `DELETE /reading-list/books/{id}`

**Response (204):** No content

## Project Structure

```
service-python-reading-list/
├── app.py               # Main Flask application
├── requirements.txt     # Python dependencies
├── Dockerfile          # Container build configuration
├── workload.yaml       # OpenChoreo workload descriptor
├── openapi.yaml        # OpenAPI specification
└── test_app.py         # Unit tests
```

## Local Development

### Prerequisites

- Python 3.x

### Run Locally

```bash
# Navigate to the service directory
cd service-python-reading-list

# Install dependencies
pip3 install -r requirements.txt

# Run the Flask application
flask run
```

The service will start on http://localhost:5000

### Run Tests

```bash
# Run unit tests
python3 test_app.py -v
```

### Test with cURL

```bash
# List all books
curl http://localhost:5000/reading-list/books

# Add a new book
curl -X POST http://localhost:5000/reading-list/books \
  -H "Content-Type: application/json" \
  -d '{
    "name": "The Hobbit",
    "author": "J. R. R. Tolkien",
    "status": "to_read"
  }'

# Get a specific book
curl http://localhost:5000/reading-list/books/1

# Update a book status
curl -X PUT http://localhost:5000/reading-list/books/1 \
  -H "Content-Type: application/json" \
  -d '{"status": "read"}'

# Delete a book
curl -X DELETE http://localhost:5000/reading-list/books/1
```