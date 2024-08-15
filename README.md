# Collaborative Coding Platform

This is a real-time collaborative coding platform built with React and Go.

## Example
https://youtu.be/S8lKWPsSFyc

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
   git clone https://github.com/seunghoon34/colab-coding-platform-go.git
   cd colab-coding-platform-go
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

## API Endpoints

The backend provides the following API endpoints:

- `POST /create-room`: Create a new room
  - Request body: `{ "username": "string" }`
  - Response: `{ "roomCode": "string" }`

- `POST /execute`: Execute code
  - Request body: `{ "code": "string", "language": "string" }`
  - Response: `{ "output": "string" }` or `{ "error": "string" }`

## WebSocket

The application uses WebSocket for real-time communication. Connect to the WebSocket at:

```
ws://localhost:8080/ws/{roomCode}?username={username}
```

### WebSocket Messages

Messages sent through the WebSocket should be JSON objects with a `type` field. The following types are supported:

1. Code update:
   ```json
   {
     "type": "code",
     "content": "// Your code here"
   }
   ```

2. Language change:
   ```json
   {
     "type": "language",
     "content": "javascript"
   }
   ```

3. Chat message:
   ```json
   {
     "type": "chat",
     "content": "Hello, world!"
   }
   ```

## Deployment

For containerized deployment, use Docker Compose:

```
docker-compose up --build
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License