from uuid import uuid4

from tortoise.models import Model
from tortoise import fields


class BaseModel(Model):
    id = fields.UUIDField(pk=True, default=uuid4)
    name = fields.CharField(null=False, max_length=100)
    label = fields.CharField(null=False, max_length=100)
    data = fields.JSONField()
    created_at = fields.DatetimeField(null=True, auto_now_add=True)
    updated_at = fields.DatetimeField(null=True, auto_now=True)

    class Meta:
        abstract = True
