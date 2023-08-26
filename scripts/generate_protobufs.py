from __future__ import annotations
from collections import defaultdict
from typing import Optional, TypedDict, cast
from pathlib import Path

from jinja2 import Template

from zorro_core.tools.command import Command
from zorro_core.utils.logger import logger


class DefinitionTypeB(TypedDict):
    type: Optional[str]
    format: Optional[str]
    items: Optional[DefinitionTypeA]

class DefinitionTypeA(TypedDict):
    type: Optional[str]
    items: Optional[DefinitionTypeB]

class DefinitionProperty(DefinitionTypeA):
    anyOf: Optional[list[DefinitionTypeA]]

SchemaDefinition = TypedDict("SchemaDefinition", {
    "name": str,
    "partial": Optional[bool],
    "enum": list[str],
    "description": str,
    "$defs": dict,
    "properties": dict[str, DefinitionProperty],
    "required": list[str],
})


PACKAGE_NAME = "zorro"
PROTO_FOLDER = Path("zorro_core") / "network" / "protos"
PROTO_FILE_ROOT = Path(__file__).parent.parent / PROTO_FOLDER

PROTO_TEMPLATE = """syntax = "proto3";
package {{ package }};

{% for import in imports %}
import "{{ (folder / import.lower()).as_posix() }}.proto";
{% endfor %}

{% for schema in schemas %}
{% if schema.enum %}
{% set index = namespace(value=0) %}
enum {{ schema.name }} {
    {% if schema.description %}
    /*{{schema.description}}*/
    {% endif %}
    {% for value in schema.enum %}
    {{ value }} = {{ index.value }};
    {% set index.value = index.value + 1 %}
    {% endfor %}
}
{% else %}
{% set index = namespace(value=1) %}
message {{ schema.name }} {
    {% if schema.description %}
    /*{{schema.description}}*/
    {% endif %}
    {% for property in schema.properties %}
    {% if property.anyOf %}
    oneof {{ property.name }} {
        {% for property_child in property.anyOf %}
        {% if property_child.format in type_mapping or property_child.type in type_mapping or property_child['$ref'] in type_mapping %}
        {{ type_mapping[property_child.format] or type_mapping[property_child.type] or type_mapping[property_child['$ref']] }} {{ property.name }}_{{ property_child.format or property_child.type }} = {{ index.value }};
        {% set index.value = index.value + 1 %}
        {% endif %}
        {% endfor %}
    }
    {% elif property.allOf %}
    {{'optional ' if type_index == 1 else ''}}{{ type_mapping[property.allOf[0].format] or type_mapping[property.allOf[0].type] or type_mapping[property.allOf[0]['$ref']] }} {{ property.name }} = {{ index.value }};
    {% set index.value = index.value + 1 %}
    {% elif property.type == 'object' %}
    map<string, {{ type_mapping[property.additionalProperties.format] or type_mapping[property.additionalProperties.type] or type_mapping[property.additionalProperties['$ref']] }}> {{ property.name }} = {{ index.value }};
    {% set index.value = index.value + 1 %}
    {% elif property.type == 'array' %}
    repeated {{ type_mapping[property['items'].format] or type_mapping[property['items'].type] or type_mapping[property['items']['$ref']] }} {{ property.name }} = {{ index.value }};
    {% set index.value = index.value + 1 %}
    {% elif property.format in type_mapping or property.type in type_mapping or property['$ref'] in type_mapping %}
    {{'optional ' if schema.partial else ''}}{{ type_mapping[property.format] or type_mapping[property.type] or type_mapping[property['$ref']] }} {{ property.name }} = {{ index.value }};
    {% set index.value = index.value + 1 %}
    {% endif %}
    {% endfor %}
}

{% endif %}
{% endfor %}
"""

PROTO_TYPES = {
    "string": "string",
    "integer": "int32",
    "number": "float",
    "binary": "bytes",
}


def generate_proto_from_schema(name: str, schemas: list[SchemaDefinition]):
    imports: list[str] = list(set(key for schema in schemas for key in schema.get("$defs", {}).keys()))

    template_values = {
        "schemas": [
            {
                **schema,
                "properties": [
                    {**property, "name": property_name}
                    for property_name, property in schema.get("properties", {}).items()
                ],
            } for schema in schemas
        ],
        "imports": imports,
        "type_mapping": {**PROTO_TYPES, **{f"#/$defs/{i}": f"{PACKAGE_NAME}.{i}" for i in imports}},
        "folder": PROTO_FOLDER,
        "package": PACKAGE_NAME,
    }

    template = Template(PROTO_TEMPLATE, trim_blocks=True, lstrip_blocks=True)
    proto_path = PROTO_FILE_ROOT / f"{name.lower()}.proto"
    with open(proto_path, "w") as f:
        f.write(template.render(template_values))

    logger.info("%s written", proto_path)


def main():
    command_model = Command.model_json_schema()
    action_model = Command.model_json_schema()
    schemas: dict[str, list[SchemaDefinition]] = defaultdict(list)

    for definition_name, definition in command_model.get("$defs", {}).items():
        schemas[definition_name] = [cast(SchemaDefinition, {**definition, "name": definition_name})]
    schemas["Command"] = [cast(SchemaDefinition, {**command_model, "name": "CommandRequest"}), cast(SchemaDefinition, {**command_model, "name": "CommandUpdate", "partial": True})]

    for definition_name, definition in action_model.get("$defs", {}).items():
        schemas[definition_name] = [cast(SchemaDefinition, {**definition, "name": definition_name})]
    schemas["Action"] = [cast(SchemaDefinition, {**action_model, "name": "ActionRequest"}), cast(SchemaDefinition, {**action_model, "name": "ActionUpdate", "partial": True})]

    for name, proto_schemas in schemas.items():
        generate_proto_from_schema(name, proto_schemas)


if __name__ == "__main__":
    main()
