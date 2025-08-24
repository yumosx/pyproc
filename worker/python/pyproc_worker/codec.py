"""Codec implementations for pyproc worker."""

import json
from abc import ABC, abstractmethod
from typing import Any

try:
    import orjson
    HAS_ORJSON = True
except ImportError:
    HAS_ORJSON = False

try:
    import msgspec
    HAS_MSGSPEC = True
except ImportError:
    HAS_MSGSPEC = False


class Codec(ABC):
    """Abstract base class for codecs."""

    @abstractmethod
    def encode(self, obj: Any) -> bytes:
        """Encode an object to bytes."""

    @abstractmethod
    def decode(self, data: bytes) -> Any:
        """Decode bytes to an object."""

    @property
    @abstractmethod
    def name(self) -> str:
        """Return the codec name."""


class JSONCodec(Codec):
    """Standard library JSON codec."""

    def encode(self, obj: Any) -> bytes:
        return json.dumps(obj).encode("utf-8")

    def decode(self, data: bytes) -> Any:
        return json.loads(data.decode("utf-8"))

    @property
    def name(self) -> str:
        return "json-stdlib"


class OrjsonCodec(Codec):
    """orjson-based JSON codec (faster)."""

    def __init__(self):
        if not HAS_ORJSON:
            raise ImportError("orjson is not installed")

    def encode(self, obj: Any) -> bytes:
        return orjson.dumps(obj)

    def decode(self, data: bytes) -> Any:
        return orjson.loads(data)

    @property
    def name(self) -> str:
        return "json-orjson"


class MsgspecCodec(Codec):
    """msgspec-based codec (fastest, with type validation)."""

    def __init__(self):
        if not HAS_MSGSPEC:
            raise ImportError("msgspec is not installed. Install with: pip install msgspec")
        self.encoder = msgspec.json.Encoder()
        self.decoder = msgspec.json.Decoder()

    def encode(self, obj: Any) -> bytes:
        return self.encoder.encode(obj)

    def decode(self, data: bytes) -> Any:
        return self.decoder.decode(data)

    @property
    def name(self) -> str:
        return "msgspec"


class MsgpackCodec(Codec):
    """MessagePack codec using msgspec."""

    def __init__(self):
        if not HAS_MSGSPEC:
            raise ImportError("msgspec is not installed. Install with: pip install msgspec")
        self.encoder = msgspec.msgpack.Encoder()
        self.decoder = msgspec.msgpack.Decoder()

    def encode(self, obj: Any) -> bytes:
        return self.encoder.encode(obj)

    def decode(self, data: bytes) -> Any:
        return self.decoder.decode(data)

    @property
    def name(self) -> str:
        return "msgpack"


def get_codec(codec_type: str = "auto") -> Codec:
    """Get a codec instance by type.

    Args:
        codec_type: One of "auto", "json", "orjson", "msgspec", "msgpack"
                   "auto" will choose the fastest available codec

    Returns:
        A Codec instance

    """
    if codec_type == "auto":
        # Try to use the fastest available codec
        if HAS_MSGSPEC:
            return MsgspecCodec()
        if HAS_ORJSON:
            return OrjsonCodec()
        return JSONCodec()
    if codec_type == "json":
        return JSONCodec()
    if codec_type == "orjson":
        return OrjsonCodec()
    if codec_type == "msgspec":
        return MsgspecCodec()
    if codec_type == "msgpack":
        return MsgpackCodec()
    raise ValueError(f"Unknown codec type: {codec_type}")


# Default codec - use orjson if available, fallback to stdlib
if HAS_ORJSON:
    default_codec = OrjsonCodec()
else:
    default_codec = JSONCodec()

