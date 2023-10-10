from zorroprotos.processor import processor_pb2 as _processor_pb2
from zorroprotos.plugin import plugin_env_pb2 as _plugin_env_pb2
from zorroprotos.plugin import plugin_tools_pb2 as _plugin_tools_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Plugin(_message.Message):
    __slots__ = ["name", "version", "label", "path", "require", "env", "tools", "processors"]
    class EnvEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: _plugin_env_pb2.PluginEnv
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[_plugin_env_pb2.PluginEnv, _Mapping]] = ...) -> None: ...
    NAME_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    LABEL_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    REQUIRE_FIELD_NUMBER: _ClassVar[int]
    ENV_FIELD_NUMBER: _ClassVar[int]
    TOOLS_FIELD_NUMBER: _ClassVar[int]
    PROCESSORS_FIELD_NUMBER: _ClassVar[int]
    name: str
    version: str
    label: str
    path: str
    require: _containers.RepeatedScalarFieldContainer[str]
    env: _containers.MessageMap[str, _plugin_env_pb2.PluginEnv]
    tools: _plugin_tools_pb2.PluginTools
    processors: _containers.RepeatedCompositeFieldContainer[_processor_pb2.Processor]
    def __init__(self, name: _Optional[str] = ..., version: _Optional[str] = ..., label: _Optional[str] = ..., path: _Optional[str] = ..., require: _Optional[_Iterable[str]] = ..., env: _Optional[_Mapping[str, _plugin_env_pb2.PluginEnv]] = ..., tools: _Optional[_Union[_plugin_tools_pb2.PluginTools, _Mapping]] = ..., processors: _Optional[_Iterable[_Union[_processor_pb2.Processor, _Mapping]]] = ...) -> None: ...
