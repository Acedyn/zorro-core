# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: zorroprotos/processor/processor.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from zorroprotos.processor import processor_status_pb2 as zorroprotos_dot_processor_dot_processor__status__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n%zorroprotos/processor/processor.proto\x12\x05zorro\x1a,zorroprotos/processor/processor_status.proto\"\xec\x03\n\tProcessor\x12\n\n\x02id\x18\x01 \x01(\t\x12\x0c\n\x04name\x18\x02 \x01(\t\x12\x0f\n\x07version\x18\x03 \x01(\t\x12\r\n\x05label\x18\x04 \x01(\t\x12\r\n\x05paths\x18\x05 \x03(\t\x12\x0f\n\x07subsets\x18\x06 \x03(\t\x12\x1e\n\x16start_program_template\x18\x07 \x01(\t\x12 \n\x18start_processor_template\x18\x08 \x01(\t\x12&\n\x06status\x18\t \x01(\x0e\x32\x16.zorro.ProcessorStatus\x12\x30\n\x08metadata\x18\n \x03(\x0b\x32\x1e.zorro.Processor.MetadataEntry\x12,\n\x06stdout\x18\x0b \x03(\x0b\x32\x1c.zorro.Processor.StdoutEntry\x12,\n\x06stderr\x18\x0c \x03(\x0b\x32\x1c.zorro.Processor.StderrEntry\x1a/\n\rMetadataEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\x1a-\n\x0bStdoutEntry\x12\x0b\n\x03key\x18\x01 \x01(\x05\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\x1a-\n\x0bStderrEntry\x12\x0b\n\x03key\x18\x01 \x01(\x05\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\x42\x35Z3github.com/Acedyn/zorro-proto/zorroprotos/processorb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'zorroprotos.processor.processor_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z3github.com/Acedyn/zorro-proto/zorroprotos/processor'
  _PROCESSOR_METADATAENTRY._options = None
  _PROCESSOR_METADATAENTRY._serialized_options = b'8\001'
  _PROCESSOR_STDOUTENTRY._options = None
  _PROCESSOR_STDOUTENTRY._serialized_options = b'8\001'
  _PROCESSOR_STDERRENTRY._options = None
  _PROCESSOR_STDERRENTRY._serialized_options = b'8\001'
  _globals['_PROCESSOR']._serialized_start=95
  _globals['_PROCESSOR']._serialized_end=587
  _globals['_PROCESSOR_METADATAENTRY']._serialized_start=446
  _globals['_PROCESSOR_METADATAENTRY']._serialized_end=493
  _globals['_PROCESSOR_STDOUTENTRY']._serialized_start=495
  _globals['_PROCESSOR_STDOUTENTRY']._serialized_end=540
  _globals['_PROCESSOR_STDERRENTRY']._serialized_start=542
  _globals['_PROCESSOR_STDERRENTRY']._serialized_end=587
# @@protoc_insertion_point(module_scope)
