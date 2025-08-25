"""Test msgspec codec functionality."""

import pytest

from pyproc_worker.codec import HAS_MSGSPEC, MsgpackCodec, MsgspecCodec, get_codec


def test_msgspec_json_codec() -> None:
    """Test msgspec JSON codec."""
    if not HAS_MSGSPEC:
        pytest.skip("msgspec not installed")

    codec = MsgspecCodec()
    assert codec.name == "msgspec"

    # Test encoding and decoding
    data = {"key": "value", "number": 42, "list": [1, 2, 3]}
    encoded = codec.encode(data)
    assert isinstance(encoded, bytes)

    decoded = codec.decode(encoded)
    assert decoded == data


def test_msgpack_codec() -> None:
    """Test MessagePack codec."""
    if not HAS_MSGSPEC:
        pytest.skip("msgspec not installed")

    codec = MsgpackCodec()
    assert codec.name == "msgpack"

    # Test encoding and decoding
    data = {"key": "value", "number": 42, "list": [1, 2, 3]}
    encoded = codec.encode(data)
    assert isinstance(encoded, bytes)

    decoded = codec.decode(encoded)
    assert decoded == data


def test_get_codec_msgspec() -> None:
    """Test explicit msgspec codec selection."""
    if not HAS_MSGSPEC:
        pytest.skip("msgspec not installed")

    codec = get_codec("msgspec")
    assert codec.name == "msgspec"


def test_get_codec_msgpack() -> None:
    """Test explicit msgpack codec selection."""
    if not HAS_MSGSPEC:
        pytest.skip("msgspec not installed")

    codec = get_codec("msgpack")
    assert codec.name == "msgpack"


def test_msgspec_complex_data() -> None:
    """Test msgspec with complex data structures."""
    if not HAS_MSGSPEC:
        pytest.skip("msgspec not installed")

    complex_data = {
        "string": "hello world",
        "int": 42,
        "float": 3.14159,
        "bool": True,
        "null": None,
        "list": [1, 2, 3, 4, 5],
        "nested": {
            "a": 1,
            "b": {"c": 2, "d": [3, 4, 5]},
        },
        "unicode": "你好世界 🌍",
        "large_list": list(range(1000)),
    }

    # Test both msgspec JSON and MessagePack
    for codec_type in ["msgspec", "msgpack"]:
        codec = get_codec(codec_type)
        encoded = codec.encode(complex_data)
        decoded = codec.decode(encoded)
        assert decoded == complex_data


def test_msgspec_performance() -> None:
    """Simple performance test to verify msgspec is faster than stdlib JSON."""
    if not HAS_MSGSPEC:
        pytest.skip("msgspec not installed")

    import time

    # Create a large dataset
    large_data = {
        f"key_{i}": {"value": i, "squared": i**2, "text": f"item_{i}"} for i in range(1000)
    }

    # Test stdlib JSON
    json_codec = get_codec("json")
    start = time.perf_counter()
    for _ in range(10):
        encoded = json_codec.encode(large_data)
        _ = json_codec.decode(encoded)
    json_time = time.perf_counter() - start

    # Test msgspec
    msgspec_codec = get_codec("msgspec")
    start = time.perf_counter()
    for _ in range(10):
        encoded = msgspec_codec.encode(large_data)
        _ = msgspec_codec.decode(encoded)
    msgspec_time = time.perf_counter() - start

    # msgspec should be faster
    # Allow some margin for CI environments
    assert msgspec_time < json_time * 1.5, (
        f"msgspec ({msgspec_time:.4f}s) should be faster than JSON ({json_time:.4f}s)"
    )
