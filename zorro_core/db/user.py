from __future__ import annotations
from tortoise import fields
from enum import Enum

from .model_base import BaseModel


class User(BaseModel):
    """
    Basic user with its authentification infos
    """

    email = fields.CharField(null=False, max_length=255)
    groups = fields.ManyToManyField('models.Group', related_name='users', through='user_group')
    roles = fields.ManyToManyField('models.Role', related_name='users', through='user_role')

    class Meta:
        name = "user"

class Group(BaseModel):
    """
    Group users together to define roles and organise users
    """

    roles = fields.ManyToManyField('models.Role', related_name='groups', through='user_role')

    class Meta:
        name = "group"

class RoleEnum(Enum):
    ADMIN = "admin"
    MOD = "mod"
    USER = "user"

class Role(BaseModel):
    """
    Group users together to define roles and organise users
    """

    name = fields.CharEnumField(RoleEnum)

    class Meta:
        name = "role"
