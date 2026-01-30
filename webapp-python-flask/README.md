# Python Flask Web Application with Nginx

A sample web application built with Flask and served through Nginx as a reverse proxy. This demonstrates a production-ready pattern for deploying Python web applications.

## Quick Start

```bash
# Using Docker (recommended)
docker build -t webapp-python-flask .
docker run -p 8080:80 webapp-python-flask

# OR using Python directly
pip install -r requirements.txt
python app.py
```

Then open http://localhost:8080 (Docker) or http://localhost:5000 (local) in your browser.

## Features

- **Flask Framework**: Lightweight Python web framework
- **Nginx Reverse Proxy**: Handles static files and proxies dynamic requests
- **RESTful API**: Sample endpoints demonstrating API patterns
- **Static File Serving**: Optimized static file delivery through Nginx
- **Production Ready**: Uses Gunicorn WSGI server for production deployment

## API Endpoints

### Web Interface
- `GET /` - Main web page with application overview

### API Endpoints
- `GET /api/health` - Health check endpoint
  - Returns: `{"status": "healthy", "timestamp": "..."}`
- `GET /api/info` - Application information
  - Returns: `{"app": "Flask Web App", "version": "1.0", "python_version": "..."}`

## Deploy in OpenChoreo

Follow these steps to deploy the application in OpenChoreo:

### 1. Create a Component
- Set up OpenChoreo following the instructions at https://openchoreo.dev
- Open the Backstage UI and navigate to **Create**
- Select **Component Type: Web Application**
- After creation, navigate to the **Workflows** tab to view builds

### 2. Build and Deploy
- Once the build completes successfully, go to the **Deploy** tab
- Click **Deploy** to deploy the application to your environment

### 3. Access the Application
- After deployment, navigate to the application URL provided
- Access the web interface and API endpoints

## Local Development

### Prerequisites

- Docker (recommended)
- OR Python 3+ with pip

### Running with Docker (Recommended)

Build and run the Docker container:

```bash
# Navigate to the application directory
cd webapp-python-flask

# Build the Docker image
docker build -t webapp-python-flask .

# Run the container
docker run -d -p 8080:80 --name flask-app webapp-python-flask
```

The application will be available at http://localhost:8080

To stop the container:

```bash
docker stop flask-app
docker rm flask-app
```

### Running Locally without Docker

```bash
# Navigate to the application directory
cd webapp-python-flask

# Install dependencies
pip install -r requirements.txt

# Run the Flask development server
python app.py
```

The application will start on http://localhost:5000

**Note**: Local development runs Flask directly without Nginx. For the full Nginx + Gunicorn setup, use Docker.

## Project Structure

```
webapp-python-flask/
├── app.py                  # Flask application
├── requirements.txt        # Python dependencies
├── Dockerfile             # Docker build with Nginx + Flask
├── templates/             # HTML templates
│   └── index.html
├── static/               # Static assets
│   └── style.css
└── .nginx/              # Nginx configuration
    └── nginx.conf
```
