from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class ProcessorStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    STARTING: _ClassVar[ProcessorStatus]
    IDLE: _ClassVar[ProcessorStatus]
    PROCESSING: _ClassVar[ProcessorStatus]
    SHUTTING_DOWN: _ClassVar[ProcessorStatus]
    SHUT_DOWN: _ClassVar[ProcessorStatus]
    NOT_RESPONDING: _ClassVar[ProcessorStatus]
STARTING: ProcessorStatus
IDLE: ProcessorStatus
PROCESSING: ProcessorStatus
SHUTTING_DOWN: ProcessorStatus
SHUT_DOWN: ProcessorStatus
NOT_RESPONDING: ProcessorStatus
