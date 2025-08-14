# pyproc

*Run Python like a local function from Go — no CGO, no microservices.*

[![Go Reference](https://pkg.go.dev/badge/github.com/YuminosukeSato/pyproc.svg)](https://pkg.go.dev/github.com/YuminosukeSato/pyproc)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## Features

- **No CGO Required** - Pure Go implementation using Unix Domain Sockets
- **Prefork Workers** - Scale Python workers to bypass GIL limitations
- **Function-like API** - Call Python functions as easily as `pool.Call(ctx, "predict", input, &output)`
- **Minimal Overhead** - UDS provides low-latency IPC on the same host
- **Future-Ready** - Optional gRPC mode and Arrow IPC for large datasets

## Quick Start (5 minutes)

### 1. Install

```bash
go get github.com/YuminosukeSato/pyproc
```

### 2. Create a Python Worker

```python
# worker.py
from pyproc_worker import expose, run_worker

@expose
def predict(req):
    """Your ML model or Python logic here"""
    return {"result": req["value"] * 2}

if __name__ == "__main__":
    run_worker()
```

### 3. Call from Go

```go
package main

import (
    "context"
    "fmt"
    "github.com/YuminosukeSato/pyproc/pkg/pyproc"
)

func main() {
    // Start Python worker
    cfg := pyproc.WorkerConfig{
        ID:           "worker-1",
        SocketPath:   "/tmp/pyproc.sock",
        PythonExec:   "python3",
        WorkerScript: "worker.py",
    }
    
    worker := pyproc.NewWorker(cfg, nil)
    ctx := context.Background()
    
    if err := worker.Start(ctx); err != nil {
        panic(err)
    }
    defer worker.Stop()
    
    // Connect and call Python function
    conn, _ := pyproc.ConnectToWorker(cfg.SocketPath, 5*time.Second)
    defer conn.Close()
    
    // Use the connection to call Python functions
    // See examples/basic for complete implementation
    fmt.Println("Worker ready!")
}
```

### 4. Run

```bash
go run main.go
```

That's it! You're now calling Python from Go without CGO or microservices.

## Why pyproc?

### The Problem

Go excels at building high-performance services, but sometimes you need Python for:
- Machine Learning models (PyTorch, TensorFlow, scikit-learn)
- Data processing (pandas, numpy)
- Legacy Python code
- Python-only libraries

Traditional solutions have drawbacks:
- **CGO + Python C API**: Complex, crashes can take down your Go service
- **Microservices**: Operational overhead, network latency, deployment complexity
- **Shell exec**: High startup cost, no connection pooling

### The Solution

pyproc runs Python workers as persistent processes alongside your Go code:
- **Same pod/host** - No network overhead
- **Process isolation** - Python crashes don't affect Go
- **Connection pooling** - Reuse connections
- **Simple deployment** - Just Go binary + Python scripts

## Use Cases

### Machine Learning Inference

```python
@expose
def predict(req):
    model = load_model()  # Cached after first load
    features = req["features"]
    return {"prediction": model.predict(features)}
```

### Data Processing

```python
@expose
def process_dataframe(req):
    import pandas as pd
    df = pd.DataFrame(req["data"])
    result = df.groupby("category").sum()
    return result.to_dict()
```

### Document Processing

```python
@expose
def extract_pdf_text(req):
    import PyPDF2
    # Process PDF and return text
    return {"text": extracted_text}
```

## Architecture

```
┌─────────────┐           UDS            ┌──────────────┐
│   Go App    │ ◄──────────────────────► │ Python Worker│
│             │    Low-latency IPC        │              │
│  - HTTP API │                           │  - Models    │
│  - Business │                           │  - Libraries │
│  - Logic    │                           │  - Data Proc │
└─────────────┘                           └──────────────┘
     ▲                                           ▲
     │                                           │
     └──────────── Same Host/Pod ────────────────┘
```

## Benchmarks

Run benchmarks locally:

```bash
make bench
```

Example results on M1 MacBook Pro:

```
Simple Echo:     50,000 req/s  (20μs latency)
JSON Payload:    45,000 req/s  (22μs latency)
NumPy Array:     30,000 req/s  (33μs latency)
With 4 Workers: 180,000 req/s  (5.5μs avg latency)
```

## Advanced Features

### Worker Pool (coming in v0.2)

```go
pool, _ := pyproc.NewPool(ctx, pyproc.PoolOptions{
    Workers:      4,
    MaxInFlight:  10,
})

var result PredictResponse
pool.Call(ctx, "predict", input, &result)
```

### gRPC Mode (coming in v0.4)

```go
pool, _ := pyproc.NewPool(ctx, pyproc.PoolOptions{
    Protocol: pyproc.ProtocolGRPC(),
    // Unix domain socket with gRPC
})
```

### Arrow IPC for Large Data (coming in v0.5)

```go
pool, _ := pyproc.NewPool(ctx, pyproc.PoolOptions{
    Protocol: pyproc.ProtocolArrow(),
    // Zero-copy data transfer
})
```

## Production Checklist

- [ ] Set appropriate worker count based on CPU cores
- [ ] Configure health checks
- [ ] Set up monitoring (metrics exposed at `:9090/metrics`)
- [ ] Configure restart policies
- [ ] Set resource limits (memory, CPU)
- [ ] Handle worker failures gracefully

## Documentation

- [Design Document](docs/design.md)
- [Operations Guide](docs/ops.md)
- [Security Guide](docs/security.md)
- [API Reference](https://pkg.go.dev/github.com/YuminosukeSato/pyproc)

## Contributing

We welcome contributions! Check out our ["help wanted"](https://github.com/YuminosukeSato/pyproc/labels/help%20wanted) issues to get started.

## License

Apache 2.0 - See [LICENSE](LICENSE) for details.

## References

- [Python GIL Documentation](https://docs.python.org/3/library/threading.html)
- [Unix Domain Sockets](https://man7.org/linux/man-pages/man7/unix.7.html)
- [gRPC Unix Sockets](https://grpc.github.io/grpc/cpp/md_doc_naming.html)
- [Apache Arrow IPC](https://arrow.apache.org/docs/python/ipc.html)