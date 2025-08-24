#!/usr/bin/env python3
"""
Base functions for RPC server benchmarking.
All RPC servers will expose these same methods for fair comparison.
"""

import time
import statistics


def predict(params):
    """Simple prediction function that multiplies input by 2."""
    if isinstance(params, dict):
        value = params.get('value', 0)
    elif isinstance(params, (list, tuple)) and len(params) > 0:
        value = params[0] if isinstance(params[0], dict) else params[0]
        if isinstance(value, dict):
            value = value.get('value', 0)
    else:
        value = params if isinstance(params, (int, float)) else 0
    
    return {
        'result': value * 2,
        'model': 'simple-multiplier',
        'confidence': 0.99
    }


def process_batch(params):
    """Process a batch of values."""
    if isinstance(params, dict):
        values = params.get('values', [])
        metadata = params.get('metadata', {})
    else:
        values = params if isinstance(params, list) else []
        metadata = {}
    
    results = [v * 2 for v in values]
    return {
        'results': results,
        'count': len(results),
        'sum': sum(results),
        'metadata': metadata
    }


def compute_stats(params):
    """Compute statistics for a list of numbers."""
    if isinstance(params, dict):
        numbers = params.get('numbers', [])
        options = params.get('options', {})
    else:
        numbers = params if isinstance(params, list) else []
        options = {}
    
    if not numbers:
        return {
            'count': 0,
            'mean': 0,
            'min': 0,
            'max': 0,
            'sum': 0
        }
    
    result = {
        'count': len(numbers),
        'mean': statistics.mean(numbers),
        'min': min(numbers),
        'max': max(numbers),
        'sum': sum(numbers)
    }
    
    # Add optional statistics based on options
    if options.get('compute_variance'):
        result['variance'] = statistics.variance(numbers) if len(numbers) > 1 else 0
    if options.get('compute_std_dev'):
        result['std_dev'] = statistics.stdev(numbers) if len(numbers) > 1 else 0
    if options.get('compute_median'):
        result['median'] = statistics.median(numbers)
    
    return result


def echo_test(params):
    """Echo back the request for testing."""
    return {
        'echo': params,
        'timestamp': time.time()
    }


def health():
    """Health check endpoint."""
    return {
        'status': 'healthy',
        'timestamp': time.time()
    }


# Method registry for easy lookup
METHODS = {
    'predict': predict,
    'process_batch': process_batch,
    'compute_stats': compute_stats,
    'echo_test': echo_test,
    'health': health,
}