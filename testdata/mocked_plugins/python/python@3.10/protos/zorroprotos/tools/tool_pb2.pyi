from zorroprotos.tools import socket_pb2 as _socket_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ToolType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    COMMAND: _ClassVar[ToolType]
    ACTION: _ClassVar[ToolType]
    WIDGET: _ClassVar[ToolType]
    HOOK: _ClassVar[ToolType]

class ToolStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    INITIALIZING: _ClassVar[ToolStatus]
    INITIALIZED: _ClassVar[ToolStatus]
    RUNNING: _ClassVar[ToolStatus]
    PAUSED: _ClassVar[ToolStatus]
    ERROR: _ClassVar[ToolStatus]
    INVALID: _ClassVar[ToolStatus]
COMMAND: ToolType
ACTION: ToolType
WIDGET: ToolType
HOOK: ToolType
INITIALIZING: ToolStatus
INITIALIZED: ToolStatus
RUNNING: ToolStatus
PAUSED: ToolStatus
ERROR: ToolStatus
INVALID: ToolStatus

class ToolBase(_message.Message):
    __slots__ = ["id", "name", "type", "label", "status", "inputs", "outputs", "tooltip", "logs"]
    class InputsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _socket_pb2.Socket
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_socket_pb2.Socket, _Mapping]] = ...) -> None: ...
    class OutputsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _socket_pb2.Socket
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_socket_pb2.Socket, _Mapping]] = ...) -> None: ...
    class LogsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: int
        value: str
        def __init__(self, key: _Optional[int] = ..., value: _Optional[str] = ...) -> None: ...
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    LABEL_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    TOOLTIP_FIELD_NUMBER: _ClassVar[int]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    type: ToolType
    label: str
    status: ToolStatus
    inputs: _containers.MessageMap[str, _socket_pb2.Socket]
    outputs: _containers.MessageMap[str, _socket_pb2.Socket]
    tooltip: str
    logs: _containers.ScalarMap[int, str]
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., type: _Optional[_Union[ToolType, str]] = ..., label: _Optional[str] = ..., status: _Optional[_Union[ToolStatus, str]] = ..., inputs: _Optional[_Mapping[str, _socket_pb2.Socket]] = ..., outputs: _Optional[_Mapping[str, _socket_pb2.Socket]] = ..., tooltip: _Optional[str] = ..., logs: _Optional[_Mapping[int, str]] = ...) -> None: ...

class Tool(_message.Message):
    __slots__ = ["base"]
    BASE_FIELD_NUMBER: _ClassVar[int]
    base: ToolBase
    def __init__(self, base: _Optional[_Union[ToolBase, _Mapping]] = ...) -> None: ...
