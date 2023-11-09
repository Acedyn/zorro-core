from zorroprotos.tools import command_pb2 as _command_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class LogLevels(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    DEBUG: _ClassVar[LogLevels]
    INFO: _ClassVar[LogLevels]
    WARNING: _ClassVar[LogLevels]
    ERROR: _ClassVar[LogLevels]
    CRITICAL: _ClassVar[LogLevels]
DEBUG: LogLevels
INFO: LogLevels
WARNING: LogLevels
ERROR: LogLevels
CRITICAL: LogLevels

class LogInput(_message.Message):
    __slots__ = ["message", "level"]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    LEVEL_FIELD_NUMBER: _ClassVar[int]
    message: str
    level: LogLevels
    def __init__(self, message: _Optional[str] = ..., level: _Optional[_Union[LogLevels, str]] = ...) -> None: ...

class LogOutput(_message.Message):
    __slots__ = ["zorro_command", "message", "timestamp"]
    ZORRO_COMMAND_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    zorro_command: _command_pb2.Command
    message: str
    timestamp: int
    def __init__(self, zorro_command: _Optional[_Union[_command_pb2.Command, _Mapping]] = ..., message: _Optional[str] = ..., timestamp: _Optional[int] = ...) -> None: ...
