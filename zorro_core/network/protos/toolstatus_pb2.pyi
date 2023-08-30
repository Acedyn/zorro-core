from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class ToolStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    INITIALIZING: _ClassVar[ToolStatus]
    INITIALIZED: _ClassVar[ToolStatus]
    RUNNING: _ClassVar[ToolStatus]
    PAUSED: _ClassVar[ToolStatus]
    ERROR: _ClassVar[ToolStatus]
    INVALID: _ClassVar[ToolStatus]
INITIALIZING: ToolStatus
INITIALIZED: ToolStatus
RUNNING: ToolStatus
PAUSED: ToolStatus
ERROR: ToolStatus
INVALID: ToolStatus
