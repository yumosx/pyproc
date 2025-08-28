#!/usr/bin/env python3
"""
Example worker that demonstrates cancellation support.
"""

import sys
import os
import time
import threading

# Add pyproc_worker to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../worker/python"))

from pyproc_worker import expose, run_worker
from pyproc_worker.cancellation import CancellationError, CancellableOperation

# Global state for testing cleanup
cleanup_performed = False
cleanup_lock = threading.Lock()


@expose
def slow_operation(req, cancel_event=None):
    """
    A slow operation that can be cancelled.
    
    Args:
        req: Dict with 'duration' key (in seconds)
        cancel_event: Threading event that signals cancellation
    
    Returns:
        Dict with result
    """
    duration = req.get("duration", 1.0)
    operation_id = req.get("id", 0)
    
    if cancel_event is None:
        # No cancellation support, just sleep
        time.sleep(duration)
        return {"completed": True, "id": operation_id}
    
    # Use CancellableOperation helper
    with CancellableOperation(cancel_event, check_interval=10) as op:
        # Simulate work with periodic cancellation checks
        steps = int(duration * 100)  # 10ms per step
        for i in range(steps):
            op.check()  # Check for cancellation
            time.sleep(0.01)  # Do some work
            
            # Simulate some processing
            if i % 10 == 0:
                print(f"Operation {operation_id}: Step {i}/{steps}")
    
    return {"completed": True, "id": operation_id, "duration": duration}


@expose
def operation_with_cleanup(req, cancel_event=None):
    """
    An operation that performs cleanup on cancellation.
    
    Args:
        req: Dict with 'duration' and 'with_cleanup' keys
        cancel_event: Threading event that signals cancellation
    
    Returns:
        Dict with result
    """
    global cleanup_performed
    
    duration = req.get("duration", 1.0)
    needs_cleanup = req.get("with_cleanup", False)
    
    # Reset cleanup flag
    with cleanup_lock:
        cleanup_performed = False
    
    try:
        if cancel_event is None:
            time.sleep(duration)
        else:
            # Simulate work with cancellation checking
            steps = int(duration * 100)
            for i in range(steps):
                if cancel_event.is_set():
                    raise CancellationError(0, "Operation cancelled")
                time.sleep(0.01)
        
        return {"completed": True, "cleanup_needed": False}
        
    except CancellationError:
        # Perform cleanup on cancellation
        if needs_cleanup:
            with cleanup_lock:
                cleanup_performed = True
                print("Cleanup performed after cancellation")
        raise
        
    finally:
        # Always cleanup resources
        if needs_cleanup and not cleanup_performed:
            with cleanup_lock:
                cleanup_performed = True
                print("Cleanup performed in finally block")


@expose
def check_cleanup(req):
    """
    Check if cleanup was performed.
    
    Returns:
        Dict with cleanup status
    """
    global cleanup_performed
    
    with cleanup_lock:
        return {"cleanup_done": cleanup_performed}


@expose
def cancellable_compute(req, cancel_event=None):
    """
    A CPU-intensive computation that can be cancelled.
    
    Args:
        req: Dict with 'iterations' key
        cancel_event: Threading event that signals cancellation
    
    Returns:
        Dict with computation result
    """
    iterations = req.get("iterations", 1000000)
    
    result = 0
    check_interval = max(1, iterations // 100)  # Check 100 times
    
    for i in range(iterations):
        # Periodic cancellation check
        if cancel_event and i % check_interval == 0:
            if cancel_event.is_set():
                return {"completed": False, "partial_result": result, "iterations_done": i}
        
        # Simulate computation
        result += i * i
    
    return {"completed": True, "result": result, "iterations": iterations}


@expose
def io_bound_operation(req, cancel_event=None):
    """
    An I/O-bound operation that respects cancellation.
    
    Args:
        req: Dict with 'chunks' key
        cancel_event: Threading event that signals cancellation
    
    Returns:
        Dict with processing result
    """
    chunks = req.get("chunks", 10)
    processed = []
    
    for i in range(chunks):
        # Check for cancellation before each I/O operation
        if cancel_event and cancel_event.is_set():
            return {"completed": False, "processed": processed}
        
        # Simulate I/O operation
        time.sleep(0.1)
        processed.append(f"chunk_{i}")
    
    return {"completed": True, "processed": processed, "count": len(processed)}


if __name__ == "__main__":
    # Run the worker
    socket_path = None
    if len(sys.argv) > 1:
        socket_path = sys.argv[1]
    
    run_worker(socket_path)