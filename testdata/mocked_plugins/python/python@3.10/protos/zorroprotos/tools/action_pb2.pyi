from zorroprotos.tools import tool_pb2 as _tool_pb2
from zorroprotos.tools import command_pb2 as _command_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ActionChild(_message.Message):
    __slots__ = ["command", "action", "upstream"]
    COMMAND_FIELD_NUMBER: _ClassVar[int]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    UPSTREAM_FIELD_NUMBER: _ClassVar[int]
    command: _command_pb2.Command
    action: Action
    upstream: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, command: _Optional[_Union[_command_pb2.Command, _Mapping]] = ..., action: _Optional[_Union[Action, _Mapping]] = ..., upstream: _Optional[_Iterable[str]] = ...) -> None: ...

class Action(_message.Message):
    __slots__ = ["base", "children"]
    class ChildrenEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: ActionChild
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[ActionChild, _Mapping]] = ...) -> None: ...
    BASE_FIELD_NUMBER: _ClassVar[int]
    CHILDREN_FIELD_NUMBER: _ClassVar[int]
    base: _tool_pb2.ToolBase
    children: _containers.MessageMap[str, ActionChild]
    def __init__(self, base: _Optional[_Union[_tool_pb2.ToolBase, _Mapping]] = ..., children: _Optional[_Mapping[str, ActionChild]] = ...) -> None: ...
