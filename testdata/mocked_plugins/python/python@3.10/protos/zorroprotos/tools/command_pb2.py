# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: zorroprotos/tools/command.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from zorroprotos.tools import tool_pb2 as zorroprotos_dot_tools_dot_tool__pb2
from zorroprotos.processor import processor_query_pb2 as zorroprotos_dot_processor_dot_processor__query__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1fzorroprotos/tools/command.proto\x12\x05zorro\x1a\x1czorroprotos/tools/tool.proto\x1a+zorroprotos/processor/processor_query.proto\"X\n\x07\x43ommand\x12\x1d\n\x04\x62\x61se\x18\x01 \x01(\x0b\x32\x0f.zorro.ToolBase\x12.\n\x0fprocessor_query\x18\x02 \x01(\x0b\x32\x15.zorro.ProcessorQueryB1Z/github.com/Acedyn/zorro-proto/zorroprotos/toolsb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'zorroprotos.tools.command_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z/github.com/Acedyn/zorro-proto/zorroprotos/tools'
  _globals['_COMMAND']._serialized_start=117
  _globals['_COMMAND']._serialized_end=205
# @@protoc_insertion_point(module_scope)
