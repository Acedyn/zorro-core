from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class PluginEnv(_message.Message):
    __slots__ = ["append", "prepend", "set"]
    APPEND_FIELD_NUMBER: _ClassVar[int]
    PREPEND_FIELD_NUMBER: _ClassVar[int]
    SET_FIELD_NUMBER: _ClassVar[int]
    append: _containers.RepeatedScalarFieldContainer[str]
    prepend: _containers.RepeatedScalarFieldContainer[str]
    set: str
    def __init__(self, append: _Optional[_Iterable[str]] = ..., prepend: _Optional[_Iterable[str]] = ..., set: _Optional[str] = ...) -> None: ...
