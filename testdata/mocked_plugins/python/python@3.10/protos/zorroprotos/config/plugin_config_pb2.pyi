from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class PluginConfig(_message.Message):
    __slots__ = ["default_require", "repos"]
    DEFAULT_REQUIRE_FIELD_NUMBER: _ClassVar[int]
    REPOS_FIELD_NUMBER: _ClassVar[int]
    default_require: _containers.RepeatedScalarFieldContainer[str]
    repos: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, default_require: _Optional[_Iterable[str]] = ..., repos: _Optional[_Iterable[str]] = ...) -> None: ...
