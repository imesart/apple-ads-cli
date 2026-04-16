#!/usr/bin/env python3
from __future__ import annotations

import argparse
import difflib
import json
import os
import re
import sys
import time
import urllib.parse
import urllib.request
from collections import deque
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
DEFAULT_DOCS_URL = "https://developer.apple.com/documentation/apple_ads"
APPLE_DOCS_HOST = "developer.apple.com"
APPLE_DOCS_JSON_PREFIX = "https://developer.apple.com/tutorials/data/documentation/"
DEFAULT_ALLOWED_PREFIX = APPLE_DOCS_JSON_PREFIX + "apple_ads"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Generate Apple Ads endpoint inventory and a lightweight OpenAPI document "
            "from Apple's docs JSON."
        )
    )
    parser.add_argument(
        "openapi_output",
        help="Path to write the generated lightweight OpenAPI JSON file.",
    )
    parser.add_argument(
        "--docs-url",
        default=DEFAULT_DOCS_URL,
        help=(
            "Apple Ads documentation entrypoint. Accepts either a normal docs URL or a "
            f"tutorials/data JSON URL (default: {DEFAULT_DOCS_URL})."
        ),
    )
    parser.add_argument(
        "--output-inventory",
        help="Optional path to write the scraped inventory JSON file.",
    )
    parser.add_argument(
        "--output-paths",
        help="Optional path to write the flat METHOD PATH list.",
    )
    parser.add_argument(
        "--api-version",
        help=(
            "Optional Apple Ads API docs version label to validate against the "
            "version crawled from Apple's changelog."
        ),
    )
    parser.add_argument(
        "--yes",
        action="store_true",
        help=(
            "Skip prompts. Required to auto-accept updates when more than 2%% of paths change."
        ),
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="Check whether generated output files are up to date instead of writing them.",
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Print crawl progress, parsing decisions, and write steps.",
    )
    return parser.parse_args()


def root_json_url(docs_url: str) -> str:
    if docs_url.endswith(".json") and "/tutorials/data/documentation/" in docs_url:
        return docs_url

    parsed = urllib.parse.urlparse(docs_url)
    if parsed.netloc and parsed.netloc != APPLE_DOCS_HOST:
        raise ValueError(f"unsupported docs host {parsed.netloc!r}")

    path = parsed.path.rstrip("/")
    if not path.startswith("/documentation/"):
        raise ValueError(
            f"unsupported docs path {path!r}: expected a /documentation/... Apple docs URL"
        )

    slug = path.removeprefix("/documentation/").strip("/")
    if not slug:
        raise ValueError(f"could not derive docs slug from {docs_url!r}")

    return APPLE_DOCS_JSON_PREFIX + slug + ".json"


def allowed_prefix(root_url: str) -> str:
    if root_url.startswith(DEFAULT_ALLOWED_PREFIX):
        return DEFAULT_ALLOWED_PREFIX
    return root_url.removesuffix(".json")


def log(verbose: bool, message: str) -> None:
    if verbose:
        print(message, file=sys.stderr)


def fetch_json(url: str, verbose: bool = False) -> dict:
    log(verbose, f"fetch {url}")
    request = urllib.request.Request(
        url,
        headers={
            "Accept": "application/json",
            "User-Agent": "aads-inventory-generator/1.0",
        },
    )
    last_error: Exception | None = None
    for attempt in range(3):
        try:
            with urllib.request.urlopen(request, timeout=30) as response:
                data = response.read()
            break
        except Exception as exc:
            last_error = exc
            if attempt == 2:
                raise
            time.sleep(1 + attempt)
            log(verbose, f"retry {url} after {exc}")
    else:
        raise RuntimeError(f"failed to fetch {url}: {last_error}")
    try:
        payload = json.loads(data)
    except json.JSONDecodeError as exc:
        raise RuntimeError(f"failed to parse JSON from {url}: {exc}") from exc
    if not isinstance(payload, dict):
        raise RuntimeError(f"unexpected top-level JSON type from {url}: {type(payload).__name__}")
    return payload


def normalize_json_url(url: str) -> str | None:
    parsed = urllib.parse.urlparse(url)
    if parsed.scheme in {"http", "https"}:
        if parsed.netloc != APPLE_DOCS_HOST:
            return None
        path = parsed.path
    elif not parsed.scheme and url.startswith("/"):
        path = url
    else:
        return None

    path = path.rstrip("/")
    if path.startswith("/tutorials/data/documentation/"):
        if path.endswith(".json"):
            return f"https://{APPLE_DOCS_HOST}{path}"
        return f"https://{APPLE_DOCS_HOST}{path}.json"
    if not path.startswith("/documentation/"):
        return None

    slug = path.removeprefix("/documentation/").strip("/")
    if not slug:
        return None
    return APPLE_DOCS_JSON_PREFIX + slug + ".json"


def fallback_json_url_from_identifier(identifier: str) -> str | None:
    if not isinstance(identifier, str):
        return None
    prefix = "doc://com.apple.appleads/documentation/"
    if not identifier.startswith(prefix):
        return None

    suffix = identifier.removeprefix(prefix).strip("/")
    parts = suffix.split("/")
    if len(parts) < 2:
        return None
    slug = parts[-1].split("#", 1)[0].lower()
    return APPLE_DOCS_JSON_PREFIX + "apple_ads/" + slug + ".json"


