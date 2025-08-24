# Tech Stack & Conventions

## Go Development
- **Go Version**: 1.24.4 (as per go.mod)
- **Module**: github.com/YuminosukeSato/pyproc
- **Dependencies**: 
  - spf13/viper for configuration
  - (planned) goccy/go-json or segmentio/encoding for JSON
  - (planned) vmihailenco/msgpack for MessagePack

### Coding Conventions
- Use short, clear variable names
- Avoid stuttering in names
- Always check errors
- Use `errors.Is/As` for error handling
- Prefer channels over shared memory for concurrency
- Follow standard Go project layout

### Testing
- Unit tests for each component
- Benchmark tests in `/bench` directory
- E2E tests combining Go and Python

## Python Development
- **Python Version**: 3.9+ (3.12 recommended)
- **Package Manager**: UV (pip is forbidden per CLAUDE.md)
- **Worker Package**: pyproc-worker

### Python Conventions
- Type hints for all functions
- Use orjson for JSON (fastest)
- Optional msgspec for typed serialization
- Follow ruff linting rules
- Maximum line length: 100 characters

## Quality Standards
- **Linters**:
  - Go: golangci-lint (configured in .golangci.yml)
  - Python: ruff, mypy
- **Testing**: go test, pytest
- **Formatting**: gofmt for Go, ruff format for Python

## Security Requirements
- HMAC authentication between Go and Python
- SO_PEERCRED verification on Unix sockets
- Socket permissions: 0660
- Run Python workers with low privileges
- Resource limits (rlimit) on Python processes