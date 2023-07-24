from __future__ import annotations
from tortoise import fields

from .model_base import BaseModel


class EntityType(BaseModel):
    """
    Entity types are like user defined classes to allow any project structure
    """

    name = fields.CharField(null=False, max_length=100, unique=True)
    parent = fields.ForeignKeyField("models.EntityType", related_name="children", null=True)
    path_template = fields.CharField(max_length=255, null=True)

    class Meta:
        name = "entity_type"


class Entity(BaseModel):
    """
    Anything that can be worked on is an entity. This very unopiniated structure
    allows to fit in most existing datastructure
    """

    type = fields.ForeignKeyField("models.EntityType", related_name="instances")
    path = fields.CharField(max_length=255, null=True)
    parent = fields.ForeignKeyField("models.Entity", related_name="children", null=True)
    casting = fields.ManyToManyField(
        "models.Entity", related_name="casted", through="entity_casting"
    )

    class Meta:
        name = "entity"