def resolve_reference_json_url(doc: dict, identifier: str, allowed: str) -> str | None:
    if not isinstance(identifier, str):
        return None
    references = doc.get("references")
    if isinstance(references, dict):
        ref = references.get(identifier)
        if isinstance(ref, dict):
            ref_url = ref.get("url")
            if isinstance(ref_url, str):
                normalized = normalize_json_url(ref_url)
                if normalized and normalized.startswith(allowed):
                    return normalized
    fallback = fallback_json_url_from_identifier(identifier)
    if fallback and fallback.startswith(allowed):
        return fallback
    return None


def extract_latest_changelog_identifier(root_doc: dict) -> tuple[str, int]:
    changelog_sections = [
        section
        for section in root_doc.get("topicSections", [])
        if isinstance(section, dict)
        and isinstance(section.get("title"), str)
        and section["title"].strip().lower() == "changelog"
    ]
    if not changelog_sections:
        raise RuntimeError("could not find the Apple Ads changelog topic section")

    candidates: list[tuple[int, str]] = []
    for section in changelog_sections:
        identifiers = section.get("identifiers")
        if not isinstance(identifiers, list):
            continue
        for identifier in identifiers:
            if not isinstance(identifier, str):
                continue
            match = re.search(r"apple-search-ads-campaign-management-api-(\d+)(?:[#/]|$)", identifier)
            if match:
                candidates.append((int(match.group(1)), identifier))

    if not candidates:
        raise RuntimeError("could not find a versioned Apple Ads changelog identifier")
    major, identifier = max(candidates, key=lambda item: item[0])
    return identifier, major


def extract_api_version_from_changelog_doc(doc: dict, source_url: str, major: int) -> str:
    version_pattern = re.compile(rf"^{major}\.\d+(?:\.\d+)*$")
    for section in doc.get("primaryContentSections", []):
        if not isinstance(section, dict) or section.get("kind") != "content":
            continue
        content = section.get("content")
        if not isinstance(content, list):
            continue
        for item in content:
            if not isinstance(item, dict):
                continue
            item_kind = item.get("kind") or item.get("type")
            text = item.get("text")
            if item_kind == "heading" and isinstance(text, str):
                normalized = text.strip()
                if version_pattern.fullmatch(normalized):
                    return normalized

    raise RuntimeError(
        f"could not extract Apple Ads API version from changelog {source_url}; "
        f"expected a heading like {major}.x"
    )


def extract_api_doc_version(root_doc: dict, allowed: str, verbose: bool = False) -> tuple[str, str]:
    identifier, major = extract_latest_changelog_identifier(root_doc)
    changelog_url = resolve_reference_json_url(root_doc, identifier, allowed)
    if not changelog_url:
        raise RuntimeError(f"could not resolve Apple Ads changelog identifier {identifier!r}")

    changelog_doc = fetch_json(changelog_url, verbose)
    version = extract_api_version_from_changelog_doc(changelog_doc, changelog_url, major)
    log(verbose, f"detected Apple Ads API docs version {version} from {changelog_url}")
    return version, changelog_url


def extract_topic_identifiers(doc: dict) -> list[str]:
    metadata = doc.get("metadata")
    role = metadata.get("role") if isinstance(metadata, dict) else None

    identifiers: list[str] = []
    for section in doc.get("topicSections", []):
        if not isinstance(section, dict):
            continue
        title = section.get("title")
        if role == "collectionGroup" and isinstance(title, str) and "endpoint" not in title.lower():
            continue
        raw_identifiers = section.get("identifiers")
        if not isinstance(raw_identifiers, list):
            continue
        for identifier in raw_identifiers:
            if isinstance(identifier, str):
                identifiers.append(identifier)
    return identifiers


def is_endpoint_doc(doc: dict) -> bool:
    metadata = doc.get("metadata")
    if isinstance(metadata, dict) and metadata.get("symbolKind") == "httpRequest":
        return True
    for section in doc.get("primaryContentSections", []):
        if isinstance(section, dict) and section.get("kind") == "restEndpoint":
            return True
    return False


def parse_method_path(doc: dict, source_url: str) -> tuple[str, str]:
    for section in doc.get("primaryContentSections", []):
        if not isinstance(section, dict) or section.get("kind") != "restEndpoint":
            continue

        tokens = section.get("tokens")
        if not isinstance(tokens, list) or not tokens:
            break

        method = None
        endpoint_text_parts: list[str] = []
        for token in tokens:
            if not isinstance(token, dict):
                continue
            text = token.get("text")
            if not isinstance(text, str):
                continue
            if token.get("kind") == "method":
                method = text.strip().upper()
                continue
            endpoint_text_parts.append(text)

        if not method:
            raise RuntimeError(f"missing HTTP method in restEndpoint tokens for {source_url}")

        endpoint_text = "".join(endpoint_text_parts).strip()
        if not endpoint_text:
            raise RuntimeError(f"missing endpoint URL text in restEndpoint tokens for {source_url}")

        parsed = urllib.parse.urlsplit(endpoint_text)
        raw_path = parsed.path or endpoint_text
        raw_path = re.sub(r"^/api/v\d+", "", raw_path)
        raw_path = re.sub(r"^v\d+", "", raw_path)
        if not raw_path.startswith("/"):
            raw_path = "/" + raw_path.lstrip("/")
        if parsed.query:
            raw_path += "?" + parsed.query
        return method, raw_path

    raise RuntimeError(f"no restEndpoint section found in endpoint doc {source_url}")


