"""
pyproc_worker - Python worker for pyproc

This module implements the Python side of the pyproc protocol,
allowing Python functions to be exposed and called from Go.
"""

import json
import socket
import struct
import sys
import traceback
from typing import Any, Callable, Dict, Optional
import logging
import os

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    stream=sys.stderr
)
logger = logging.getLogger(__name__)

# Registry for exposed functions
_exposed_functions: Dict[str, Callable] = {}


def expose(func: Callable) -> Callable:
    """
    Decorator to expose a Python function to Go.
    
    Usage:
        @expose
        def my_function(req):
            return {"result": req["input"] * 2}
    """
    _exposed_functions[func.__name__] = func
    logger.info(f"Exposed function: {func.__name__}")
    return func


class FramedConnection:
    """Handles framed message communication over a socket."""
    
    def __init__(self, conn: socket.socket):
        self.conn = conn
    
    def read_message(self) -> Optional[bytes]:
        """Read a framed message from the socket."""
        # Read 4-byte length header
        length_bytes = self._read_exact(4)
        if not length_bytes:
            return None
        
        # Parse length (big-endian)
        length = struct.unpack('>I', length_bytes)[0]
        
        # Read message body
        message = self._read_exact(length)
        if not message:
            raise Exception("Failed to read complete message")
        
        return message
    
    def write_message(self, data: bytes) -> None:
        """Write a framed message to the socket."""
        # Write 4-byte length header (big-endian)
        length = len(data)
        self.conn.sendall(struct.pack('>I', length))
        
        # Write message body
        self.conn.sendall(data)
    
    def _read_exact(self, n: int) -> Optional[bytes]:
        """Read exactly n bytes from the socket."""
        data = b''
        while len(data) < n:
            chunk = self.conn.recv(n - len(data))
            if not chunk:
                return None if len(data) == 0 else data
            data += chunk
        return data


class Worker:
    """Main worker class that handles requests from Go."""
    
    def __init__(self, socket_path: str):
        self.socket_path = socket_path
        self.conn = None
        self.framed_conn = None
    
    def start(self):
        """Start the worker and listen for requests."""
        # Remove socket file if it exists
        if os.path.exists(self.socket_path):
            os.unlink(self.socket_path)
        
        # Create Unix domain socket
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        sock.bind(self.socket_path)
        sock.listen(1)
        
        logger.info(f"Worker listening on {self.socket_path}")
        
        while True:
            try:
                # Accept connection
                conn, _ = sock.accept()
                self.conn = conn
                self.framed_conn = FramedConnection(conn)
                
                logger.info("Accepted connection")
                
                # Handle requests on this connection
                self._handle_connection()
                
            except KeyboardInterrupt:
                logger.info("Worker shutting down")
                break
            except Exception as e:
                logger.error(f"Connection error: {e}")
                if self.conn:
                    self.conn.close()
    
    def _handle_connection(self):
        """Handle requests on the current connection."""
        while True:
            try:
                # Read request
                message = self.framed_conn.read_message()
                if not message:
                    logger.info("Connection closed by client")
                    break
                
                # Parse request
                request = json.loads(message.decode('utf-8'))
                logger.debug(f"Received request: {request}")
                
                # Process request
                response = self._process_request(request)
                
                # Send response
                response_bytes = json.dumps(response).encode('utf-8')
                self.framed_conn.write_message(response_bytes)
                
            except Exception as e:
                logger.error(f"Error handling request: {e}")
                # Try to send error response
                try:
                    error_response = {
                        "id": 0,
                        "ok": False,
                        "error": str(e)
                    }
                    response_bytes = json.dumps(error_response).encode('utf-8')
                    self.framed_conn.write_message(response_bytes)
                except:
                    pass
                break
    
    def _process_request(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Process a single request and return a response."""
        req_id = request.get("id", 0)
        method = request.get("method", "")
        body = request.get("body", {})
        
        # Check if method is exposed
        if method not in _exposed_functions:
            return {
                "id": req_id,
                "ok": False,
                "error": f"Method '{method}' not found"
            }
        
        try:
            # Call the exposed function
            func = _exposed_functions[method]
            result = func(body)
            
            return {
                "id": req_id,
                "ok": True,
                "body": result
            }
        except Exception as e:
            # Capture the full traceback for debugging
            tb = traceback.format_exc()
            logger.error(f"Error in method '{method}': {tb}")
            
            return {
                "id": req_id,
                "ok": False,
                "error": str(e)
            }


def run_worker(socket_path: Optional[str] = None):
    """
    Run the worker with the specified socket path.
    
    Args:
        socket_path: Path to the Unix domain socket.
                    If not provided, uses environment variable PYPROC_SOCKET_PATH.
    """
    if socket_path is None:
        socket_path = os.environ.get("PYPROC_SOCKET_PATH")
        if not socket_path:
            raise ValueError("Socket path must be provided or set in PYPROC_SOCKET_PATH")
    
    worker = Worker(socket_path)
    worker.start()


# Health check method (always available)
@expose
def health(req):
    """Health check endpoint."""
    return {"status": "healthy", "pid": os.getpid()}