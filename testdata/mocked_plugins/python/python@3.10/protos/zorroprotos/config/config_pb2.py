# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: zorroprotos/config/config.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from zorroprotos.config import user_config_pb2 as zorroprotos_dot_config_dot_user__config__pb2
from zorroprotos.config import plugin_config_pb2 as zorroprotos_dot_config_dot_plugin__config__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1fzorroprotos/config/config.proto\x12\x05zorro\x1a$zorroprotos/config/user_config.proto\x1a&zorroprotos/config/plugin_config.proto\"a\n\x06\x43onfig\x12+\n\x10user_preferences\x18\x01 \x01(\x0b\x32\x11.zorro.UserConfig\x12*\n\rplugin_config\x18\x02 \x01(\x0b\x32\x13.zorro.PluginConfigB2Z0github.com/Acedyn/zorro-proto/zorroprotos/configb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'zorroprotos.config.config_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z0github.com/Acedyn/zorro-proto/zorroprotos/config'
  _globals['_CONFIG']._serialized_start=120
  _globals['_CONFIG']._serialized_end=217
# @@protoc_insertion_point(module_scope)
