from zorroprotos.processor import processor_pb2 as _processor_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ProcessorRegistration(_message.Message):
    __slots__ = ["processor", "host"]
    PROCESSOR_FIELD_NUMBER: _ClassVar[int]
    HOST_FIELD_NUMBER: _ClassVar[int]
    processor: _processor_pb2.Processor
    host: str
    def __init__(self, processor: _Optional[_Union[_processor_pb2.Processor, _Mapping]] = ..., host: _Optional[str] = ...) -> None: ...