def section_by_kind(doc: dict, kind: str) -> dict | None:
    for section in doc.get("primaryContentSections", []):
        if isinstance(section, dict) and section.get("kind") == kind:
            return section
    return None


def type_info(entry: dict | None) -> dict | None:
    if not isinstance(entry, dict):
        return None
    return {
        "name": entry.get("text"),
        "identifier": entry.get("identifier"),
        "preciseIdentifier": entry.get("preciseIdentifier"),
    }


def doc_type_info(doc: dict) -> dict:
    metadata = doc.get("metadata") if isinstance(doc.get("metadata"), dict) else {}
    identifier = doc.get("identifier") if isinstance(doc.get("identifier"), dict) else {}
    return {
        "name": metadata.get("title"),
        "identifier": identifier.get("url"),
        "preciseIdentifier": metadata.get("externalID"),
    }


def abstract_text(doc: dict) -> str:
    parts: list[str] = []
    for item in doc.get("abstract", []):
        if isinstance(item, dict):
            text = item.get("text")
            if isinstance(text, str) and text.strip():
                parts.append(text.strip())
    return " ".join(parts)


def inline_content_text(content_items: list) -> str:
    parts: list[str] = []
    for content in content_items:
        if not isinstance(content, dict):
            continue
        inline = content.get("inlineContent")
        if not isinstance(inline, list):
            continue
        text = "".join(
            piece.get("text") or piece.get("code") or ""
            for piece in inline
            if isinstance(piece, dict)
        )
        if text.strip():
            parts.append(text.strip())
    return " ".join(parts)


def parse_parameters(doc: dict) -> list[dict]:
    section = section_by_kind(doc, "restParameters")
    if not section:
        return []

    source = section.get("source")
    items = section.get("items")
    if not isinstance(items, list):
        return []

    params: list[dict] = []
    for item in items:
        if not isinstance(item, dict):
            continue
        name = item.get("name")
        if not isinstance(name, str):
            continue
        type_name = None
        raw_types = item.get("type")
        if isinstance(raw_types, list) and raw_types:
            first = raw_types[0]
            if isinstance(first, dict):
                type_name = first.get("text")
        params.append(
            {
                "name": name,
                "in": source or "path",
                "required": bool(item.get("required", False)),
                "type": type_name,
                "description": inline_content_text(item.get("content", [])),
            }
        )
    return params


def parse_request_body(doc: dict) -> dict | None:
    section = section_by_kind(doc, "restBody")
    if not section:
        return None

    body_type = None
    raw_types = section.get("bodyContentType")
    if isinstance(raw_types, list) and raw_types:
        first = raw_types[0]
        if isinstance(first, dict):
            body_type = type_info(first)

    return {
        "contentType": section.get("mimeType") or "application/json",
        "type": body_type,
        "description": inline_content_text(section.get("content", [])),
    }


def parse_responses(doc: dict) -> list[dict]:
    section = section_by_kind(doc, "restResponses")
    if not section:
        return []

    items = section.get("items")
    if not isinstance(items, list):
        return []

    responses: list[dict] = []
    for item in items:
        if not isinstance(item, dict):
            continue
        status = item.get("status")
        if not isinstance(status, int):
            continue

        response_type = None
        raw_types = item.get("type")
        if isinstance(raw_types, list) and raw_types:
            first = raw_types[0]
            if isinstance(first, dict):
                response_type = type_info(first)

        description_parts: list[str] = []
        reason = item.get("reason")
        if isinstance(reason, str) and reason.strip():
            description_parts.append(reason.strip())
        description = inline_content_text(item.get("content", []))
        if description:
            description_parts.append(description)

        responses.append(
            {
                "status": status,
                "contentType": item.get("mimeType") or "application/json",
                "type": response_type,
                "description": " ".join(description_parts),
            }
        )
    return responses


def parse_endpoint_doc(doc: dict, source_url: str) -> dict:
    method, path = parse_method_path(doc, source_url)
    metadata = doc.get("metadata") if isinstance(doc.get("metadata"), dict) else {}
    title = metadata.get("title") or ""
    external_id = metadata.get("externalID") or ""

    return {
        "method": method,
        "path": path,
        "title": title,
        "summary": title,
        "description": abstract_text(doc),
        "source_url": source_url,
        "identifier": doc.get("identifier", {}).get("url") if isinstance(doc.get("identifier"), dict) else None,
        "external_id": external_id,
        "parameters": parse_parameters(doc),
        "request_body": parse_request_body(doc),
        "responses": parse_responses(doc),
    }


