# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a simple Go HTTP API server for a changing tables location service. The project is minimal with only two files:
- `go.mod`: Go module definition
- `main.go`: Complete API implementation

## Architecture

The application is a single-file Go web server that:
- Serves a REST API endpoint `/locations`
- Returns JSON data for changing table locations with mock data
- Uses only Go standard library (net/http, encoding/json, log)
- Listens on port 8080

The `Location` struct defines the data model with fields: ID, Name, Address, Latitude, Longitude.

## Development Commands

**Run the server:**
```bash
go run main.go
```

**Build the application:**
```bash
go build
```

**Test the application:**
```bash
go test
```

**Format code:**
```bash
go fmt
```

**Vet code for issues:**
```bash
go vet
```

**Test the API endpoint:**
```bash
curl http://localhost:8080/locations
```

## Project Structure

This is a minimal Go project with no additional directories or complex architecture. All functionality is contained in `main.go` with mock data for location information.

## Git Workflow

**IMPORTANT:** Always work on the `jai-dev` branch. Never commit directly to `master`.

- Development branch: `jai-dev`
- Main branch: `master` (for PRs only)
- Always verify you're on `jai-dev` before making commits