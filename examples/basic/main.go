package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/YuminosukeSato/pyproc/internal/framing"
	"github.com/YuminosukeSato/pyproc/internal/protocol"
	"github.com/YuminosukeSato/pyproc/pkg/pyproc"
)

func main() {
	// Create logger
	logger := pyproc.NewLogger(pyproc.LoggingConfig{
		Level:        "info",
		Format:       "text",
		TraceEnabled: true,
	})

	// Get the worker script path relative to the repo root
	workerScript, err := filepath.Abs(filepath.Join("examples", "basic", "worker.py"))
	if err != nil {
		log.Fatal(err)
	}

	// Socket path
	socketPath := "/tmp/pyproc-example.sock"

	// Create worker configuration
	cfg := pyproc.WorkerConfig{
		ID:           "example-worker",
		SocketPath:   socketPath,
		PythonExec:   "python3",
		WorkerScript: workerScript,
		StartTimeout: 10 * time.Second,
	}

	// Create and start worker
	ctx := context.Background()
	worker := pyproc.NewWorker(cfg, logger)

	fmt.Println("Starting Python worker...")
	if err := worker.Start(ctx); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}
	defer func() {
		fmt.Println("\nStopping worker...")
		worker.Stop()
	}()

	fmt.Printf("Worker started (PID: %d)\n\n", worker.GetPID())

	// Connect to the worker
	conn, err := pyproc.ConnectToWorker(socketPath, 5*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect to worker: %v", err)
	}
	defer conn.Close()

	framer := framing.NewFramer(conn)

	// Example 1: Simple prediction
	fmt.Println("=== Example 1: Simple Prediction ===")
	if err := callPredict(framer, 42); err != nil {
		log.Printf("Predict failed: %v", err)
	}

	// Example 2: Batch processing
	fmt.Println("\n=== Example 2: Batch Processing ===")
	if err := callBatchProcess(framer, []int{1, 2, 3, 4, 5}); err != nil {
		log.Printf("Batch process failed: %v", err)
	}

	// Example 3: Text transformation
	fmt.Println("\n=== Example 3: Text Transformation ===")
	if err := callTextTransform(framer, "Hello PyProc!", "upper"); err != nil {
		log.Printf("Text transform failed: %v", err)
	}
	if err := callTextTransform(framer, "Hello PyProc!", "reverse"); err != nil {
		log.Printf("Text transform failed: %v", err)
	}

	// Example 4: Compute statistics
	fmt.Println("\n=== Example 4: Compute Statistics ===")
	if err := callComputeStats(framer, []float64{1.5, 2.5, 3.5, 4.5, 5.5}); err != nil {
		log.Printf("Compute stats failed: %v", err)
	}

	// Example 5: Health check
	fmt.Println("\n=== Example 5: Health Check ===")
	if err := callHealth(framer); err != nil {
		log.Printf("Health check failed: %v", err)
	}

	fmt.Println("\n✅ All examples completed successfully!")
}

func callPredict(framer *framing.Framer, value int) error {
	req, err := protocol.NewRequest(1, "predict", map[string]interface{}{
		"value": value,
	})
	if err != nil {
		return err
	}

	// Send request
	reqData, _ := req.Marshal()
	if err := framer.WriteMessage(reqData); err != nil {
		return err
	}

	// Read response
	respData, err := framer.ReadMessage()
	if err != nil {
		return err
	}

	var resp protocol.Response
	if err := resp.Unmarshal(respData); err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf("error from Python: %s", resp.ErrorMsg)
	}

	var result map[string]interface{}
	if err := resp.UnmarshalBody(&result); err != nil {
		return err
	}

	fmt.Printf("Input: %d\n", value)
	fmt.Printf("Result: %.0f (model: %s, confidence: %.2f)\n",
		result["result"], result["model"], result["confidence"])
	return nil
}

func callBatchProcess(framer *framing.Framer, values []int) error {
	req, err := protocol.NewRequest(2, "process_batch", map[string]interface{}{
		"values": values,
	})
	if err != nil {
		return err
	}

	// Send request
	reqData, _ := req.Marshal()
	if err := framer.WriteMessage(reqData); err != nil {
		return err
	}

	// Read response
	respData, err := framer.ReadMessage()
	if err != nil {
		return err
	}

	var resp protocol.Response
	if err := resp.Unmarshal(respData); err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf("error from Python: %s", resp.ErrorMsg)
	}

	var result map[string]interface{}
	if err := resp.UnmarshalBody(&result); err != nil {
		return err
	}

	fmt.Printf("Input values: %v\n", values)
	fmt.Printf("Processed results: %v\n", result["results"])
	fmt.Printf("Count: %.0f, Sum: %.0f\n", result["count"], result["sum"])
	return nil
}

func callTextTransform(framer *framing.Framer, text, operation string) error {
	req, err := protocol.NewRequest(3, "transform_text", map[string]interface{}{
		"text":      text,
		"operation": operation,
	})
	if err != nil {
		return err
	}

	// Send request
	reqData, _ := req.Marshal()
	if err := framer.WriteMessage(reqData); err != nil {
		return err
	}

	// Read response
	respData, err := framer.ReadMessage()
	if err != nil {
		return err
	}

	var resp protocol.Response
	if err := resp.Unmarshal(respData); err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf("error from Python: %s", resp.ErrorMsg)
	}

	var result map[string]interface{}
	if err := resp.UnmarshalBody(&result); err != nil {
		return err
	}

	fmt.Printf("Operation: %s\n", operation)
	fmt.Printf("Original: %s\n", result["original"])
	fmt.Printf("Transformed: %s\n", result["transformed"])
	return nil
}

func callComputeStats(framer *framing.Framer, numbers []float64) error {
	req, err := protocol.NewRequest(4, "compute_stats", map[string]interface{}{
		"numbers": numbers,
	})
	if err != nil {
		return err
	}

	// Send request
	reqData, _ := req.Marshal()
	if err := framer.WriteMessage(reqData); err != nil {
		return err
	}

	// Read response
	respData, err := framer.ReadMessage()
	if err != nil {
		return err
	}

	var resp protocol.Response
	if err := resp.Unmarshal(respData); err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf("error from Python: %s", resp.ErrorMsg)
	}

	var result map[string]interface{}
	if err := resp.UnmarshalBody(&result); err != nil {
		return err
	}

	fmt.Printf("Numbers: %v\n", numbers)
	fmt.Printf("Statistics:\n")
	fmt.Printf("  Count: %.0f\n", result["count"])
	fmt.Printf("  Mean: %.2f\n", result["mean"])
	fmt.Printf("  Min: %.2f\n", result["min"])
	fmt.Printf("  Max: %.2f\n", result["max"])
	fmt.Printf("  Sum: %.2f\n", result["sum"])
	return nil
}

func callHealth(framer *framing.Framer) error {
	req, err := protocol.NewRequest(5, "health", map[string]interface{}{})
	if err != nil {
		return err
	}

	// Send request
	reqData, _ := req.Marshal()
	if err := framer.WriteMessage(reqData); err != nil {
		return err
	}

	// Read response
	respData, err := framer.ReadMessage()
	if err != nil {
		return err
	}

	var resp protocol.Response
	if err := resp.Unmarshal(respData); err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf("error from Python: %s", resp.ErrorMsg)
	}

	var result map[string]interface{}
	if err := resp.UnmarshalBody(&result); err != nil {
		return err
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Health check response:\n%s\n", jsonBytes)
	return nil
}
