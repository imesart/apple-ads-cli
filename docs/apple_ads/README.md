# Apple Ads Inventory

This folder contains generated Apple Ads API artifacts scraped from Apple's documentation JSON.

## Current source inputs:

- Apple Developer documentation, defaulting to `https://developer.apple.com/documentation/apple_ads`

The official Apple Ads docs URL is the primary reference recorded in the generated inventory metadata.

## Generated files:

- `openapi-v5.5.json`: lightweight OpenAPI document derived from Apple's endpoint docs
- `openapi-latest.json`: symlink to the latest generated versioned OpenAPI file
- `paths.txt`: flattened list of known endpoint patterns

## Regenerate with:

```bash
python3 scripts/generate_openapi.py docs/apple_ads/openapi.json --output-paths docs/apple_ads/paths.txt
```

Check whether the generated files are current without rewriting them:

```bash
python3 scripts/generate_openapi.py docs/apple_ads/openapi.json --output-paths docs/apple_ads/paths.txt --check
```

Override the Apple docs URL if needed:

```bash
python3 scripts/generate_openapi.py docs/apple_ads/openapi.json --output-paths docs/apple_ads/paths.txt --docs-url https://developer.apple.com/documentation/apple_ads
```
