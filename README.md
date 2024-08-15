# Collaborative Coding Platform

This is a real-time collaborative coding platform built with React and Go.

## Features

- Real-time collaborative code editing
- Support for multiple programming languages (JavaScript and Python)
- Code execution with output display
- Chat functionality
- Room creation and joining

## Setup

### Prerequisites

- Node.js
- Go
- Docker (optional, for containerized deployment)

### Running the application

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/collaborative-coding-platform.git
   cd collaborative-coding-platform
   ```

2. Backend setup:
   ```
   cd backend
   go mod download
   go run main.go
   ```

3. Frontend setup:
   ```
   cd frontend
   npm install
   npm start
   ```

4. Open `http://localhost:3000` in your browser.

## Deployment

For containerized deployment, use Docker Compose:

```
docker-compose up --build
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.