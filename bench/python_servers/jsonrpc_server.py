#!/usr/bin/env python3
"""
JSON-RPC 2.0 server implementation over Unix Domain Socket.
"""

import json
import os
import socket
import sys
import threading
from base_server import METHODS


class JSONRPCServer:
    """JSON-RPC 2.0 server using Unix Domain Socket."""

    def __init__(self, socket_path):
        self.socket_path = socket_path
        self.sock = None
        self.running = False

    def start(self):
        """Start the JSON-RPC server."""
        # Remove existing socket file if it exists
        if os.path.exists(self.socket_path):
            os.unlink(self.socket_path)

        # Create Unix Domain Socket
        self.sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.sock.bind(self.socket_path)
        self.sock.listen(5)
        self.running = True

        print(f"JSON-RPC server listening on {self.socket_path}")

        while self.running:
            try:
                conn, _ = self.sock.accept()
                # Handle each connection in a separate thread
                thread = threading.Thread(target=self.handle_connection, args=(conn,))
                thread.daemon = True
                thread.start()
            except Exception as e:
                if self.running:
                    print(f"Error accepting connection: {e}")

    def handle_connection(self, conn):
        """Handle a client connection."""
        try:
            while True:
                # Read JSON-RPC request
                data = self.read_message(conn)
                if not data:
                    break

                # Process request
                response = self.process_request(data)

                # Send response
                self.send_message(conn, response)
        except Exception as e:
            print(f"Error handling connection: {e}")
        finally:
            conn.close()

    def read_message(self, conn):
        """Read a JSON message from the connection."""
        # Read until we get a complete JSON object
        buffer = b""
        decoder = json.JSONDecoder()

        while True:
            chunk = conn.recv(4096)
            if not chunk:
                return None

            buffer += chunk
            try:
                # Try to decode JSON from buffer
                text = buffer.decode("utf-8")
                obj, idx = decoder.raw_decode(text)
                # Remove processed data from buffer
                buffer = text[idx:].encode("utf-8")
                return obj
            except (json.JSONDecodeError, UnicodeDecodeError):
                # Need more data
                continue

    def send_message(self, conn, message):
        """Send a JSON message to the connection."""
        data = json.dumps(message).encode("utf-8")
        conn.sendall(data)

    def process_request(self, request):
        """Process a JSON-RPC request and return a response."""
        # Validate request format
        if not isinstance(request, dict):
            return self.error_response(None, -32700, "Parse error")

        jsonrpc = request.get("jsonrpc")
        method = request.get("method")
        params = request.get("params")
        req_id = request.get("id")

        # Validate JSON-RPC version
        if jsonrpc != "2.0":
            return self.error_response(req_id, -32600, "Invalid Request")

        # Check if method exists
        if method not in METHODS:
            return self.error_response(req_id, -32601, "Method not found")

        try:
            # Call the method
            result = METHODS[method](params)

            # Return success response
            return {"jsonrpc": "2.0", "result": result, "id": req_id}
        except Exception as e:
            # Return error response
            return self.error_response(req_id, -32603, f"Internal error: {str(e)}")

    def error_response(self, req_id, code, message):
        """Create an error response."""
        return {
            "jsonrpc": "2.0",
            "error": {"code": code, "message": message},
            "id": req_id,
        }

    def stop(self):
        """Stop the server."""
        self.running = False
        if self.sock:
            self.sock.close()
        if os.path.exists(self.socket_path):
            os.unlink(self.socket_path)


def main():
    """Main entry point."""
    if len(sys.argv) < 2:
        print("Usage: python jsonrpc_server.py <socket_path>")
        sys.exit(1)

    socket_path = sys.argv[1]
    server = JSONRPCServer(socket_path)

    try:
        server.start()
    except KeyboardInterrupt:
        print("\nShutting down JSON-RPC server...")
        server.stop()


if __name__ == "__main__":
    main()
