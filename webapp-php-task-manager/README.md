# PHP Task Manager Web Application

A full-featured server-side PHP web application demonstrating session management, form handling, and dynamic UI rendering.

## Quick Start

```bash
# Using Docker (recommended)
docker build -t php-task-manager .
docker run -p 8080:80 php-task-manager

# OR using PHP built-in server
php -S localhost:8080
```

Then open http://localhost:8080 in your browser.

## Features

- **Task Management**: Add, complete, and delete tasks
- **Server-Side Sessions**: Tasks persist across page reloads
- **Responsive UI**: Modern gradient design with smooth animations
- **Real-Time Statistics**: Track total, completed, and pending tasks
- **Server Information**: Display PHP version and server details
- **Form Security**: Input sanitization with `htmlspecialchars()`
- **POST/Redirect/GET**: Prevents form resubmission issues

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
- Start managing your tasks with the web interface

## Local Development

### Prerequisites

- Docker (recommended)
- OR PHP 8.1+ with Apache/Nginx

### Running with Docker (Recommended)

Build and run the Docker container:

```bash
# Navigate to the application directory
cd webapp-php-task-manager

# Build the Docker image
docker build -t php-task-manager .

# Run the container
docker run -d -p 8080:80 --name php-task-app php-task-manager
```

The application will be available at http://localhost:8080

To stop the container:

```bash
docker stop php-task-app
docker rm php-task-app
```

### Running Locally without Docker

```bash
# Navigate to the application directory
cd webapp-php-task-manager

# Install dependencies (if using Composer)
composer install

# Start PHP built-in server
php -S localhost:8080
```

The application will start on http://localhost:8080

## How to Use

### Managing Tasks
1. **Add Tasks**: Enter a task description and click **Add Task**
2. **Complete Tasks**: Click **Done** to mark as completed (or **Undo** to revert)
3. **Delete Tasks**: Click **Delete** to remove individual tasks
4. **Clear All**: Use **Clear All Tasks** button to remove all tasks at once

### View Statistics
- See task counts at the top: Total, Completed, and Pending
- Task metadata shows creation timestamp
- Server information panel displays PHP version and session details

## Project Structure

```
webapp-php-task-manager/
├── index.php            # Main application file (PHP + HTML)
├── Dockerfile           # Multi-stage Docker build
├── .dockerignore        # Docker build exclusions
├── composer.json        # PHP dependencies (optional)
├── workload.yaml        # OpenChoreo workload descriptor
└── README.md            # This file
```
