# React Web Application with Nginx

A production-ready React application served through Nginx. This demonstrates a modern pattern for building and deploying React single-page applications with optimized static file serving.

## Quick Start

```bash
# Using Docker (recommended)
docker build -t react-nginx-app .
docker run -p 8080:80 react-nginx-app

# OR using Docker Compose
docker compose up -d

# OR for local development
npm install
npm start
```

Then open http://localhost:8080 (Docker) or http://localhost:3000 (local dev) in your browser.

## Features

- **React Framework**: Modern JavaScript library for building user interfaces
- **Nginx Web Server**: High-performance static file serving
- **Multi-Stage Build**: Optimized Docker image with separate build and runtime stages
- **Production Optimized**: Minified and bundled assets for fast loading
- **Hot Reload**: Development server with live reloading
- **Responsive Design**: Mobile-friendly UI components

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
- Explore the React web interface

## Local Development

### Prerequisites

- Docker (recommended)
- OR Node.js 18+ with npm/yarn

### Running with Docker (Recommended)

Build and run the Docker container:

```bash
# Navigate to the application directory
cd webapp-react-nginx

# Build the Docker image
docker build -t react-nginx-app .

# Run the container
docker run -d -p 8080:80 react-nginx-app
```

The application will be available at http://localhost:8080

To stop the container:

```bash
docker stop react-app
docker rm react-app
```

### Running with Docker Compose

```bash
# Navigate to the application directory
cd webapp-react-nginx

# Start the application
docker compose up -d

# View logs
docker compose logs -f

# Stop the application
docker compose down
```

The application will be available at http://localhost:80

### Running Locally without Docker

For development with hot reload:

```bash
# Navigate to the application directory
cd webapp-react-nginx

# Install dependencies
npm install
# OR
yarn install

# Start the development server
npm start
# OR
yarn start
```

The development server will start on http://localhost:3000

**Note**: Local development runs the React dev server with hot reload. For the production Nginx setup, use Docker.

## Project Structure

```
webapp-react-nginx/
├── Dockerfile              # Multi-stage Docker build
├── .dockerignore           # Docker build exclusions
├── compose.yaml            # Docker Compose configuration
├── package.json            # Node.js dependencies and scripts
├── yarn.lock               # Dependency lock file
├── workload.yaml           # OpenChoreo workload descriptor
├── README.md               # This file
├── .nginx/                 # Nginx configuration
│   └── nginx.conf          # Production server config
├── public/                 # Static public assets
│   ├── index.html          # HTML entry point
│   ├── favicon.ico
│   └── robots.txt
└── src/                    # React source code
    ├── App.js              # Main React component
    ├── App.css             # Application styles
    ├── index.js            # React entry point
    └── ...                 # Other components
```
