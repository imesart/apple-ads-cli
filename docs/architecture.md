# aads — Implementation Notes

This document explains how `aads` is structured and why key engineering decisions were made.
Use it as the reference for architecture, package boundaries, code organization, and implementation strategy.

## Goals

The project optimizes for:

- Single static binary distribution
- Strong terminal UX for interactive use
- Safe day-to-day operational use (budget limits, confirmation flags, secret redaction)
- First-class scripting support (stable JSON output, non-zero exit codes, no stdout noise)
- Clean internal separation between CLI, config/auth, and API client layers

## Reference Projects

This implementation draws from several existing projects, each contributing different strengths:

| Project | Language | What inspired us |
|---|---|---|
| [rudrankriyam/App-Store-Connect-CLI](https://github.com/rudrankriyam/App-Store-Connect-CLI) | Go | CLI architecture, command builders, retry infra, config, spec cache, release pipeline |
| [mihai8804858/swift-apple-search-ads-api](https://github.com/mihai8804858/swift-apple-search-ads-api) | Swift | Type definitions (96 models), snapshot test strategy, auth flow, plugin middleware |
| [phiture/searchads_api](https://github.com/phiture/searchads_api) | Python | Agency operational knowledge: defaults, API gotchas, report patterns, multi-org |
| [SaadBelfqih/apple-ads-cli](https://github.com/SaadBelfqih/apple-ads-cli) | Go | Endpoint mapping reference, doc generation tooling |
| [TrebuhS/Apple-Search-Ads-CLI](https://github.com/TrebuhS/Apple-Search-Ads-CLI) | Go | Safety limits, multi-profile support, service layer pattern |

## Language and Dependencies

### Language: Go

Go is chosen over Swift for CLI distribution advantages: static binaries, trivial cross-compilation, GitHub Actions-based release automation, and a mature CLI ecosystem (see `docs/requirements.md` for rationale).

### Go version

Go 1.24+ (might work with earlier versions).

### Direct dependencies

| Package | Purpose |
|---|---|
| `github.com/peterbourgon/ff/v3` | CLI framework |
| `github.com/golang-jwt/jwt/v5` | ES256 JWT signing for OAuth2 |
| `github.com/olekukonko/tablewriter` | Table output formatting |
| `gopkg.in/yaml.v3` | YAML config parsing and output |
| `golang.org/x/term` | TTY detection for output format defaults |

We use `ff/v3` instead of `cobra` as the extra features are not currently needed.

## Project Structure

This section is intentionally selective. It shows:

- major folders and package boundaries
- startup entry points and coordinating files that are useful for orientation
- workflow-defining docs and build entry points

This section intentionally does not show:

- test files
- every sibling file in large packages
- incidental/generated files unless they are part of the documented workflow contract

Use `...` to indicate that a directory contains additional files not listed here.

```
aads/
├── main.go                           # Program entry point; calls into cmd/root.go
├── Makefile                          # Primary build, test, docs, and release entry points
├── install.sh                        # Binary-download installer for GitHub Releases
│
├── cmd/
│   ├── root.go                       # Root command, global flags, top-level wiring
│   ├── exit_codes.go                 # Process exit code definitions
│   └── ...
│
├── internal/
│   ├── api/                          # HTTP client, middleware, retries, pagination
│   │   ├── client.go                 # Core API client and request execution
│   │   ├── middleware.go             # Request middleware chain
│   │   ├── retry.go                  # Retry helpers, including auth retry
│   │   ├── errors.go                 # API error parsing and helpers
│   │   ├── request.go                # Shared request interface
│   │   ├── requests/                 # Request types grouped by endpoint family
│   │   │   ├── campaigns/
│   │   │   ├── adgroups/
│   │   │   ├── keywords/
│   │   │   ├── negatives_campaign/
│   │   │   ├── negatives_adgroup/
│   │   │   ├── ads/
│   │   │   ├── creatives/
│   │   │   ├── budgetorders/
│   │   │   ├── product_pages/
│   │   │   ├── ad_rejections/
│   │   │   ├── reports/
│   │   │   ├── impression_share/
│   │   │   ├── apps/
│   │   │   ├── geo/
│   │   │   ├── acls/
│   │   │   └── ...
│   │   ├── testdata/
│   │   │   └── request_snapshots/     # Golden HTTP request snapshots
│   │   └── ...
│   │
│   ├── auth/                         # JWT signing, token exchange, token cache
│   │   ├── jwt.go
│   │   ├── oauth2.go
│   │   ├── token_cache.go
│   │   └── ...
│   │
│   ├── cli/                          # User-facing command groups and shared CLI helpers
│   │   ├── registry/                 # Command registration
│   │   ├── shared/                   # Shared flags, client setup, output, validation
│   │   ├── campaigns/
│   │   ├── adgroups/
│   │   ├── keywords/
│   │   ├── negatives/
│   │   ├── ads/
│   │   ├── creatives/
│   │   ├── budgetorders/
│   │   ├── productpages/
│   │   ├── adrejections/
│   │   ├── reports/
│   │   ├── impressionshare/
│   │   ├── apps/
│   │   ├── geo/
│   │   ├── acls/
│   │   ├── profiles/
│   │   ├── structure/
│   │   ├── schema/
│   │   ├── completion/
│   │   ├── version/
│   │   └── ...
│   │
│   ├── columnname/                   # Output column naming helpers
│   ├── config/                       # Config loading, env/profile resolution, decimal parsing
│   │   ├── config.go
│   │   ├── decimal_text.go
│   │   └── ...
│   │
│   ├── fieldmeta/                    # Field metadata used by selection/output logic
│   ├── output/                       # Output formats and field/selection helpers
│   │   ├── format.go
│   │   ├── json.go
│   │   ├── table.go
│   │   ├── yaml.go
│   │   ├── markdown.go
│   │   ├── ids.go
│   │   ├── select.go
│   │   └── ...
│   ├── testutil/                     # Shared test helpers for integration/live tests
│   ├── types/                        # Shared API/domain types
│   │   ├── response.go
│   │   ├── selector.go
│   │   ├── campaign.go
│   │   ├── adgroup.go
│   │   ├── keyword.go
│   │   ├── report.go
│   │   ├── token.go
│   │   └── ...
│   └── ...
│
├── docs/
│   ├── requirements.md               # Product and behavior requirements
│   ├── architecture.md               # Architecture, package boundaries, workflows
│   ├── coverage.md                   # Endpoint coverage status and path discrepancies
│   ├── commands.md                   # Generated command reference
│   ├── apple_ads/                    # Generated Apple Ads reference artifacts
│   ├── schemas/                      # Versioned JSON schemas
│   └── ...
│
├── scripts/
│   ├── generate_commands_doc.py      # Regenerates docs/commands.md from CLI help
│   ├── generate_openapi.py           # Regenerates Apple Ads reference artifacts
│   ├── create_release_artifacts.py   # Builds release archives and checksums
│   └── ...
│
├── .github/
│   └── workflows/
│       ├── release.yml               # Tagged release build and publish
│       ├── govulncheck.yml           # Vulnerability scanning
│       └── ...
│
├── AGENTS.md                         # Repository-specific coding-agent instructions
├── README.md                         # User-facing overview and quick start
├── go.mod
└── go.sum
```

## Separation of Concerns

### `internal/auth`

Owns credential handling: ES256 JWT assertion building, OAuth2 token exchange, and file-based token caching. Produces a valid access token for the API client. Must not know about CLI flags, profiles, or terminal output.

### `internal/api`

Owns the Apple Ads API client: request execution, response parsing, middleware chain, retry logic, and pagination. Must not know about CLI flags, config files, or output formatting.

### `internal/api/requests`

Owns request construction: one file per API endpoint, each implementing the `Request` interface (Method, Path, Body, Query). Pure data — no HTTP execution, no CLI awareness.

### `internal/types`

Owns API type definitions: domain models, enums, request/response shapes. Shared across `api`, `cli`, and `output`. Must not import any other internal package.

Use `internal/types` for stable API wire-format types that are shared across packages or represent canonical request/response shapes we want to name once and reuse consistently. Small endpoint-local helper structs are still acceptable when they are one-off partial decodes, raw JSON probes, or command-specific projections that are not useful as shared domain types. Do not force every anonymous struct into `internal/types`; move it there when the shape is part of the common API vocabulary or duplication starts appearing across packages.

### `internal/config`

Owns the persisted configuration format: loading, saving, profile resolution, env var overrides. Produces a resolved config struct. Must not know about CLI flags or terminal output.

### `internal/cli`

Owns command definitions, flag parsing, validation, and wiring. This layer should stay thin: parse flags, validate combinations, call `internal/api` services, print via `internal/output`. Must not contain API request logic or output rendering.

The `internal/cli/structure` package is the one intentional exception to “thin handlers”: it owns multi-step orchestration for structure export/import because the feature spans selection, normalization, validation, collision checks, mock-ID planning for `--check`, and ordered mutation sequencing across multiple endpoint families.

### `internal/output`

Owns rendering: JSON, table, YAML output formatting. Separate from command handlers so tests can validate data shape independently of presentation. Must not import `internal/cli`.

## API Coverage

Endpoint-level coverage is tracked in [`docs/coverage.md`](coverage.md). Update it when adding or removing request files under `internal/api/requests/`, or after regenerating the Apple Ads OpenAPI docs artifacts.

## Architecture Decisions

### One request file per endpoint

Each API endpoint gets its own file in `internal/api/requests/<group>/`. Each file is ~15-30 lines: a struct implementing a `RequestType` interface with method, path, and body.

```go
// internal/api/requests/campaigns/create.go
package campaigns

type CreateRequest struct {
    Campaign types.Campaign
}

func (r CreateRequest) Method() string      { return "POST" }
func (r CreateRequest) Path() string        { return "/api/v5/campaigns" }
func (r CreateRequest) Body() any           { return r.Campaign }
func (r CreateRequest) Query() url.Values   { return nil }
```

Rationale: when Apple adds an endpoint, you add one file. No merge conflicts, no editing a 300-line file. 67 small files are easier to maintain than 15 large ones.

### Type files: one per domain concept, not one per API type

We group related types. `campaign.go` contains `Campaign`, its nested enums (`CampaignStatus`, `CampaignDisplayStatus`, `CampaignServingStatus`, `CampaignAdChannelType`, etc.), and constants. `campaign_update.go` is separate because the update envelope (`{"campaign": {...}}`) is structurally different.

This is a pragmatic middle ground: fewer files than one per API type (~30 vs 96), but each file stays under 400 lines.

### Middleware chain for request preparation

Cross-cutting concerns are handled by a middleware chain rather than being baked into the transport:

```go
type Middleware func(r *http.Request) error

func NewClient(cfg *config.Config, tokenStore *auth.TokenStore) *Client {
    return &Client{
        middlewares: []Middleware{
            InjectHost("api.searchads.apple.com"),
            InjectAcceptHeaders(),
            InjectAuthorization(tokenStore),
            InjectOrgContext(cfg.OrgID),
        },
    }
}
```

Each middleware is independently testable. Adding a new header (e.g., for proxy support) is adding one middleware, not editing the HTTP transport.

### Dual retry strategy

Using:

```
withExponentialBackoff {      ← rate limit retry (429, 5xx): 2s → 4s → 8s → 16s
    withAuthRetry {           ← auth retry (401): refresh token, retry once
        perform(request)
    }
}
```

The outer layer handles rate limiting with exponential backoff. The inner layer handles token expiration. They compose independently and each can be tested in isolation.

### Command builders

Common command shapes are abstracted:

- **`BuildIDGetCommand`**: GET by ID. Takes an ID flag, calls a fetch function, prints the result.
- **`BuildListCommand`**: GET with pagination. Handles `--limit`, `--offset`, fetch-all-by-default.
- **`BuildDeleteCommand`**: DELETE with `--confirm` safety flag.
- **`BuildSmartListCommand`**: Unified list/find. Uses GET by default; auto-routes to POST find when `--filter`, `--sort`, or `--selector` is provided. Supports `FindAllExec` for cross-parent search when parent IDs are omitted. Supports `FindWhenMissingParents` to auto-switch to find when optional parent IDs are not provided.
- **`BuildCreateCommand`**: POST with `--from-json` file/stdin. Used by entities that haven't been converted to shortcut flags yet.

Each command builder takes a config struct and returns a `*ffcli.Command`. Simple commands (get, list, delete) use builders; create and update commands are implemented as custom `ffcli.Command` instances to support shortcut flags.

### List pagination helpers

List commands must load all pages by default. The shared implementation uses:

- `api.FetchAll` for typed GET-list pagination.
- `api.FetchAllRaw` for CLI handlers that need a merged raw JSON response shaped as `{"data":[...]}`.
- `shared.FetchAllSelectorPages` for POST find/selector pagination. `BuildSmartListCommand` uses this helper when `--limit` is omitted or `0`; explicit `--limit N` makes one request with that limit.

Command-specific list handlers should construct the request once, call `api.FetchAllRaw` when `limit == 0`, and only call `client.Do` directly for explicit limits. Custom selector-based list handlers should use `shared.FetchAllSelectorPages` for the same default behavior.

### Shortcut flags for create and update

All create and update commands support shortcut flags as an alternative to `--from-json`:

```
# Instead of: aads campaigns create --from-json campaign.json
aads campaigns create --name "My Campaign" --adam-id 123 --daily-budget-amount 50 --countries-or-regions US
```

Shared helpers in `internal/cli/shared/shortcuts.go`:

- **`ParseMoneyFlag(value)`**: Parses `"100"` (uses default currency) or `"100 USD"`. Returns `{"amount": "100", "currency": "USD"}`.
- **`ParseDateFlag(value)`**: Parses date-only flags as `YYYY-MM-DD`, `now`, or signed offsets like `-5d` / `+1mo`. Returns normalized `YYYY-MM-DD`.
- **`NormalizeStatus(input, activeValue)`**: Normalizes `0`/`1`/`pause`/`enable`/etc. to API status values. `activeValue` is `"ENABLED"` for campaigns/adgroups/ads, `"ACTIVE"` for keywords/negatives.
- **`ParseTextList(value)`**: Splits comma-separated text with quote support for items containing commas.

Each entity's create/update command builds a JSON body from shortcut flags, applies safety checks (budget/bid limits), and sends it via the same API request path as `--from-json`.

JSON-bearing flags use one shared convention:

- inline JSON: pass the JSON string directly
- file input: `@file.json`
- stdin input: `@-`

This applies to `--from-json` and `--selector`. A leading `@` is reserved for file/stdin input and is not treated as inline JSON.

The structure workflow extends the same convention to `--from-structure`, `--campaigns-from-json`, `--adgroups-from-json`, and `--loc-invoice-details`.

### Structure export/import orchestration

`aads structure export` and `aads structure import` are CLI-owned orchestration commands layered on top of the existing request files.

Design points:

- Export owns a versioned JSON schema with root fields `schemaVersion`, `type`, `scope`, and `creationTime`.
- Export schema uses API-style camelCase field names rather than table/output column names.
- Export supports two scopes:
  - `campaigns`: campaigns -> campaign negatives -> ad groups -> ad group negatives -> keywords
  - `adgroups`: ad groups -> ad group negatives -> keywords
- Export reuses the existing selector/filter parsing rules, but binds separate `--campaigns-*` and `--adgroups-*` selector flags.
- Export builds the full JSON structure in memory before writing stdout so failed exports never produce partial JSON that could accidentally flow into `structure import`.
- Export defaults to normalized creation-oriented entity payloads; `--*-fields ""` switches an entity type to full-fidelity export with no default-value omission.

Import is deliberately plan-first:

- Parse and validate the structure schema.
- Normalize each entity by dropping read-only and relationship fields from imported data before applying overrides.
- Apply per-entity override JSON (`--campaigns-from-json`, `--adgroups-from-json`) and shortcut flags using the precedence rule:
  - CLI override
  - transformed name/template value
  - JSON value
  - command default
- Respect skip flags during planning, including `--no-adgroups` for campaigns-scope imports so ad group creation and all ad group descendants are removed from the execution plan before validation/output.
- Resolve name templates, including capture groups from `--*-name-pattern` and `CAMPAIGN_` variables for ad group naming.
- Detect within-batch and remote name collisions before the first mutating request.
- Build the same ordered execution plan for `--check` and live import.
- For `--check`, substitute sequential mock numeric IDs and emit a mapping document instead of calling mutating endpoints.
- For live import, emit the same mapping schema with real IDs, and preserve partial mapping output on failure.

This keeps single-entity create/update commands simple while containing structure-clone complexity inside one package.

### Unified negatives commands

The `negatives` command group uses 5 unified commands (`list`, `get`, `create`, `update`, `delete`) instead of separate `campaign-*` and `adgroup-*` variants. Presence of `--adgroup-id` determines which API endpoints are called. This halves the command count while keeping full endpoint coverage.

### Orgs naming

The CLI exposes Apple Ads ACL data under the `orgs` command group because that
matches the user-facing concept better than Apple's internal ACL terminology.
Internally, the request package and API paths still use ACL naming (`/acls`,
`/me`).

### Stdin piping

ID flags accept `-` as a value to read from stdin, enabling pipes like:

```
aads campaigns list -o ids | aads adgroups list --campaign-id -
```

`CollectStdinFlags` detects `-` values and `RunWithStdin` processes each line, calling the command's exec function per ID.

### Snapshot tests

Each endpoint has a golden file containing the expected full HTTP request:

```
POST https://api.searchads.apple.com/api/v5/campaigns
Accept: application/json
Authorization: Bearer <token>
X-Ap-Context: orgId=12345
Content-Type: application/json

{
  "adamId": 789,
  "name": "My Campaign",
  "dailyBudgetAmount": {"amount": "10", "currency": "USD"},
  "countriesOrRegions": ["US"],
  "adChannelType": "SEARCH",
  "billingEvent": "TAPS",
  "supplySources": ["APPSTORE_SEARCH_RESULTS"]
}
```

Tests build a request using the Go types, then compare the resulting `http.Request` against the golden file. This catches URL path typos, missing headers, wrong HTTP methods, and serialization bugs.

Golden files live under `internal/api/testdata/request_snapshots/`.

### Reported command failures

When a user reports that a concrete command fails, handle it as an end-to-end CLI bug first, not as a helper-level hypothesis.

Recommended workflow:

1. Reproduce the exact reported command, including stdin piping, output flags, and argument order.
2. Inspect the actual request and response shape involved in that failing path.
3. Add a command-level regression test in `cmd/root_test.go` that matches the reported invocation as closely as possible.
4. Implement the smallest fix that makes that command-level regression pass.
5. Add narrower package-level unit tests only if they clarify the parser, formatter, or helper behavior behind the fix.
6. Do not report the issue fixed until the exact command path has been exercised by a regression test or direct reproduction.

Use `cmd/root_test.go` when the failure is about real CLI behavior:
- flag parsing
- stdin handling
- command routing
- output formatting
- interaction between multiple layers

Use package-level tests when the failure is isolated to one component:
- selector parsing
- local filtering
- report flattening
- safety checks
- date parsing

## API Spec Cache

### Why

Apple does not publish an OpenAPI spec for the Ads API. The spec cache serves three purposes:
1. Powers `aads schema` for runtime introspection
2. Detects API changes
3. Documents the full API surface in a machine-readable format

### How

Apple Ads has no OpenAPI spec. Instead, Apple serves documentation JSON at predictable URLs:

```
https://developer.apple.com/tutorials/data/documentation/apple_ads.json
https://developer.apple.com/tutorials/data/documentation/apple_ads/campaigns.json
https://developer.apple.com/tutorials/data/documentation/apple_ads/campaign.json
...
```

The pipeline:

1. `scripts/generate_openapi.py` — crawls Apple's doc JSON format, extracts the current API version from the latest changelog, extracts endpoint metadata, and writes `docs/apple_ads/openapi-v<version>.json`, `docs/apple_ads/openapi-latest.json`, and `docs/apple_ads/paths.txt`
2. `scripts/generate_schema_index.py` — builds compact `schema_index.json` for embedding

Generated artifacts:

- `docs/apple_ads/openapi-v<version>.json`
  - versioned lightweight OpenAPI document derived from Apple's endpoint docs
- `docs/apple_ads/openapi-latest.json`
  - symlink to the latest versioned OpenAPI document
- `docs/apple_ads/paths.txt`
  - flat `METHOD /path` index derived from the same crawl
- `internal/cli/schema/schema_index.json`
  - compact embedded schema index used by the `schema` command

The normal regeneration entrypoint for the Apple Ads docs artifacts is:

```bash
make openapi
```

To verify the committed artifacts without rewriting them:

```bash
make openapi-check
```

```makefile
make test-integration     # mocked command-integration coverage in ./cmd
make test-live            # live command-integration coverage in ./cmd
make commands-doc         # refresh docs/commands.md from CLI help
make commands-doc-check   # verify docs/commands.md matches current CLI help
make openapi              # step 1
make openapi-check        # verify step 1 artifacts without rewriting them
make schema-index         # step 2
```

### CI freshness check

`.github/workflows/api-freshness.yaml` runs weekly:
1. Re-runs `scripts/generate_openapi.py`
2. Diffs the committed generated artifacts
3. If changed, opens a PR with the diff and a summary of new/changed endpoints

## Agency Considerations

### Multi-org as a first-class feature

Named profiles in `~/.aads/config.yaml` support managing multiple client organizations. The `--profile` flag switches between them. For tests and isolated workflows, `--config-dir` or `AADS_CONFIG_DIR` can redirect both `config.yaml` and `token.json` to another directory. Prefer `AADS_CONFIG_DIR` in tests so isolated config and token cache paths can be set up without relying on root-flag placement in every helper.

### Defaults optimized for bulk operations

- List commands fetch all records by default (API max page size: 1000, loop until done)
- `--limit N` constrains the fetch
- Report commands default to `--group-by countryOrRegion` for campaign reports

### Known API quirks

These are encoded as CLI behavior, not left for users to discover:

1. **Search term reports require ORTZ timezone**. The CLI forces this for `reports searchterms` and warns if the user passes `--timezone UTC`. (From Phiture's production bugs).

2. **Granularity and totals are mutually exclusive**. When `--granularity` is set, `--row-totals` and `--grand-totals` are auto-disabled with a warning. (From Phiture's production bugs).

3. **Ad group targeting updates require full dimensions**. The `--merge` flag on `adgroups update` handles this safely. (From Phiture's production bugs).

4. **Campaign update envelope**. PUT `/campaigns/{id}` wraps the body in `{"campaign": {...}}`. Other resources don't. The CLI handles this transparently. (From Phiture's production bugs).

5. **Money amounts are strings**. The API requires `{"amount": "10.00", "currency": "USD"}`, not `{"amount": 10.00}`. The types model this correctly. (From Phiture's production bugs).

6. **Keywords bulk limit: 1000 per call**. The CLI should chunk larger keyword lists automatically. (From Phiture's production bugs).

## Build and Release

### Makefile targets

```makefile
build              # go build -o bin/aads
install            # go install
install-git-hooks  # git config core.hooksPath .githooks
test               # go test ./...
test-integration   # go test ./cmd -run '^TestIntegration_'
test-live          # AADS_INTEGRATION_TEST=1 go test ./cmd -run '^TestLive_' -count=1
lint               # golangci-lint run
vet                # go vet ./...
fmt                # gofmt -w .
commands-doc       # scripts/generate_commands_doc.py docs/commands.md
commands-doc-check # scripts/generate_commands_doc.py docs/commands.md --check
openapi            # scripts/generate_openapi.py → docs/apple_ads/openapi-v<version>.json + openapi-latest.json + paths.txt
openapi-check      # scripts/generate_openapi.py --check
schema-index      # scripts/generate_schema_index.py
release-artifacts  # python3 scripts/create_release_artifacts.py --version <x.y.z> --output-dir release
```

The repository tracks git hooks under `.githooks/`. Run `make install-git-hooks`
to configure the local clone to use them via `core.hooksPath`. The current
pre-commit hook runs `make format-check`, `make lint`, and `make test` and
fails the commit on the first failing check.

### Release Packaging

`make release-artifacts VERSION=x.y.z` builds release archives for:
- `darwin/arm64` (Apple Silicon)
- `darwin/amd64` (Intel Mac)
- `linux/arm64`
- `linux/amd64`
- `windows/arm64`
- `windows/amd64`

Artifacts are written to `release/` as:
- `.tar.gz` archives for macOS and Linux
- `.zip` archives for Windows
- a versioned SHA-256 checksum file for the published archives

The packaging tool signs macOS binaries with `codesign` when `APPLE_DEVELOPER_ID`
is set and assumes the signing certificate has already been imported into the
keychain. Official releases run on `macos-latest` and fail if signed macOS
artifacts cannot be produced.

### Install Script

`install.sh` installs the latest GitHub Release binary without requiring Go. It
resolves `https://github.com/imesart/apple-ads-cli/releases/latest`, detects the
local OS and architecture, downloads the matching `.tar.gz` archive and
versioned checksum file, verifies SHA-256, and installs `aads`.

By default it installs into `$HOME/.local/bin` when `HOME` is set, otherwise
`/usr/local/bin`. Set `INSTALL_DIR` to override the destination, `VERSION` to
install a specific release, or `REPO` to test against another GitHub repository
using the same release asset naming scheme.

### GitHub Release Workflow

`.github/workflows/release.yml` is the release entrypoint for tagged versions
matching `x.y.z`.

It:
- validates the tag format
- runs formatting, lint, and test guardrails
- imports the Apple Developer ID certificate
- runs `make release-artifacts VERSION=<tag>`
- uploads `release/*` to GitHub Releases
- updates `github.com/imesart/homebrew-tap`

### Version injection

```go
var version = "dev"           // overridden by ldflags at build time
var targetAPIVersion = "5.5"  // derived from docs/apple_ads/openapi-latest.json at build time
```

Release builds set `-ldflags "-X main.version=1.0.0 -X main.commit=<sha> -X main.date=<rfc3339> -X main.targetAPIVersion=<openapi-version>"`.

## Testing Strategy

See [`docs/requirements.md`](requirements.md#testing-requirements) for the full testing requirements (unit, snapshot, mocked-response, command-level, and integration tests).

## Non-Goals

The following are explicitly out of scope:

- Plugin or extension systems
- Raw HTTP passthrough mode
- Embedded scripting engines
- Interactive TUI workflows
- Full endpoint parity with every historical API version
- Windows distribution (unless clear user need emerges)
