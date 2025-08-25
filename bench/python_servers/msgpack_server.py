#!/usr/bin/env python3
"""
MessagePack-RPC server implementation over Unix Domain Socket.
"""

import os
import socket
import struct
import sys
import threading
import msgpack
from base_server import METHODS


class MessagePackRPCServer:
    """MessagePack-RPC server using Unix Domain Socket."""

    def __init__(self, socket_path):
        self.socket_path = socket_path
        self.sock = None
        self.running = False

    def start(self):
        """Start the MessagePack-RPC server."""
        # Remove existing socket file if it exists
        if os.path.exists(self.socket_path):
            os.unlink(self.socket_path)

        # Create Unix Domain Socket
        self.sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.sock.bind(self.socket_path)
        self.sock.listen(5)
        self.running = True

        print(f"MessagePack-RPC server listening on {self.socket_path}")

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
                # Read length prefix (4 bytes)
                length_data = self.recv_exact(conn, 4)
                if not length_data:
                    break

                # Unpack message length
                msg_length = struct.unpack(">I", length_data)[0]

                # Read message data
                msg_data = self.recv_exact(conn, msg_length)
                if not msg_data:
                    break

                # Process request
                response = self.process_request(msg_data)

                # Send response
                self.send_message(conn, response)
        except Exception as e:
            print(f"Error handling connection: {e}")
        finally:
            conn.close()

    def recv_exact(self, conn, n):
        """Receive exactly n bytes from connection."""
        data = b""
        while len(data) < n:
            chunk = conn.recv(n - len(data))
            if not chunk:
                return None
            data += chunk
        return data

    def send_message(self, conn, message):
        """Send a MessagePack message with length prefix."""
        # Pack the message
        data = msgpack.packb(message)

        # Send length prefix
        length_data = struct.pack(">I", len(data))
        conn.sendall(length_data)

        # Send message data
        conn.sendall(data)

    def process_request(self, data):
        """Process a MessagePack-RPC request and return a response."""
        try:
            # Unpack the request
            request = msgpack.unpackb(data, raw=False)

            # Validate request format [type, msgid, method, params]
            if not isinstance(request, list) or len(request) != 4:
                return self.error_response(0, "Invalid request format")

            msg_type, msg_id, method, params = request

            # Check message type (0 = request)
            if msg_type != 0:
                return self.error_response(msg_id, "Invalid message type")

            # Check if method exists
            if method not in METHODS:
                return self.error_response(msg_id, f"Method not found: {method}")

            try:
                # Call the method
                result = METHODS[method](params)

                # Return success response [type, msgid, error, result]
                return [1, msg_id, None, result]
            except Exception as e:
                # Return error response
                return self.error_response(msg_id, f"Internal error: {str(e)}")
        except Exception as e:
            return self.error_response(0, f"Failed to process request: {str(e)}")

    def error_response(self, msg_id, error):
        """Create an error response."""
        # Response format: [type, msgid, error, result]
        return [1, msg_id, error, None]

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
        print("Usage: python msgpack_server.py <socket_path>")
        sys.exit(1)

    socket_path = sys.argv[1]
    server = MessagePackRPCServer(socket_path)

    try:
        server.start()
    except KeyboardInterrupt:
        print("\nShutting down MessagePack-RPC server...")
        server.stop()


if __name__ == "__main__":
    main()
