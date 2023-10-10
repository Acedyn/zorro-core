from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class PluginTools(_message.Message):
    __slots__ = ["commands", "actions", "hooks", "widgets"]
    COMMANDS_FIELD_NUMBER: _ClassVar[int]
    ACTIONS_FIELD_NUMBER: _ClassVar[int]
    HOOKS_FIELD_NUMBER: _ClassVar[int]
    WIDGETS_FIELD_NUMBER: _ClassVar[int]
    commands: _containers.RepeatedScalarFieldContainer[str]
    actions: _containers.RepeatedScalarFieldContainer[str]
    hooks: _containers.RepeatedScalarFieldContainer[str]
    widgets: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, commands: _Optional[_Iterable[str]] = ..., actions: _Optional[_Iterable[str]] = ..., hooks: _Optional[_Iterable[str]] = ..., widgets: _Optional[_Iterable[str]] = ...) -> None: ...
