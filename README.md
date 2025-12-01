# TCE - Document Analysis Engine

A YAML-based document parser and analysis engine with AI-powered capabilities.

## Prerequisites

### System Dependencies

This project requires **pandoc** for DOCX parsing fallback support.

#### Ubuntu/Debian
```bash
apt-get update && apt-get install -y pandoc
```

#### macOS
```bash
brew install pandoc
```

#### Windows
Download and install from: https://pandoc.org/installing.html

### Go Dependencies

Install Go dependencies:
```bash
go mod download
```

## How DOCX Parsing Works

The system uses a two-tier approach for parsing DOCX files:

1. **Primary**: Native Go `unioffice` library (no external dependencies)
2. **Fallback**: `pandoc` command-line tool (provides rich content extraction when unioffice fails)

If pandoc is not installed, the parser will still attempt to use unioffice, but may fail on certain document formats.

## Running the Project

```bash
go run cmd/server/main.go
```

## Project Structure

- `internal/docparser/` - Document parsing logic
  - Uses `unioffice` library for native Go DOCX parsing
  - Falls back to `pandoc` for complex documents
- `cmd/server/` - HTTP server entry point
