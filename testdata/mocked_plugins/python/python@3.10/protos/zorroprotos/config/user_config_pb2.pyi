from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Language(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    English: _ClassVar[Language]
    French: _ClassVar[Language]
    Spanish: _ClassVar[Language]
    Dutch: _ClassVar[Language]
English: Language
French: Language
Spanish: Language
Dutch: Language

class UserConfig(_message.Message):
    __slots__ = ["language"]
    LANGUAGE_FIELD_NUMBER: _ClassVar[int]
    language: Language
    def __init__(self, language: _Optional[_Union[Language, str]] = ...) -> None: ...
