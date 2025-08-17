#!/usr/bin/env python3
"""
XML-RPC server implementation over Unix Domain Socket.
"""

import os
import socket
import sys
from xmlrpc.server import SimpleXMLRPCServer
from socketserver import UnixStreamServer
from base_server import METHODS


class UnixXMLRPCServer(UnixStreamServer, SimpleXMLRPCServer):
    """XML-RPC server using Unix Domain Socket."""
    
    def __init__(self, socket_path, requestHandler=None, logRequests=False):
        # Remove existing socket file if it exists
        if os.path.exists(socket_path):
            os.unlink(socket_path)
        
        self.socket_path = socket_path
        
        # Initialize UnixStreamServer with the socket path
        UnixStreamServer.__init__(self, socket_path, requestHandler)
        
        # Initialize SimpleXMLRPCServer functionality
        SimpleXMLRPCServer.__init__(self, None, requestHandler, logRequests, 
                                    allow_none=True, encoding=None, 
                                    use_builtin_types=True)
        
        # Register methods
        for name, func in METHODS.items():
            self.register_function(func, name)
        
        print(f"XML-RPC server listening on {socket_path}")
    
    def server_bind(self):
        """Override to use Unix socket binding."""
        # Socket is already bound by UnixStreamServer
        pass
    
    def server_activate(self):
        """Override to use Unix socket activation."""
        self.socket.listen(5)
    
    def shutdown(self):
        """Clean shutdown."""
        super().shutdown()
        if os.path.exists(self.socket_path):
            os.unlink(self.socket_path)


def main():
    """Main entry point."""
    if len(sys.argv) < 2:
        print("Usage: python xmlrpc_server.py <socket_path>")
        sys.exit(1)
    
    socket_path = sys.argv[1]
    
    # Create and run the server
    server = UnixXMLRPCServer(socket_path)
    
    try:
        print(f"XML-RPC server running on {socket_path}")
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down XML-RPC server...")
        server.shutdown()


if __name__ == '__main__':
    main()