def attribute_values(attributes: list, kind: str) -> list[str]:
    values: list[str] = []
    if not isinstance(attributes, list):
        return values
    for attr in attributes:
        if not isinstance(attr, dict) or attr.get("kind") != kind:
            continue
        raw_values = attr.get("values")
        if isinstance(raw_values, list):
            for value in raw_values:
                if not isinstance(value, str):
                    continue
                values.extend(part.strip() for part in value.split(",") if part.strip())
        raw_value = attr.get("value")
        if isinstance(raw_value, str) and raw_value.strip():
            values.append(raw_value.strip())
    return values


def type_tokens_text(tokens: list) -> str:
    return "".join(
        token.get("text", "")
        for token in tokens
        if isinstance(token, dict) and isinstance(token.get("text"), str)
    ).strip()


def schema_for_type_tokens(tokens: list) -> tuple[dict, list[dict]]:
    if not isinstance(tokens, list) or not tokens:
        return {"type": "object"}, []

    refs = [type_info(token) for token in tokens if isinstance(token, dict) and token.get("kind") == "typeIdentifier"]
    refs = [ref for ref in refs if ref is not None]
    rendered = type_tokens_text(tokens)

    if rendered.startswith("[") and rendered.endswith("]"):
        if refs:
            return {"type": "array", "items": component_ref(refs[0])}, refs
        inner = rendered[1:-1].strip()
        return {"type": "array", "items": schema_for_scalar(inner)}, []

    if refs:
        return component_ref(refs[0]), refs

    if rendered:
        return schema_for_scalar(rendered), []
    return {"type": "object"}, []


def component_ref(type_obj: dict | None) -> dict:
    name = schema_name(type_obj)
    if not name:
        return {"type": "object"}
    return {"$ref": f"#/components/schemas/{name}"}


def parse_schema_doc(doc: dict) -> tuple[dict, list[dict]]:
    schema_doc_type = doc_type_info(doc)
    metadata = doc.get("metadata") if isinstance(doc.get("metadata"), dict) else {}
    schema: dict = {
        "name": schema_name(schema_doc_type),
        "title": metadata.get("title") or schema_name(schema_doc_type),
        "type": "object",
        "identifier": schema_doc_type.get("identifier"),
        "preciseIdentifier": schema_doc_type.get("preciseIdentifier"),
        "description": abstract_text(doc),
        "properties": [],
        "enum": [],
    }
    refs: list[dict] = []

    properties_section = section_by_kind(doc, "properties")
    if isinstance(properties_section, dict):
        required: list[str] = []
        for item in properties_section.get("items", []):
            if not isinstance(item, dict) or not isinstance(item.get("name"), str):
                continue
            prop_schema, prop_refs = schema_for_type_tokens(item.get("type", []))
            refs.extend(prop_refs)
            enum_values = attribute_values(item.get("attributes", []), "allowedValues")
            if enum_values and prop_schema.get("type") == "string":
                prop_schema = {**prop_schema, "enum": enum_values}
            default_values = attribute_values(item.get("attributes", []), "default")
            prop: dict = {
                "name": item["name"],
                "schema": prop_schema,
                "description": inline_content_text(item.get("content", [])),
            }
            if default_values:
                prop["default"] = default_values[0]
            schema["properties"].append(prop)
            if item.get("required") is True:
                required.append(item["name"])
        if required:
            schema["required"] = required

    possible_values = section_by_kind(doc, "possibleValues")
    if isinstance(possible_values, dict):
        values = [
            item.get("name")
            for item in possible_values.get("values", [])
            if isinstance(item, dict) and isinstance(item.get("name"), str)
        ]
        schema["enum"] = values
        if values:
            schema["type"] = "string"

    return schema, refs


def type_refs_from_endpoint(endpoint: dict) -> list[dict]:
    refs: list[dict] = []
    request_body = endpoint.get("request_body")
    if isinstance(request_body, dict) and isinstance(request_body.get("type"), dict):
        refs.append(request_body["type"])
    for response in endpoint.get("responses", []):
        if isinstance(response, dict) and isinstance(response.get("type"), dict):
            refs.append(response["type"])
    return refs


