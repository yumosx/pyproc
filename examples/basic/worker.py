#!/usr/bin/env python3
"""
Basic example worker for pyproc.
Demonstrates simple function exposure and usage.
"""

import sys
import os
import json

# Add pyproc_worker to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../worker/python'))

from pyproc_worker import expose, run_worker

@expose
def predict(req):
    """
    Simple prediction function that multiplies input by 2.
    
    Args:
        req: Dict with 'value' key
    
    Returns:
        Dict with 'result' key
    """
    value = req.get('value', 0)
    return {
        'result': value * 2,
        'model': 'simple-multiplier',
        'confidence': 0.99
    }

@expose
def process_batch(req):
    """
    Process a batch of values.
    
    Args:
        req: Dict with 'values' list
    
    Returns:
        Dict with processed results
    """
    values = req.get('values', [])
    results = [v * 2 for v in values]
    return {
        'results': results,
        'count': len(results),
        'sum': sum(results)
    }

@expose
def transform_text(req):
    """
    Simple text transformation.
    
    Args:
        req: Dict with 'text' and 'operation' keys
    
    Returns:
        Dict with transformed text
    """
    text = req.get('text', '')
    operation = req.get('operation', 'upper')
    
    if operation == 'upper':
        result = text.upper()
    elif operation == 'lower':
        result = text.lower()
    elif operation == 'reverse':
        result = text[::-1]
    else:
        result = text
    
    return {
        'original': text,
        'transformed': result,
        'operation': operation
    }

@expose
def compute_stats(req):
    """
    Compute statistics for a list of numbers.
    
    Args:
        req: Dict with 'numbers' list
    
    Returns:
        Dict with statistics
    """
    numbers = req.get('numbers', [])
    
    if not numbers:
        return {
            'count': 0,
            'mean': 0,
            'min': 0,
            'max': 0,
            'sum': 0
        }
    
    return {
        'count': len(numbers),
        'mean': sum(numbers) / len(numbers),
        'min': min(numbers),
        'max': max(numbers),
        'sum': sum(numbers)
    }

if __name__ == '__main__':
    # Run the worker
    # Socket path can be provided as command line argument or environment variable
    socket_path = None
    if len(sys.argv) > 1:
        socket_path = sys.argv[1]
    
    run_worker(socket_path)