from zorro_core.network.protos import socket_pb2 as _socket_pb2
from zorro_core.network.protos import scheduler_pb2 as _scheduler_pb2
from zorro_core.network.protos import toolstatus_pb2 as _toolstatus_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ActionRequest(_message.Message):
    __slots__ = ["name", "type", "id", "label", "status", "inputs", "output", "tooltip", "logs", "implementation_key", "scheduler"]
    class InputsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _socket_pb2.Socket
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_socket_pb2.Socket, _Mapping]] = ...) -> None: ...
    class OutputEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _socket_pb2.Socket
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_socket_pb2.Socket, _Mapping]] = ...) -> None: ...
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    LABEL_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    TOOLTIP_FIELD_NUMBER: _ClassVar[int]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    IMPLEMENTATION_KEY_FIELD_NUMBER: _ClassVar[int]
    SCHEDULER_FIELD_NUMBER: _ClassVar[int]
    name: str
    type: str
    id: str
    label: str
    status: _toolstatus_pb2.ToolStatus
    inputs: _containers.MessageMap[str, _socket_pb2.Socket]
    output: _containers.MessageMap[str, _socket_pb2.Socket]
    tooltip: str
    logs: _containers.RepeatedScalarFieldContainer[str]
    implementation_key: str
    scheduler: _scheduler_pb2.Scheduler
    def __init__(self, name: _Optional[str] = ..., type: _Optional[str] = ..., id: _Optional[str] = ..., label: _Optional[str] = ..., status: _Optional[_Union[_toolstatus_pb2.ToolStatus, str]] = ..., inputs: _Optional[_Mapping[str, _socket_pb2.Socket]] = ..., output: _Optional[_Mapping[str, _socket_pb2.Socket]] = ..., tooltip: _Optional[str] = ..., logs: _Optional[_Iterable[str]] = ..., implementation_key: _Optional[str] = ..., scheduler: _Optional[_Union[_scheduler_pb2.Scheduler, _Mapping]] = ...) -> None: ...

class ActionUpdate(_message.Message):
    __slots__ = ["name", "type", "id", "label", "status", "inputs", "output", "tooltip", "logs", "implementation_key", "scheduler"]
    class InputsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _socket_pb2.Socket
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_socket_pb2.Socket, _Mapping]] = ...) -> None: ...
    class OutputEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _socket_pb2.Socket
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_socket_pb2.Socket, _Mapping]] = ...) -> None: ...
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    LABEL_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    TOOLTIP_FIELD_NUMBER: _ClassVar[int]
    LOGS_FIELD_NUMBER: _ClassVar[int]
    IMPLEMENTATION_KEY_FIELD_NUMBER: _ClassVar[int]
    SCHEDULER_FIELD_NUMBER: _ClassVar[int]
    name: str
    type: str
    id: str
    label: str
    status: _toolstatus_pb2.ToolStatus
    inputs: _containers.MessageMap[str, _socket_pb2.Socket]
    output: _containers.MessageMap[str, _socket_pb2.Socket]
    tooltip: str
    logs: _containers.RepeatedScalarFieldContainer[str]
    implementation_key: str
    scheduler: _scheduler_pb2.Scheduler
    def __init__(self, name: _Optional[str] = ..., type: _Optional[str] = ..., id: _Optional[str] = ..., label: _Optional[str] = ..., status: _Optional[_Union[_toolstatus_pb2.ToolStatus, str]] = ..., inputs: _Optional[_Mapping[str, _socket_pb2.Socket]] = ..., output: _Optional[_Mapping[str, _socket_pb2.Socket]] = ..., tooltip: _Optional[str] = ..., logs: _Optional[_Iterable[str]] = ..., implementation_key: _Optional[str] = ..., scheduler: _Optional[_Union[_scheduler_pb2.Scheduler, _Mapping]] = ...) -> None: ...
