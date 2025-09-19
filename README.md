# Changing Tables API

A simple REST API service for finding changing table locations, built with Go.

## Features

- REST endpoint to retrieve changing table locations
- JSON response format
- Lightweight HTTP server using Go standard library

## API Endpoints

- `GET /locations` - Returns a list of all changing table locations

## Response Format

```json
[
  {
    "id": 1,
    "name": "Popo Downtown",
    "address": "123 Main St",
    "latitude": 50.6829,
    "longitude": -26.9890
  }
]
```

## Running the Server

```bash
go run main.go
```

The server will start on port 8080.

## Testing

Test the API endpoint:
```bash
curl http://localhost:8080/locations
```

## Development

- `go build` - Build the application
- `go fmt` - Format code
- `go vet` - Check for issues
- `go test` - Run tests