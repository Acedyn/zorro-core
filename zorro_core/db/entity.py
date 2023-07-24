from __future__ import annotations
from tortoise import fields

from .model_base import BaseModel


class EntityType(BaseModel):
    """
    Entity types are like user defined classes to allow any project structure
    """

    parent = fields.ForeignKeyField("models.EntityType", related_name="children")
    path_template = fields.CharField(null=False, max_length=255)

    class Meta:
        name = "entity_type"


class Entity(BaseModel):
    """
    Anything that can be worked on is an entity. This very unopiniated structure
    allows to fit in most existing datastructure
    """

    type = fields.ForeignKeyField("models.EntityType", related_name="instances")
    path = fields.CharField(null=False, max_length=255)
    parent = fields.ForeignKeyField("models.Entity", related_name="children")
    casting = fields.ManyToManyField(
        "models.Entity", related_name="casted", through="entity_casting"
    )

    class Meta:
        name = "entity"