def scrape_inventory(docs_url: str, verbose: bool = False) -> tuple[dict, list[str], dict]:
    root_url = root_json_url(docs_url)
    allowed = allowed_prefix(root_url)
    root_doc = fetch_json(root_url, verbose)

    if not isinstance(root_doc.get("references"), dict):
        raise RuntimeError(
            f"{root_url} is missing a references object; Apple's JSON structure may have changed"
        )
    if not isinstance(root_doc.get("topicSections"), list) or not root_doc["topicSections"]:
        raise RuntimeError(
            f"{root_url} is missing topicSections; Apple's JSON structure may have changed"
        )

    api_version, api_version_source_url = extract_api_doc_version(
        root_doc,
        allowed,
        verbose,
    )

    queue: deque[str] = deque([root_url])
    seen_urls: set[str] = set()
    endpoint_entries: list[dict] = []
    unresolved_identifiers: list[tuple[str, str]] = []
    parse_failures: list[tuple[str, str]] = []
    fetched_docs = 0
    endpoint_docs = 0

    while queue:
        url = queue.popleft()
        if url in seen_urls:
            continue
        if not url.startswith(allowed):
            log(verbose, f"skip out-of-scope {url}")
            continue

        doc = fetch_json(url, verbose)
        seen_urls.add(url)
        fetched_docs += 1

        if is_endpoint_doc(doc):
            endpoint_docs += 1
            try:
                endpoint_entries.append(parse_endpoint_doc(doc, url))
                log(
                    verbose,
                    f"endpoint {endpoint_entries[-1]['method']} {endpoint_entries[-1]['path']}",
                )
            except RuntimeError as exc:
                parse_failures.append((url, str(exc)))
                log(verbose, f"parse failure {url}: {exc}")
            continue

        for identifier in extract_topic_identifiers(doc):
            child_url = resolve_reference_json_url(doc, identifier, allowed)
            if child_url:
                queue.append(child_url)
                log(verbose, f"queue {child_url}")
            else:
                unresolved_identifiers.append((url, identifier))
                log(verbose, f"unresolved {identifier} from {url}")

    if fetched_docs < 10:
        raise RuntimeError(f"only fetched {fetched_docs} docs from Apple; traversal likely failed")
    if endpoint_docs == 0 or not endpoint_entries:
        raise RuntimeError("found no endpoint documents in Apple's JSON docs")
    if len(endpoint_entries) < 20:
        raise RuntimeError(
            f"only extracted {len(endpoint_entries)} endpoint documents; parsing rules likely need updates"
        )
    if parse_failures and len(parse_failures) > max(3, endpoint_docs // 4):
        sample = "; ".join(f"{url}: {msg}" for url, msg in parse_failures[:3])
        raise RuntimeError(
            f"failed to parse too many endpoint docs ({len(parse_failures)}/{endpoint_docs}): {sample}"
        )

    endpoint_entries.sort(key=lambda item: (item["path"], item["method"]))
    paths = sorted({f'{item["method"]} {item["path"]}' for item in endpoint_entries})

    schema_docs: dict[str, dict] = {}
    schema_parse_failures: list[tuple[str, str]] = []
    type_queue: deque[tuple[dict, dict]] = deque()
    seen_type_urls: set[str] = set()
    for endpoint in endpoint_entries:
        for type_ref in type_refs_from_endpoint(endpoint):
            type_queue.append((type_ref, root_doc))

    while type_queue:
        type_obj, source_doc = type_queue.popleft()
        type_url = resolve_reference_json_url(source_doc, type_obj.get("identifier", ""), allowed)
        if not type_url:
            type_url = resolve_reference_json_url(root_doc, type_obj.get("identifier", ""), allowed)
        if not type_url:
            type_url = fallback_json_url_from_identifier(type_obj.get("identifier", "") or "")
        if not type_url or not type_url.startswith(allowed) or type_url in seen_type_urls:
            continue
        seen_type_urls.add(type_url)
        try:
            doc = fetch_json(type_url, verbose)
            parsed_schema, nested_refs = parse_schema_doc(doc)
        except Exception as exc:
            schema_parse_failures.append((type_url, str(exc)))
            log(verbose, f"schema parse failure {type_url}: {exc}")
            continue
        name = parsed_schema.get("name")
        if isinstance(name, str) and name:
            schema_docs[name] = parsed_schema
            log(verbose, f"schema {name}")
        for nested_ref in nested_refs:
            type_queue.append((nested_ref, doc))

    inventory = {
        "sources": {
            "official_docs_url": docs_url,
            "root_json_url": root_url,
            "allowed_prefix": allowed,
            "api_version_source_url": api_version_source_url,
        },
        "generation": {
            "api_version": api_version,
            "fetched_documents": fetched_docs,
            "endpoint_documents": endpoint_docs,
            "extracted_paths": len(paths),
            "unresolved_topic_identifiers": len(unresolved_identifiers),
            "endpoint_parse_failures": len(parse_failures),
            "schema_documents": len(schema_docs),
            "schema_parse_failures": len(schema_parse_failures),
        },
        "endpoints": endpoint_entries,
        "schemas": schema_docs,
    }
    diagnostics = {
        "root_json_url": root_url,
        "allowed_prefix": allowed,
        "api_version": api_version,
        "api_version_source_url": api_version_source_url,
        "fetched_documents": fetched_docs,
        "endpoint_documents": endpoint_docs,
        "unresolved_identifiers": unresolved_identifiers[:10],
        "parse_failures": parse_failures[:10],
        "schema_documents": len(schema_docs),
        "schema_parse_failures": schema_parse_failures[:10],
    }
    return inventory, paths, diagnostics


def schema_name(type_obj: dict | None) -> str | None:
    if not isinstance(type_obj, dict):
        return None
    precise = type_obj.get("preciseIdentifier")
    if isinstance(precise, str) and precise:
        return precise.rsplit(":", 1)[-1]
    identifier = type_obj.get("identifier")
    if isinstance(identifier, str) and identifier:
        return identifier.rstrip("/").split("/")[-1].replace("-", "_")
    name = type_obj.get("name")
    if isinstance(name, str) and name:
        return re.sub(r"[^A-Za-z0-9_]+", "_", name)
    return None


def schema_for_scalar(type_name: str | None) -> dict:
    normalized = (type_name or "").strip().lower()
    if normalized in {"int64", "integer", "int"}:
        return {"type": "integer", "format": "int64"}
    if normalized in {"double", "float", "number"}:
        return {"type": "number"}
    if normalized in {"boolean", "bool"}:
        return {"type": "boolean"}
    return {"type": "string"}


def operation_id(endpoint: dict) -> str:
    external_id = endpoint.get("external_id")
    if isinstance(external_id, str) and external_id:
        cleaned = re.sub(r"[^A-Za-z0-9]+", "_", external_id).strip("_")
        if cleaned:
            return cleaned
    title = endpoint.get("title") or f'{endpoint["method"]}_{endpoint["path"]}'
    return re.sub(r"[^A-Za-z0-9]+", "_", title).strip("_")


def normalize_version_label(version: str) -> str:
    version = version.strip()
    if not version:
        raise ValueError("api version cannot be empty")
    if version.startswith("v"):
        version = version[1:]
    if not re.fullmatch(r"\d+\.\d+(?:\.\d+)*", version):
        raise ValueError(f"invalid api version {version!r}: use a minor version like 5.5")
    return version


def versioned_openapi_output(path: Path, api_version: str) -> Path:
    normalized = normalize_version_label(api_version)
    return path.with_name(f"openapi-v{normalized}{path.suffix}")


def latest_openapi_link(path: Path) -> Path:
    return path.with_name(f"openapi-latest{path.suffix}")


def openapi_component_from_schema_doc(schema_doc: dict) -> dict:
    component: dict = {
        "type": schema_doc.get("type") or "object",
        "title": schema_doc.get("title") or schema_doc.get("name"),
        "x-apple-identifier": schema_doc.get("identifier"),
        "x-apple-preciseIdentifier": schema_doc.get("preciseIdentifier"),
    }
    description = schema_doc.get("description")
    if isinstance(description, str) and description:
        component["description"] = description

    enum_values = schema_doc.get("enum")
    if isinstance(enum_values, list) and enum_values:
        component["enum"] = enum_values

    properties = schema_doc.get("properties")
    if isinstance(properties, list) and properties:
        component["properties"] = {}
        for prop in properties:
            if not isinstance(prop, dict) or not isinstance(prop.get("name"), str):
                continue
            prop_schema = prop.get("schema") if isinstance(prop.get("schema"), dict) else {"type": "object"}
            prop_component = dict(prop_schema)
            prop_description = prop.get("description")
            if isinstance(prop_description, str) and prop_description:
                prop_component["description"] = prop_description
            if "default" in prop:
                prop_component["default"] = prop["default"]
            component["properties"][prop["name"]] = prop_component

    required = schema_doc.get("required")
    if isinstance(required, list) and required:
        component["required"] = [item for item in required if isinstance(item, str)]

    return component


def build_openapi(inventory: dict, api_version: str) -> dict:
    normalized_version = normalize_version_label(api_version)
    components: dict[str, dict] = {}
    paths: dict[str, dict] = {}
    parsed_schemas = inventory.get("schemas")
    if not isinstance(parsed_schemas, dict):
        parsed_schemas = {}

    def ensure_component(type_obj: dict | None) -> dict:
        name = schema_name(type_obj)
        if not name:
            return {"type": "object"}
        if name not in components:
            parsed_schema = parsed_schemas.get(name)
            if isinstance(parsed_schema, dict):
                components[name] = openapi_component_from_schema_doc(parsed_schema)
            else:
                components[name] = {
                    "type": "object",
                    "title": type_obj.get("name") or name,
                    "x-apple-identifier": type_obj.get("identifier"),
                    "x-apple-preciseIdentifier": type_obj.get("preciseIdentifier"),
                }
        return {"$ref": f"#/components/schemas/{name}"}

    for name, parsed_schema in sorted(parsed_schemas.items()):
        if isinstance(name, str) and isinstance(parsed_schema, dict):
            components.setdefault(name, openapi_component_from_schema_doc(parsed_schema))

    for endpoint in inventory["endpoints"]:
        path_item = paths.setdefault(endpoint["path"], {})
        op: dict = {
            "operationId": operation_id(endpoint),
            "summary": endpoint.get("summary") or endpoint.get("title") or "",
            "description": endpoint.get("description") or "",
            "x-source-url": endpoint.get("source_url"),
        }

        parameters = []
        for param in endpoint.get("parameters", []):
            parameters.append(
                {
                    "name": param["name"],
                    "in": param.get("in") or "path",
                    "required": bool(param.get("required")) or param.get("in") == "path",
                    "description": param.get("description") or "",
                    "schema": schema_for_scalar(param.get("type")),
                }
            )
        if parameters:
            op["parameters"] = parameters

        request_body = endpoint.get("request_body")
        if isinstance(request_body, dict):
            content_type = request_body.get("contentType") or "application/json"
            op["requestBody"] = {
                "required": True,
                "description": request_body.get("description") or "",
                "content": {
                    content_type: {
                        "schema": ensure_component(request_body.get("type"))
                    }
                },
            }

        responses = {}
        for response in endpoint.get("responses", []):
            content_type = response.get("contentType")
            content = {}
            if content_type:
                content = {
                    content_type: {
                        "schema": ensure_component(response.get("type"))
                    }
                }
            responses[str(response["status"])] = {
                "description": response.get("description") or "",
                **({"content": content} if content else {}),
            }
        if not responses:
            responses = {"default": {"description": "Response"}}
        op["responses"] = responses

        path_item[endpoint["method"].lower()] = op

    return {
        "openapi": "3.0.3",
        "info": {
            "title": "Apple Ads",
            "version": normalized_version,
            "description": "Lightweight OpenAPI generated from Apple Ads documentation JSON.",
        },
        "servers": [{"url": "https://api.searchads.apple.com/api/v5"}],
        "paths": paths,
        "components": {"schemas": components},
    }


def load_existing_paths() -> list[str]:
    return []


def load_existing_paths_from_file(path_file: Path | None) -> list[str]:
    if path_file is None or not path_file.exists():
        return []
    return [line for line in path_file.read_text().splitlines() if line.strip()]


def diff_summary(old_paths: list[str], new_paths: list[str]) -> tuple[int, int, int, float]:
    old_set = set(old_paths)
    new_set = set(new_paths)
    removed = len(old_set - new_set)
    added = len(new_set - old_set)
    changed = added + removed
    baseline = max(len(old_set), len(new_set), 1)
    percent = (changed / baseline) * 100
    return added, removed, changed, percent


def print_path_diff(old_paths: list[str], new_paths: list[str], label: str) -> None:
    diff = difflib.unified_diff(
        [line + "\n" for line in old_paths],
        [line + "\n" for line in new_paths],
        fromfile=label,
        tofile=label,
    )
    sys.stdout.writelines(diff)


def prompt_yes_no(question: str, default: bool = False) -> bool:
    suffix = "[Y/n]" if default else "[y/N]"
    reply = input(f"{question} {suffix} ").strip().lower()
    if not reply:
        return default
    return reply in {"y", "yes"}


def render_json(value: dict) -> str:
    return json.dumps(value, indent=2) + "\n"


def render_paths(paths: list[str]) -> str:
    return "\n".join(paths) + "\n"


def is_latest_openapi_up_to_date(latest_link: Path, openapi_output: Path, rendered_openapi: str) -> bool:
    if not latest_link.exists() and not latest_link.is_symlink():
        return False
    if latest_link.is_symlink():
        expected_target = os.path.relpath(openapi_output, latest_link.parent)
        return os.readlink(latest_link) == expected_target
    try:
        return latest_link.read_text(encoding="utf-8") == rendered_openapi
    except OSError:
        return False


def check_output_file(path: Path, expected: str, stale_outputs: list[Path]) -> None:
    try:
        current = path.read_text(encoding="utf-8") if path.exists() else ""
    except OSError:
        current = ""
    if current != expected:
        stale_outputs.append(path)


def check_outputs(
    inventory: dict,
    paths: list[str],
    openapi: dict,
    openapi_output: Path,
    inventory_output: Path | None,
    paths_output: Path | None,
) -> None:
    rendered_openapi = render_json(openapi)
    stale_outputs: list[Path] = []

    check_output_file(openapi_output, rendered_openapi, stale_outputs)

    latest_link = latest_openapi_link(openapi_output)
    if not is_latest_openapi_up_to_date(latest_link, openapi_output, rendered_openapi):
        stale_outputs.append(latest_link)

    if inventory_output is not None:
        check_output_file(inventory_output, render_json(inventory), stale_outputs)

    if paths_output is not None:
        check_output_file(paths_output, render_paths(paths), stale_outputs)

    if stale_outputs:
        print("Generated Apple Ads OpenAPI artifacts are out of date:", file=sys.stderr)
        for path in stale_outputs:
            print(f"  - {path}", file=sys.stderr)
        print("Run `make openapi` to regenerate them.", file=sys.stderr)
        raise SystemExit(1)


def write_outputs(
    inventory: dict,
    paths: list[str],
    openapi: dict,
    openapi_output: Path,
    inventory_output: Path | None,
    paths_output: Path | None,
) -> None:
    openapi_output.parent.mkdir(parents=True, exist_ok=True)
    rendered_openapi = render_json(openapi)
    openapi_output.write_text(rendered_openapi)
    print(f"Wrote {openapi_output}")

    latest_link = latest_openapi_link(openapi_output)
    if latest_link.exists() or latest_link.is_symlink():
        latest_link.unlink()
    target_name = os.path.relpath(openapi_output, latest_link.parent)
    latest_link.symlink_to(target_name)
    print(f"Linked {latest_link} -> {target_name}")

    if inventory_output is not None:
        inventory_output.parent.mkdir(parents=True, exist_ok=True)
        inventory_output.write_text(render_json(inventory))
        print(f"Wrote {inventory_output}")

    if paths_output is not None:
        paths_output.parent.mkdir(parents=True, exist_ok=True)
        paths_output.write_text(render_paths(paths))
        print(f"Wrote {paths_output}")


def main() -> None:
    args = parse_args()
    requested_openapi_output = Path(args.openapi_output).expanduser()
    inventory_output = Path(args.output_inventory).expanduser() if args.output_inventory else None
    paths_output = Path(args.output_paths).expanduser() if args.output_paths else None

    try:
        inventory, new_paths, diagnostics = scrape_inventory(args.docs_url, args.verbose)
    except Exception as exc:
        print(f"Error: {exc}", file=sys.stderr)
        sys.exit(1)

    try:
        api_version = normalize_version_label(diagnostics["api_version"])
        if args.api_version is not None:
            requested_api_version = normalize_version_label(args.api_version)
            if requested_api_version != api_version:
                print(
                    "Error: specified API docs version "
                    f"{requested_api_version} does not match Apple's crawled version {api_version} "
                    f"from {diagnostics['api_version_source_url']}",
                    file=sys.stderr,
                )
                sys.exit(1)
    except ValueError as exc:
        print(f"Error: {exc}", file=sys.stderr)
        sys.exit(1)

    openapi_output = versioned_openapi_output(requested_openapi_output, api_version)
    openapi = build_openapi(inventory, api_version)
    old_paths = load_existing_paths_from_file(paths_output)
    added, removed, changed, percent = diff_summary(old_paths, new_paths)
    diff_label = str(paths_output) if paths_output is not None else "paths"

    print(f"Using Apple Ads docs URL: {args.docs_url}")
    print(f"Using API docs version label: {api_version}")
    print(f"Detected API docs version from: {diagnostics['api_version_source_url']}")
    print(f"Resolved docs JSON root: {diagnostics['root_json_url']}")
    print(f"Crawl prefix: {diagnostics['allowed_prefix']}")
    print(
        f"Fetched {diagnostics['fetched_documents']} docs; "
        f"parsed {diagnostics['endpoint_documents']} endpoint docs; "
        f"parsed {diagnostics['schema_documents']} schema docs; "
        f"extracted {len(new_paths)} unique paths"
    )
    if diagnostics["unresolved_identifiers"]:
        print(
            f"Warning: could not resolve {len(diagnostics['unresolved_identifiers'])} topic identifiers; "
            "showing up to 10 samples:"
        )
        for source_url, identifier in diagnostics["unresolved_identifiers"]:
            print(f"  - {identifier} (from {source_url})")
    if diagnostics["parse_failures"]:
        print(
            f"Warning: failed to parse {len(diagnostics['parse_failures'])} endpoint docs; "
            "showing up to 10 samples:"
        )
        for url, message in diagnostics["parse_failures"]:
            print(f"  - {url}: {message}")
    if diagnostics["schema_parse_failures"]:
        print(
            f"Warning: failed to parse {len(diagnostics['schema_parse_failures'])} schema docs; "
            "showing up to 10 samples:"
        )
        for url, message in diagnostics["schema_parse_failures"]:
            print(f"  - {url}: {message}")

    if changed == 0:
        print("No path changes detected.")
        if args.check:
            log(args.verbose, "checking outputs")
            check_outputs(
                inventory,
                new_paths,
                openapi,
                openapi_output,
                inventory_output,
                paths_output,
            )
            return
        log(args.verbose, "writing outputs")
        write_outputs(
            inventory,
            new_paths,
            openapi,
            openapi_output,
            inventory_output,
            paths_output,
        )
        return

    print(
        f"Detected {changed} path changes: {added} added, {removed} removed "
        f"({percent:.2f}% of the API surface)"
    )

    if args.check:
        print("Path diff:")
        print_path_diff(old_paths, new_paths, diff_label)
        print("Generated Apple Ads OpenAPI artifacts are out of date.", file=sys.stderr)
        print("Run `make openapi` to regenerate them.", file=sys.stderr)
        sys.exit(1)

    if args.yes:
        log(args.verbose, "auto-accepting changes due to --yes")
        write_outputs(
            inventory,
            new_paths,
            openapi,
            openapi_output,
            inventory_output,
            paths_output,
        )
        return

    if percent > 2.0:
        print("Path diff:")
        print_path_diff(old_paths, new_paths, diff_label)
        if not sys.stdin.isatty():
            print(
                "Error: refusing to write because more than 2% of paths changed and no TTY is available. "
                "Re-run with --yes to accept.",
                file=sys.stderr,
            )
            sys.exit(1)
        if not prompt_yes_no("More than 2% of the API changed. Write updated files?"):
            print("Aborted.")
            sys.exit(1)
        log(args.verbose, "writing outputs after confirmation")
        write_outputs(
            inventory,
            new_paths,
            openapi,
            openapi_output,
            inventory_output,
            paths_output,
        )
        return

    if sys.stdin.isatty() and prompt_yes_no("Show path diff before writing?"):
        print("Path diff:")
        print_path_diff(old_paths, new_paths, diff_label)

    log(args.verbose, "writing outputs")
    write_outputs(
        inventory,
        new_paths,
        openapi,
        openapi_output,
        inventory_output,
        paths_output,
    )


if __name__ == "__main__":
    main()
