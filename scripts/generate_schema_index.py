#!/usr/bin/env python3

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any


HTTP_METHODS = {"get", "post", "put", "patch", "delete"}


def endpoint_group(path: str) -> str:
    if path.startswith("/reports/"):
        return "reports"
    if path.startswith("/custom-reports"):
        return "impression-share"
    if "product-page-reasons" in path or "/assets/find" in path:
        return "ad-rejections"
    if "product-pages" in path or path in {
        "/countries-or-regions",
        "/creativeappmappings/devices",
    }:
        return "product-pages"
    if "targetingkeywords" in path:
        return "keywords"
    if "negativekeywords" in path:
        return "negatives-adgroup" if "/adgroups/" in path else "negatives-campaign"
    if "/ads" in path or path.startswith("/ads"):
        return "ads"
    if path.startswith("/search/apps") or "/eligibilities/" in path:
        return "apps"
    if path.startswith("/search/geo") or path.startswith("/geodata"):
        return "geo"
    if path.startswith("/budgetorders"):
        return "budgetorders"
    if path.startswith("/adgroups"):
        return "adgroups"

    parts = [part for part in path.split("/") if part and not part.startswith("{")]
    if not parts:
        return "other"
    first = parts[0]
    if first == "campaigns" and len(parts) > 1:
        second = parts[1]
        if second == "adgroups":
            return "adgroups"
        if second == "negativekeywords":
            return "negatives-campaign"
    return first


def type_group(name: str) -> str:
    lower = name.lower()
    if "customreport" in lower:
        return "impression-share"
    if "apppreviewdevices" in lower or "countriesorregions" in lower:
        return "product-pages"
    if "report" in lower or lower == "reportingrequest":
        return "reports"
    if "campaign" in lower:
        return "campaigns"
    if "adgroup" in lower or "ad_group" in lower:
        return "adgroups"
    if "keyword" in lower:
        return "negatives" if "negative" in lower else "keywords"
    if lower.startswith("ad") and "adam" not in lower:
        return "ads"
    if "creative" in lower:
        return "creatives"
    if "budgetorder" in lower:
        return "budgetorders"
    if "productpage" in lower:
        return "product-pages"
    if "app" in lower or "eligibility" in lower:
        return "apps"
    if "geo" in lower or "searchentity" in lower:
        return "geo"
    if "acl" in lower or lower == "useracl":
        return "acls"
    return "common"


def schema_fields(schema: dict[str, Any]) -> list[str]:
    fields: list[str] = []
    properties = schema.get("properties")
    if isinstance(properties, dict):
        fields.extend(properties)
    all_of = schema.get("allOf")
    if isinstance(all_of, list):
        for nested in all_of:
            if isinstance(nested, dict):
                for field in schema_fields(nested):
                    if field not in fields:
                        fields.append(field)
    return fields


def schema_type_name(name: str, schema: dict[str, Any]) -> str:
    title = schema.get("title")
    if isinstance(title, str) and re.search(r"[A-Za-z0-9]", title):
        return title
    return name


def build_schema_index(openapi: dict[str, Any]) -> dict[str, Any]:
    paths = openapi.get("paths")
    if not isinstance(paths, dict):
        raise ValueError("OpenAPI JSON must contain an object at paths")

    endpoints: list[dict[str, str]] = []
    for path, path_item in paths.items():
        if not isinstance(path, str) or not isinstance(path_item, dict):
            continue
        for method, operation in path_item.items():
            if method not in HTTP_METHODS or not isinstance(operation, dict):
                continue
            description = operation.get("summary") or operation.get("description") or ""
            endpoints.append(
                {
                    "method": method.upper(),
                    "path": path,
                    "group": endpoint_group(path),
                    "description": str(description),
                }
            )

    schemas = openapi.get("components", {}).get("schemas", {})
    if not isinstance(schemas, dict):
        raise ValueError("OpenAPI JSON must contain an object at components.schemas")

    types: list[dict[str, Any]] = []
    for name, schema in schemas.items():
        if not isinstance(name, str) or not isinstance(schema, dict):
            continue
        type_name = schema_type_name(name, schema)
        types.append(
            {
                "name": type_name,
                "group": type_group(type_name),
                "fields": schema_fields(schema),
            }
        )

    return {"endpoints": endpoints, "types": types}


def render(index: dict[str, Any]) -> str:
    return json.dumps(index, indent=2) + "\n"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate the embedded aads schema index from generated OpenAPI JSON."
    )
    parser.add_argument(
        "openapi_path",
        help="Path to an OpenAPI JSON file produced by scripts/generate_openapi.py.",
    )
    parser.add_argument("output_path", help="Path where schema_index.json should be written.")
    parser.add_argument(
        "--check",
        action="store_true",
        help="Check whether the output file is up to date instead of writing it.",
    )
    return parser.parse_args()


def resolve_path(path: str, repo_root: Path) -> Path:
    resolved = Path(path)
    if not resolved.is_absolute():
        resolved = repo_root / resolved
    return resolved


def main() -> None:
    repo_root = Path(__file__).resolve().parent.parent
    args = parse_args()
    openapi_path = resolve_path(args.openapi_path, repo_root)
    output_path = resolve_path(args.output_path, repo_root)

    openapi = json.loads(openapi_path.read_text(encoding="utf-8"))
    rendered = render(build_schema_index(openapi))

    if args.check:
        current = output_path.read_text(encoding="utf-8") if output_path.exists() else ""
        if current != rendered:
            print(
                f"{output_path} is out of date; run `make schema-index`.",
                file=sys.stderr,
            )
            raise SystemExit(1)
        return

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(rendered, encoding="utf-8")


if __name__ == "__main__":
    main()
