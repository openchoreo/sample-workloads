# Patient Management Service (MediFlow) - Ballerina

A REST API service for managing patient records, built with Ballerina. This service demonstrates CRUD operations with an in-memory data store, providing endpoints to add, retrieve, and list patient information.

## Deploy in OpenChoreo

Follow these steps to deploy the application in OpenChoreo:

### 1. Create a Component
- Set up OpenChoreo following the instructions at https://openchoreo.dev
- Open the Backstage UI and navigate to **Create**
- Select **Component Type: Service**
- Provide the repository URL and application path
- After creation, navigate to the **Workflows** tab to view component builds

### 2. Build and Deploy
- Once the build completes successfully, go to the **Deploy** tab
- Click **Deploy** to deploy the service to your environment

## API Endpoints

All endpoints are prefixed with `/mediflow`.

### Health Check

**Endpoint:** `GET /mediflow/health`

**Response:**
```json
{
  "status": "MediFlow is operational"
}
```

### Add a New Patient

**Endpoint:** `POST /mediflow/patients`

**Request:**
```json
{
  "name": "Alice",
  "age": 30,
  "condition": "Healthy"
}
```

**Response:**
```json
{
  "message": "Patient added",
  "patient": {
    "name": "Alice",
    "age": 30,
    "condition": "Healthy"
  }
}
```

### Retrieve a Patient

**Endpoint:** `GET /mediflow/patients/{name}`

**Response:**
```json
{
  "name": "Alice",
  "age": 30,
  "condition": "Healthy"
}
```

### List All Patients

**Endpoint:** `GET /mediflow/patients`

**Response:**
```json
{
  "alice": {
    "name": "Alice",
    "age": 30,
    "condition": "Healthy"
  }
}
```