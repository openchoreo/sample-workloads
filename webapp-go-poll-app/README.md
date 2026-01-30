# Go Poll & Voting Application

A full-featured server-side Go web application for creating polls and gathering votes in real-time.

## Quick Start

```bash
# Using Docker (recommended)
docker build -t go-poll-app .
docker run -p 8080:8080 go-poll-app

# OR using Go directly
go mod download
go run main.go
```

Then open http://localhost:8080 in your browser.

## Features

- **Create Polls**: Design custom polls with multiple answer options
- **Real-time Voting**: Cast votes and see results instantly
- **Visual Results**: Beautiful progress bars showing vote distribution
- **Session Management**: Prevents duplicate voting using secure cookies
- **Concurrent Access**: Thread-safe data handling with Go's sync primitives
- **Modern UI**: Responsive gradient design with smooth animations
- **Sample Data**: Pre-loaded with example polls for demonstration

## Architecture

### Technology Stack
- **Go 1.21+** - Core language and standard library
- **Gorilla Sessions** - Secure cookie-based session management
- **Google UUID** - Unique poll identifier generation
- **HTML Templates** - Server-side rendering
- **In-Memory Storage** - Fast, thread-safe data storage with sync.RWMutex

### Server-Side Features
- Concurrent request handling with goroutines
- Thread-safe poll data access
- Session-based vote tracking
- POST/Redirect/GET pattern for form submissions
- Template functions for percentage calculations

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
- Create polls, vote, and view real-time results

## Local Development

### Prerequisites

- Docker (recommended)
- OR Go 1.21+ and a web browser

### Running with Docker (Recommended)

Build and run the Docker container:

```bash
# Navigate to the application directory
cd webapp-go-poll-app

# Build the Docker image
docker build -t go-poll-app .

# Run the container
docker run -d -p 8080:8080 --name go-poll-app go-poll-app
```

The application will be available at http://localhost:8080

To stop the container:

```bash
docker stop go-poll-app
docker rm go-poll-app
```

### Running Locally without Docker

```bash
# Navigate to the application directory
cd webapp-go-poll-app

# Install dependencies
go mod download

# Run the application
go run main.go
```

The application will start on http://localhost:8080

## How to Use

### Creating a Poll
1. Click **Create New Poll** from the home page
2. Enter your poll question
3. Add at least 2 answer options (use **Add Option** for more)
4. Click **Create Poll** to publish
5. Share the poll URL with others

### Voting on a Poll
1. Select a poll from the home page
2. Choose your preferred option
3. Click **Cast Your Vote**
4. View real-time results with visual progress bars and percentages
5. Sessions prevent duplicate voting on the same poll

### Viewing Results
- Progress bars show vote distribution for each option
- Percentages and vote counts update in real-time
- Results are visible immediately after voting
- Pre-loaded sample polls available for demonstration
