from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Socket(_message.Message):
    __slots__ = ["raw_string", "raw_integer", "raw_number", "raw_binary", "cast"]
    RAW_STRING_FIELD_NUMBER: _ClassVar[int]
    RAW_INTEGER_FIELD_NUMBER: _ClassVar[int]
    RAW_NUMBER_FIELD_NUMBER: _ClassVar[int]
    RAW_BINARY_FIELD_NUMBER: _ClassVar[int]
    CAST_FIELD_NUMBER: _ClassVar[int]
    raw_string: str
    raw_integer: int
    raw_number: float
    raw_binary: bytes
    cast: str
    def __init__(self, raw_string: _Optional[str] = ..., raw_integer: _Optional[int] = ..., raw_number: _Optional[float] = ..., raw_binary: _Optional[bytes] = ..., cast: _Optional[str] = ...) -> None: ...
