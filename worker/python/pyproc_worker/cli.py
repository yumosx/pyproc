#!/usr/bin/env python3
"""
Command-line interface for pyproc-worker
"""

import sys
import argparse
import logging
from . import run_worker

def main():
    """Main entry point for CLI"""
    parser = argparse.ArgumentParser(
        description='Python worker for pyproc - Call Python from Go without CGO'
    )
    parser.add_argument(
        'worker_script',
        help='Path to the Python worker script'
    )
    parser.add_argument(
        '--socket-path',
        help='Unix domain socket path (overrides PYPROC_SOCKET_PATH)',
        default=None
    )
    parser.add_argument(
        '--log-level',
        help='Logging level',
        choices=['debug', 'info', 'warning', 'error'],
        default='info'
    )
    
    args = parser.parse_args()
    
    # Set up logging
    logging.basicConfig(
        level=getattr(logging, args.log_level.upper()),
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    # Import and run the worker script
    import importlib.util
    spec = importlib.util.spec_from_file_location("worker", args.worker_script)
    if spec and spec.loader:
        worker_module = importlib.util.module_from_spec(spec)
        sys.modules["worker"] = worker_module
        spec.loader.exec_module(worker_module)
    else:
        print(f"Error: Could not load worker script: {args.worker_script}")
        sys.exit(1)

if __name__ == '__main__':
    main()