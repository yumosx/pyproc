# PyProc Project Overview

## Purpose
PyProc is a high-performance Go-Python IPC (Inter-Process Communication) library that allows calling Python functions from Go as if they were local functions, without CGO or microservices.

## Key Features
- **No CGO Required**: Pure Go implementation using Unix Domain Sockets
- **Bypass Python GIL**: Run multiple Python processes in parallel
- **Low Latency**: 45μs p50 latency, 200,000+ req/s with 8 workers
- **Process Isolation**: Python crashes don't affect Go service
- **Connection Pooling**: Reuse connections for high throughput

## Target Performance
- p50 latency: < 100μs
- p99 latency: < 500μs
- Throughput: 1-5k RPS per service instance
- Payload size: < 100KB JSON

## Project Structure
- `/pkg/pyproc/`: Core Go library implementation
- `/worker/python/`: Python worker implementation
- `/examples/`: Usage examples
- `/bench/`: Benchmark suite
- `/docs/`: Documentation
- `/cmd/`: Command-line tools

## Current Implementation Status
- ✅ Basic codec system (JSON, MessagePack)
- ✅ UDS transport layer
- ✅ Connection pooling
- ✅ Basic security features
- 🚧 Request multiplexing
- 🚧 Type-safe API with generics
- 🚧 OpenTelemetry integration
- 🚧 CLI scaffolding tool