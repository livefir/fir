# Sanity Example

This is a simple Todo application built with FIR to demonstrate basic functionality and serve as a sanity test.

## Features

- Create new todos
- Toggle todo completion status
- Delete todos
- Real-time updates using FIR's reactive system

## Running the Example

```bash
cd examples/sanity
go mod tidy
go run example.go
```

Then open your browser to `http://localhost:8080`

## Running Tests

The sanity test uses ChromeDP to perform end-to-end testing:

```bash
go test
```

This test verifies that the basic todo creation functionality works correctly.

## Code Structure

- `example.go` - Main application with todo handlers and routes
- `example.html` - HTML template with FIR directives
- `example_test.go` - End-to-end sanity test
