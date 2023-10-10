from zorroprotos.tools import tool_pb2 as _tool_pb2
from zorroprotos.processor import processor_query_pb2 as _processor_query_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Command(_message.Message):
    __slots__ = ["base", "processor_query"]
    BASE_FIELD_NUMBER: _ClassVar[int]
    PROCESSOR_QUERY_FIELD_NUMBER: _ClassVar[int]
    base: _tool_pb2.ToolBase
    processor_query: _processor_query_pb2.ProcessorQuery
    def __init__(self, base: _Optional[_Union[_tool_pb2.ToolBase, _Mapping]] = ..., processor_query: _Optional[_Union[_processor_query_pb2.ProcessorQuery, _Mapping]] = ...) -> None: ...
