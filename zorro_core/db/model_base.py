from uuid import uuid4

from tortoise.models import Model
from tortoise import fields


class BaseModel(Model):
    id = fields.UUIDField(pk=True, default=uuid4)
    name = fields.CharField(max_length=100)
    label = fields.CharField(max_length=100)
    data = fields.JSONField(default=dict)
    created_at = fields.DatetimeField(auto_now_add=True)
    updated_at = fields.DatetimeField(auto_now=True)

    class Meta:
        abstract = True
