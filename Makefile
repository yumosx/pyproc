.PHONY: demo bench bench-quick bench-full fmt lint test

demo:
	GO111MODULE=on go run ./examples/basic

bench: bench-quick

bench-quick:
	@echo "Running quick benchmarks..."
	@go test -bench=BenchmarkPool -benchtime=10x ./bench

bench-full:
	@echo "Running full benchmark suite..."
	@go test -bench=. -benchtime=100x -benchmem ./bench

test:
	@echo "Running tests..."
	@go test -v ./...

fmt:
	go fmt ./...

lint:
	@echo "Running linters..."
	@golangci-lint run ./...
	@cd worker/python && ruff check . || true

