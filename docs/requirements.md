# aads — Apple Ads CLI Specification

This document defines the product and behavior requirements for `aads`.
Use it as the source of truth for CLI behavior, user-facing conventions, safety rules, and testing expectations.

## Overview

`aads` is a fast, lightweight, scriptable command-line interface for the Apple Ads Campaign Management API v5. It enables advertisers and agencies to manage campaigns, ad groups, keywords, creatives, and reporting from the terminal, IDE, or CI/CD pipeline.

Designed for humans and AI agents alike.

### Target Audience

- Individual app marketers managing their own campaigns
- Growth teams running campaigns across multiple apps
- Agencies managing multiple client organizations
- Automation and CI/CD workflows

The CLI must work well for both interactive terminal use and machine-readable scripting.

## Target API

- **API**: Apple Ads Campaign Management API
- **Version**: 5 (current: 5.5, February 2026)
- **Base URL**: `https://api.searchads.apple.com/api/v5/`
- **Reference**: https://developer.apple.com/documentation/apple_ads

## Authentication

Apple Ads uses OAuth 2.0 with a client credentials grant and ES256 JWT assertion.

### Flow

1. Build a JWT assertion signed with ES256 (P-256):
   - `iss` = team_id
   - `sub` = client_id
   - `aud` = `https://appleid.apple.com`
   - `iat` = now
   - `exp` = now + 86400 (configurable, max 180 days)
   - Header `kid` = key_id
   - Header `alg` = ES256
2. POST to `https://appleid.apple.com/auth/oauth2/token`:
   - `grant_type=client_credentials`
   - `client_id=<client_id>`
   - `client_secret=<jwt>`
   - `scope=searchadsorg`
3. Receive `access_token` (typically valid ~1 hour).
4. On every API request:
   - `Authorization: Bearer <access_token>`
   - `X-AP-Context: orgId=<org_id>`

### Credentials

| Field | Description | Config key | Env var |
|---|---|---|---|
| Client ID | Received when uploading a public key | `client_id` | `AADS_CLIENT_ID` |
| Team ID | Received when uploading a public key | `team_id` | `AADS_TEAM_ID` |
| Key ID | Received when uploading a public key | `key_id` | `AADS_KEY_ID` |
| Private key path | Path to PEM file (ES256 / P-256) | `private_key_path` | `AADS_PRIVATE_KEY_PATH` |
| Org ID | Organization identifier | `org_id` | `AADS_ORG_ID` |

### Token Caching

Access tokens are cached at `~/.aads/token.json` with expiry tracking. The token is refreshed automatically when expired or on 401 response.

## Configuration

### Config file

Default location: `~/.aads/config.yaml`

Override the configuration directory with `--config-dir <dir>` or `AADS_CONFIG_DIR`.
When set, the CLI reads and writes `config.yaml` and `token.json` under that directory.

### Named profiles

```yaml
default_profile: "default"
profiles:
  default:
    client_id: "SEARCHADS.abc123-..."
    team_id: "SEARCHADS.abc123-..."
    key_id: "abc123-..."
    org_id: "1234567"
    private_key_path: "~/.aads/private-key.pem"
    default_currency: "USD"
    default_timezone: "Europe/Luxembourg"
    default_time_of_day: "09:00"
    max_daily_budget: "1000.00"
    max_bid: "10.00"
    max_cpa_goal: ""
    max_budget: "50000.00"

  client-acme:
    client_id: "SEARCHADS.def456-..."
    team_id: "SEARCHADS.def456-..."
    key_id: "def456-..."
    org_id: "7654321"
    private_key_path: "~/.aads/acme-key.pem"
    default_currency: "EUR"
    default_timezone: ""
    default_time_of_day: ""
    max_daily_budget: ""
    max_bid: ""
    max_cpa_goal: ""
    max_budget: ""
```

Safety limits in config are stored as decimal strings in `default_currency`.
Use `""` to disable a limit.

Usage: `aads --profile client-acme campaigns list`

Profile creation requires `org_id`, but `aads profiles create` may infer it when
enough Apple Ads credentials are provided. If `--org-id` is omitted, the CLI
calls `orgs user`, uses `parentOrgId` as the profile `org_id`, then calls
`orgs list` to find the matching organization row and infer
`default_currency` and `default_timezone`. Apple refers to this data as ACLs;
the CLI exposes it under the `orgs` command group. Explicit CLI flags take
precedence over inferred ACL values. If the matching ACL row is missing, the
CLI warns and still creates the profile with the resolved `org_id`.

`aads profiles create --interactive` is a TTY-only setup wizard. It prompts for
missing fields, can guide the user through API-user invitation and key setup,
parses pasted `clientId` / `teamId` / `keyId` text, and gathers safety limits
before writing the profile. The interactive flow creates the profile only after
all prompts complete successfully. Browser launch is best-effort: known browsers
are opened in private mode when possible, otherwise the default browser is used.

### Precedence

1. CLI flags (`--org-id`, `--client-id`, etc.)
2. Environment variables (`AADS_ORG_ID`, etc.)
3. Named profile (via `--profile`)
4. Default profile in config file

### Environment variables

| Variable | Description |
|---|---|
| `AADS_CLIENT_ID` | Client ID |
| `AADS_TEAM_ID` | Team ID |
| `AADS_KEY_ID` | Key ID |
| `AADS_PRIVATE_KEY_PATH` | Private key path |
| `AADS_ORG_ID` | Organization ID |
| `AADS_PROFILE` | Override profile selection |
| `AADS_CONFIG_DIR` | Override the configuration directory for `config.yaml` and `token.json` |
| `AADS_DEFAULT_OUTPUT` | Default output format |
| `AADS_TIMEOUT` | Request timeout in seconds |
| `AADS_DEBUG` | Enable debug output |
| `AADS_API_DEBUG` | Enable API request/response debug output |

Environment variables override stored config values.

### Profile setup

`aads profiles genkey --name <name>` generates a P-256 private key at `~/.aads/keys/<name>-private-key.pem` and prints the corresponding public key to stdout.
`aads profiles create --name <name>` creates a named profile in `~/.aads/config.yaml` from credentials passed via flags. When `--private-key-path` is omitted, it stores the default key path `~/.aads/keys/<name>-private-key.pem`. If `--private-key-path` is provided explicitly, the file must already exist.
`aads profiles create --interactive` prompts for missing fields in a terminal wizard. It defaults the profile name to `default`, can generate the default key file before profile creation, prompts for `max_daily_budget`, `max_bid`, and `max_cpa_goal` with `0` meaning "no limit", and writes the profile only after the final confirmation-free create step succeeds.
`aads profiles update --name <name>` updates an existing profile.
`aads profiles delete --name <name> --confirm --delete-private-key` deletes the profile and then attempts to delete the configured private key file, warning if the file is missing or cannot be removed.
`aads profiles get --name <name> --show-key` prints the configured profile public key instead of profile details.

## Command Structure

### Global flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--verbose` | `-v` | false | Verbose output (show HTTP requests/responses) |
| `--profile` | `-p` | `default` | Config profile name |
| `--config-dir` | | | Configuration directory override for `config.yaml` and `token.json` |
| `--org-id` | | | Override org ID from config |
| `--fields` | | | Comma-separated fields for partial fetch |
| `--currency` | | | Override currency for money fields |
| `--no-color` | | false | Disable color output |

### Output behavior

- When stdout is a TTY: default to `table`
- When stdout is piped: default to `json`
- Override with `--format` / `-f`
- `pipe` renders tab-separated records with a header row for downstream `aads` stdin pipelines
- Commands that accept `--format` / `-f` also accept `--pretty` to pretty-print JSON even when stdout is not a TTY

### Command groups

```
aads campaigns list        List campaigns
aads campaigns get         Get a campaign by ID
aads campaigns create      Create a campaign
aads campaigns update      Update a campaign
aads campaigns delete      Delete a campaign

aads adgroups list         List ad groups
aads adgroups get          Get an ad group by ID
aads adgroups create       Create an ad group
aads adgroups update       Update an ad group
aads adgroups delete       Delete an ad group

aads keywords list         List targeting keywords
aads keywords get          Get a targeting keyword
aads keywords create       Create targeting keywords
aads keywords update       Update targeting keywords
aads keywords delete       Delete targeting keywords

aads negatives list        List negative keywords (campaign or ad group level)
aads negatives get         Get a negative keyword by ID
aads negatives create      Create negative keywords
aads negatives update      Update negative keywords
aads negatives delete      Delete negative keywords

aads ads list              List ads
aads ads get               Get an ad by ID
aads ads create            Create an ad
aads ads update            Update an ad
aads ads delete            Delete an ad

aads creatives list        List all creatives
aads creatives get         Get a creative by ID
aads creatives create      Create a creative

aads budgetorders list     List all budget orders
aads budgetorders get      Get a budget order
aads budgetorders create   Create a budget order
aads budgetorders update   Update a budget order

aads product-pages list         List custom product pages
aads product-pages get          Get a product page
aads product-pages locales      Get product page locales
aads product-pages countries    Get supported countries/regions
aads product-pages devices      Get app preview device sizes

aads ad-rejections list         List ad creative rejection reasons
aads ad-rejections get          Get rejection reasons by ID
aads ad-rejections assets       List app assets

aads reports campaigns     Campaign-level reports
aads reports adgroups      Ad group-level reports
aads reports keywords      Keyword-level reports
aads reports searchterms   Search term-level reports
aads reports ads           Ad-level reports

aads impression-share create   Create an impression share report
aads impression-share get      Get a single impression share report
aads impression-share list     List all impression share reports

aads apps search           Search for iOS apps
aads apps eligibility      Check app eligibility
aads apps details          Get app details
aads apps localized        Get localized app details

`aads apps eligibility` uses a Selector request body for `--from-json`.
Shortcut flags build selector conditions automatically.
As an alternate input form, `--from-json` also accepts `adamId`, `countryOrRegion`,
`deviceClass`, `supplySource`, and `minAge`, which the CLI translates into selector conditions.
`aads apps search` requires at least one of `--query` or `--only-owned-apps`.

Endpoint-specific query flags:

- `apps search --only-owned-apps` sends `returnOwnedApps=true`.
- `creatives get --include-deleted-creative-set-assets` sends `includeDeletedCreativeSetAssets=true`.
- `product-pages list --filter "name=VALUE"` and `--filter "state=VALUE"` send the corresponding query parameters. Other filter fields are rejected.
- `product-pages countries --countries-or-regions US,GB` sends `countriesOrRegions=US,GB`.
- `impression-share list --sort "field:asc|desc"` sends `field=<field>` and `sortOrder=ASCENDING|DESCENDING`.

aads geo search            Search for geolocations
aads geo get               Get geolocation details

`aads geo get` requires `--entity` and `--geo-id` to fetch details for a
specific geographic identifier.

`aads reports adgroups` accepts `--adgroup-id` as a selector shortcut.
It adds `adGroupId EQUALS <value>` to the report selector body and supports `-`
for stdin-driven campaign/ad group pipelines.

aads orgs list             List organizations
aads orgs user             Get current user details

aads structure export      Export a structure JSON for campaigns/ad groups, keywords, and negatives
aads structure import      Import campaigns/ad groups, keywords, and negatives from structure JSON

aads profiles list         List all configured profiles
aads profiles get          Show profile details
aads profiles genkey       Generate a P-256 private key and print its public key
aads profiles create       Create a new profile
aads profiles update       Update an existing profile
aads profiles delete       Delete a profile
aads profiles set-default  Set the default profile

aads version               Print the CLI version and target Apple Ads API version (5.5)
aads schema                Query embedded API schema information
aads completion <shell>    Generate shell completion (bash, zsh, fish)
```

**Total: 69 commands across 19 command groups.**

For per-endpoint implementation status and source file mapping, see [`docs/coverage.md`](coverage.md).

### Help text

- Command help should describe user-visible behavior and inputs, not HTTP method details.
- Help output should avoid request-method wording such as `GET`, `POST`, `PUT`, and `DELETE` unless the command is explicitly about schema or API method inspection.
- For selector-based `list` commands, document `Sortable fields:` separately only when a distinct sortable subset is known.
- If sortable and filterable fields are the same, or the source data does not distinguish them, prefer `Searchable and filterable fields:`.

Note: The `negatives` commands use a unified interface — `--adgroup-id` presence
determines whether operations target campaign-level or ad group-level negative keywords.
Without `--adgroup-id`, commands operate at the campaign level.

## Input Rules

### Resource ID flags

Resource identifiers must always use explicit flag names:

- `--campaign-id`
- `--adgroup-id`
- `--keyword-id`
- `--ad-id`
- `--budget-order-id`
- `--report-id`
- `--geo-id`

Never bare `--id`. This rule applies consistently across get, update, delete, and any child-resource command that targets a specific resource.

### Collection querying

There are no separate `find` subcommands. The `list` command handles both plain listing and filtered queries:

- `list` for plain listing
- `list --filter ...` for natural filter queries
- `list --sort ...` for sorted results
- `list --selector ...` for raw selector JSON

When `--filter`, `--sort`, or `--selector` is provided, `list` automatically switches to filtered query mode. For resources where parent IDs are optional (ad groups, ads), omitting the parent ID routes to a cross-parent query.

Rules:

- `--filter` and `--selector` are mutually exclusive
- `--sort` is valid on its own or combined with `--filter`
- `--sort` is not allowed with `--selector`
- Repeated `--selector` flags are invalid
- Repeated `--filter` flags produce multiple conditions (AND)
- Repeated `--sort` flags are preserved in order

Selector input forms:

- Inline JSON: `--selector '{"conditions":[...]}'`
- From file: `--selector @file.json`
- From stdin: `--selector @-`

Natural query syntax:

- `--filter "status=ENABLED"` or `--filter "status = ENABLED"` — EQUALS
- `--filter "name!=Test"` or `--filter "name != Test"` — local NOT_EQUALS
- `--filter "name STARTSWITH MyApp"` — named operator
- `--filter "spend [0, 5]"` — BETWEEN (2 values)
- `--filter "status [ENABLED, PAUSED]"` — IN (3+ values)
- `--sort "name:asc"` — sort by field

Operator names are case-insensitive.

`!=` is a local-only filter operator. For list commands, other filters are still sent to the API first, then `!=` conditions are applied to the returned rows locally. On reports, all `--filter` conditions are already local. Missing field values make `field != x` evaluate to true, except when comparing to `''`, `""`, or `null`, in which case it evaluates to false. Unquoted `null` is treated as a null literal; quoted `'null'` and `"null"` are treated as text.

Field aliases and carried stdin fields:

- API-shaped fields remain the default output contract (`id`, `name`, etc.).
- `--fields` preserves the requested field names in output.
- Alias fields are accepted only in `--fields`, `--filter`, and `--sort`.
- On campaign-outputting commands, `campaignId` aliases `id` and `campaignName` aliases `name`.
- On ad-group-outputting commands, `adGroupId` aliases `id` and `adGroupName` aliases `name`.
- On keyword-outputting commands, `keywordId` aliases `id`.
- On creative-outputting commands, `creativeId` aliases `id`.
- On product-page-outputting commands, `productPageId` aliases `id`.
- Compact ad-group spellings like `ADGROUP_ID`, `adgroupId`, and `ADGROUP_NAME` are accepted as input aliases in `--fields`, `--filter`, and `--sort`, but canonical output uses `AD_GROUP_*` / `adGroup*`.
- Compact creative/product-page spellings like `CREATIVE_ID` and `PRODUCT_PAGE_ID` are accepted as input aliases in `--fields`, `--filter`, and `--sort`.
- Carried stdin fields are available only when stdin rows provide them and each hop includes them in `--fields`.
- Supported carried stdin fields are `adamId`, `appName`, `appNameShort`, `creativeId`, `campaignName`, `budgetAmount`, `dailyBudgetAmount`, `productPageId`, `adGroupName`, `defaultBidAmount`, and `cpaGoal`.
- Real downstream fields win over carried stdin fields with the same name.
- Raw `--selector` input does not support aliases or carried-field resolution. For selector-based list commands, the CLI may override `selector.pagination` to implement default fetch-all behavior or an explicit `--limit`.

Synthetic fields in filters:

- Synthetic carried fields may be used as the value side of `--filter` expressions.
- Resolution happens per executed request, not once for the whole CLI invocation.
- Missing synthetic values are validation errors.
- For amount-vs-amount comparisons without `.amount`, currencies must match or validation fails.
- For amount-vs-scalar comparisons like `localSpend > 10`, currency is ignored and only the numeric amount is compared.

Synthetic fields in sorting:

- Alias fields in `--sort` are translated to their API field names and may be sorted remotely.
- Carried stdin fields in `--sort` are sorted locally because the API cannot see them.
- When any sort key requires local sorting, the CLI sends any remote-sortable keys to the API, fetches and merges all rows, then applies the full `--sort` list locally in flag order.
- Local synthetic sorting is invalid with `--limit > 0`; omit `--limit` or use `--limit 0` so sorting can be applied to all rows.
- Missing numeric fields sort as `0`; missing string fields sort as `""`.
- Sorting money fields by amount rejects mixed currencies. Sorting an explicit `.currency` field is a string sort.
- Raw `--selector` input does not support aliases or carried-field sorting.

### JSON payloads

Mutation commands accept:

- Shortcut flags for common fields
- `--from-json @file.json` for file input
- `--from-json @-` for stdin
- `--from-json '{"key":"value"}'` for inline JSON

A leading `@` is reserved for file/stdin input.

ID flags that accept `-` for stdin support both:

- tab-separated `-f ids` output, with or without a header row
- JSON records containing the needed `campaignId`, `adGroupId`, and/or `id` fields

When a JSON record uses a generic `id` field, it is resolved to the related resource ID (`CAMPAIGN_ID`, `AD_GROUP_ID`, `KEYWORD_ID`, and so on) using the same hierarchy rules as `-f ids`.

Endpoint-level coverage and documented-vs-implemented API drift are tracked in [`docs/coverage.md`](coverage.md).

## Pagination

### GET endpoints (list)

Use query parameters `?limit=N&offset=M`. Default page size: 1000 (API maximum).

All list commands load all records across all pages by default (agency convention). Use `--limit N` to constrain results.

### POST endpoints (find)

Use request body: `{ "pagination": { "offset": 0, "limit": 1000 } }`.
Selector-based `list` commands also load all matching records across all pages by default. Use `--limit N` to constrain results.

### POST endpoints (reports)

Use `selector.pagination` within the reporting request body.

## Reporting

### Common report flags

| Flag | Default | Description |
|---|---|---|
| `--start` | (required) | Start date: YYYY-MM-DD, `now`, or signed offset like `-5d` |
| `--end` | (required) | End date: YYYY-MM-DD, `now`, or signed offset like `-5d` |
| `--granularity` | | HOURLY, DAILY, WEEKLY, MONTHLY |
| `--group-by` | | countryOrRegion, deviceClass, etc. |
| `--filter` | | Local post-fetch filter: `field=value` or `field OPERATOR value` (repeatable) |
| `--timezone` | UTC | Timezone (ORTZ for search terms) |
| `--row-totals` | true | Include row totals |
| `--grand-totals` | true | Include grand totals |
| `--no-metrics` | false by default; true for `reports campaigns` and `reports adgroups` | Include records with no metrics |
| `--condition` | | API request condition: `field=operator=value` |
| `--sort` | `impressions:desc` | Sort field:order (commands with impressions default to `impressions:desc`) |

### API constraints

- Search term reports (`reports searchterms`) force `--timezone ORTZ`. The CLI enforces this automatically and warns if the user passes a different timezone.
- Report day flags only support report timezone semantics for `UTC` and `ORTZ`.
- Profile `default_timezone` does not affect report day flags unless it is `UTC`.
- When `--granularity` is set, `--row-totals` and `--grand-totals` are automatically disabled. The API rejects requests where granularity and totals are both present.
- Campaign reports default to `--group-by countryOrRegion` when no group-by is specified.
- All current `reports` commands default `--sort` to `impressions:desc`.
- All `reports` commands support repeatable local `--filter` expressions, applied after report flattening and combined with AND.
- `reports campaigns` and `reports adgroups` default `--no-metrics` to true so local filters like `IMPRESSIONS < 10` include zero-impression rows.

### Impression share reports

These use a different request schema from standard reports:

| Flag | Description |
|---|---|
| `--name` | Report name (required) |
| `--start` | Start date |
| `--end` | End date |
| `--granularity` | DAILY or WEEKLY |
| `--date-range` | LAST_WEEK, LAST_2_WEEKS, LAST_4_WEEKS (WEEKLY only) |
| `--condition` | Filter conditions |

Date-only flags (`--start`, `--end`, `--start-date`, `--end-date`) also accept relative expressions:

- `now` means local today, formatted as `YYYY-MM-DD`
- signed offsets like `-5d`, `+1mo`, `-2weeks`
- supported units: canonical `d`, `w`, `mo`, `y`, plus aliases `day/days`, `week/weeks`, `month/months`, `year/years`
- plain `YYYY-MM-DD` remains valid
- relative expressions are either `now` or `<signed-offset>`, where `<signed-offset>` is `[+-]<int><unit>`
- no whitespace is allowed inside a relative expression
- for report day flags, relative expressions use `UTC` day boundaries when the effective report timezone is `UTC`
- otherwise report day flags do not use profile `default_timezone` unless it is `UTC`

Time flags (`--start-time`, `--end-time`) accept:

- full datetimes in the existing ISO/RFC3339 forms
- `YYYY-MM-DD`
- `now`
- signed offsets like `+5d`

The CLI resolves these shortcut forms and sends the API datetime format.
Help text for time flags should describe them as UTC API times and should list
the accepted shortcut forms in the flag help itself rather than duplicating that
guidance in the command description.

## Safety Limits

Configurable spend safeguards to prevent accidental overspend:

```yaml
# In config.yaml
max_daily_budget: "1000.00"
max_bid: "10.00"
```

- `campaigns create` / `campaigns update`: validates `dailyBudgetAmount` and `budgetAmount` against `max_daily_budget` (via `CheckBudgetLimitJSON`)
- `keywords create` / `keywords update`: validates `bidAmount` against `max_bid` (via `CheckBidLimitJSON`)
- `adgroups create` / `adgroups update`: validates `defaultBidAmount` against `max_bid` (via `CheckBidLimitJSON`)
- Profile limits are always interpreted in `default_currency`
- In YAML, `""` disables a limit; unquoted numeric values are still accepted on read
- Override with `--force`
- If no limits are configured, no validation is performed

## Shortcut Flags for Create and Update Commands

All create and update commands support shortcut flags as an alternative to `--from-json` (JSON file input). This enables common operations without writing JSON.

### Create command shortcut flags

| Command | Required flags | Optional flags |
|---|---|---|
| `campaigns create` | `--name`, `--adam-id`, `--daily-budget-amount`, `--countries-or-regions` | `--budget-amount` (deprecated), `--loc-invoice-details`, `--ad-channel-type` (SEARCH), `--supply-sources` (APPSTORE_SEARCH_RESULTS), `--billing-event` (TAPS), `--status`, `--start-time`, `--end-time` |
| `adgroups create` | `--campaign-id`, `--name`, `--default-bid` | `--status`, `--start-time`, `--end-time`, `--automated-keywords-opt-in`, `--cpa-goal` (Search only), `--age`, `--gender`, `--device-class`, `--country-code`, `--admin-area`, `--locality` |
| `keywords create` | `--campaign-id`, `--adgroup-id`, `--text` | `--match-type` (EXACT), `--bid`, `--status` |
| `negatives create` | `--campaign-id`, `--text` | `--adgroup-id`, `--match-type` (EXACT), `--status` |
| `ads create` | `--campaign-id`, `--adgroup-id`, `--name`, `--creative-id` | `--status`, `--metadata-json` |
| `creatives create` | `--adam-id`, `--name`, `--type` | `--product-page-id` |
| `budgetorders create` | `--name` | `--start-date`, `--end-date`, `--budget-amount`, `--order-number`, `--primary-buyer-email`, `--primary-buyer-name`, `--billing-email`, `--client-name` |
| `impression-share create` | `--name` | `--granularity` (default `DAILY`), `--dateRange`, `--startTime`, `--endTime` (`--dateRange` is invalid with `DAILY`; when no date flags are provided, shortcut mode defaults to `startTime=-7d` and `endTime=now`) |

### Update command shortcut flags

| Command | ID flags | Shortcut flags |
|---|---|---|
| `campaigns update` | `--campaign-id` | `--status`, `--name`, `--budget-amount` (deprecated), `--daily-budget-amount`, `--loc-invoice-details`, `--countries-or-regions` |
| `adgroups update` | `--campaign-id`, `--adgroup-id` | `--default-bid`, `--update-inherited-bids` (requires `--default-bid`), `--cpa-goal` (Search only), `--status`, `--name`, `--start-time`, `--end-time`, `--merge` |
| `keywords update` | `--campaign-id`, `--adgroup-id`, `--keyword-id` | `--bid`, `--status` |
| `negatives update` | `--campaign-id`, `--keyword-id` | `--adgroup-id`, `--status` |
| `ads update` | `--campaign-id`, `--adgroup-id`, `--ad-id` | `--status` |
| `budgetorders update` | `--budget-order-id` | `--name`, `--start-date`, `--end-date`, `--budget-amount` |

### Conventions

- **Status normalization**: All status flags accept `0`/`1`, `pause`/`enable`, `paused`/`enabled`, or the API values (`PAUSED`/`ENABLED`/`ACTIVE`). Campaigns, ad groups, and ads use `ENABLED`/`PAUSED`; keywords and negatives use `ACTIVE`/`PAUSED`.
- **Money amounts**: Accept `"100"` (uses default currency from `--currency` flag or config `default_currency`) or `"100 USD"` (explicit currency). Errors if no currency is available.
- **Create-name templates**: `campaigns create --name` and `adgroups create --name` accept `%(fieldName)` and `%(FIELD_NAME)` variables resolved from the other fields being created. Array values render as comma-separated text.
- `adgroups create` defaults `startTime` to `now` when neither shortcut flags nor `--from-json` provide one.
- **Comma-separated text**: `--text` flag for keywords/negatives accepts comma-separated values with quote support for items containing commas: `--text '"hello, world",other'`.
- **Delete by keyword ID**: `keywords delete` and `negatives delete` accept `--keyword-id` with one ID, a comma-separated list of IDs, or `-` to read IDs from stdin. `--keyword-id` and `--from-json` are mutually exclusive. For targeting keywords, one `--keyword-id` uses the single-keyword delete endpoint; multiple IDs use the JSON-array delete endpoint. Negative keyword deletes always use the JSON-array delete endpoint.
- **`--from-json` / `--selector` JSON input**: JSON-bearing flags accept inline JSON directly, `@file.json` to read from a file, or `@-` to read from stdin. A leading `@` is reserved for file/stdin input.
- **Targeting flags** (adgroups create): `--age 18-65`, `--gender M,F`, `--device-class IPHONE,IPAD`, `--country-code US,GB`, `--admin-area "US|CA,US|NY"`, `--locality "US|CA|Los Angeles"`.
- **Budget order envelope**: Shortcut flags automatically wrap in `{"orgIds": [...], "bo": {...}}` using `--org-id` or config `org_id`.
- **Campaign update envelope**: Automatically wraps in `{"campaign": {...}}`.
- **CPA goal validation**: `--cpa-goal` fetches the campaign to verify `adChannelType` is `SEARCH` before proceeding.

## Structure Export and Import

`aads structure export` and `aads structure import` provide a higher-level workflow for copying campaign structures that include campaigns, ad groups, keywords, and negative keywords.

### Export

`aads structure export --scope campaigns|adgroups` always writes JSON to stdout.

- `--scope campaigns` exports campaigns, campaign negatives, ad groups, ad group negatives, and keywords.
- `--scope adgroups` exports ad groups, ad group negatives, and keywords from the selected campaigns.
- `--campaign-id` directly selects one campaign to export and supports `-` for stdin-driven campaign ID pipelines.
- `--shareable` applies a shareable-export preset: omit keywords, negatives, `adamId`, budget/bid/CPA/invoice fields, `startTime`, and `endTime`, and redact names.
- `--no-adam-id` omits `campaign.adamId` from the exported structure JSON.
- `--no-budgets` omits campaign budget/invoice fields, ad group bid/CPA fields, and keyword bid fields unless explicitly requested by `--campaigns-fields`, `--adgroups-fields`, or `--keywords-fields`.
- `--no-times` omits campaign/ad group `startTime` and `endTime` from exported structure JSON unless the relevant fields are explicitly requested by `--campaigns-fields` or `--adgroups-fields`.
- `--redact-names` rewrites exported campaign/ad group names to safer placeholders when high-confidence matches are found.
- `--no-negatives` skips campaign and ad group negative keyword export.
- `--no-keywords` skips keyword export.
- Campaign selection uses `--campaigns-filter` / `--campaigns-sort` / `--campaigns-selector`.
- Ad group selection uses `--adgroups-filter` / `--adgroups-sort` / `--adgroups-selector`.
- Keyword selection uses `--keywords-filter` / `--keywords-sort` / `--keywords-selector`.
- `--campaigns-filter` and `--campaigns-selector` are mutually exclusive. The same applies to the ad group selector flags.
- `--campaign-id` is mutually exclusive with `--campaigns-filter`, `--campaigns-sort`, and `--campaigns-selector`.
- If no campaign selector/filter flags are provided, all campaigns are included.
- If no ad group selector/filter flags are provided, all ad groups in the selected campaigns are included.
- If no keyword selector/filter flags are provided, all keywords in the selected ad groups are included.
- Exported JSON root fields are `schemaVersion`, `type`, `scope`, and `creationTime`.
- Exported structure JSON uses `type: "structure"` and API-style camelCase field names.
- A formal JSON Schema for structure-export output is published at `docs/schemas/aads-v1.schema.json`.
- Exported structure JSON includes a top-level `"$schema"` field pointing to the versioned schema URL.
- `--pretty` pretty-prints the JSON export even when stdout is not a TTY.
- By default, each entity exports a normalized creation-oriented field set and omits empty/default fields.
- With `--redact-names`, campaign-name redaction may replace bounded matches of the fetched app name with `%(appName)`, bounded or conservative fuzzy prefix matches of the fetched app short name with `%(appNameShort)`, and exact rendered country-list matches with `%(countriesOrRegions)`. The derived `appNameShort` uses the same separator rules as export-side name redaction.
- With `--redact-names`, ad group names may replace bounded exact matches of the original campaign name with `%(campaignName)` and apply the same app/app-short-name redaction rules as campaigns, but do not replace countries.
- `--redact-names` fetches app details from campaign `adamId` and fails the export if the lookup fails.
- `--no-budgets` may omit `dailyBudgetAmount`, `budgetAmount`, `locInvoiceDetails`, and `budgetOrders` on campaigns, `defaultBidAmount` and `cpaGoal` on ad groups, and `bidAmount` on keywords unless explicitly re-requested.
- Normalized campaign export preserves non-default `adChannelType`, `billingEvent`, `supplySources`, and non-empty `biddingStrategy`.
- Normalized ad group export preserves non-default/non-empty `pricingModel`, `automatedKeywordsOptIn`, `cpaGoal`, `paymentModel`, and `biddingStrategy`.
- `--campaigns-fields`, `--adgroups-fields`, `--keywords-fields`, `--campaigns-negatives-fields`, and `--adgroups-negatives-fields` change what is exported.
- Omitting a `--*-fields` flag uses normalized defaults.
- Passing `--*-fields ""` exports all fields for that entity type and disables default-value omission for that entity type.
- Passing `--*-fields "fieldA,fieldB"` exports those fields plus the entity’s required creation fields.
- Explicit `--*-fields` selections override shareable/no-export omissions for the matching entity type. For example, `--keywords-fields ""` re-includes keywords and `--campaigns-fields adamId,startTime` re-includes those campaign fields.
- Campaign and ad group export omit `startTime` by default when the value is already in the past, unless the relevant field flag explicitly requests `startTime` or full export.
- Ad group export normalizes `targetingDimensions` by removing null child fields, omitting default `deviceClass={"included":["IPHONE","IPAD"]}`, and omitting `targetingDimensions` entirely if nothing non-default remains, unless ad group field flags explicitly request full export.
- Keyword export omits `bidAmount` by default when it matches the parent ad group `defaultBidAmount`, unless `--keywords-fields ""` or `--keywords-fields` explicitly requests `bidAmount`.
- On export failure, no partial structure JSON may be written to stdout.

### Import

`aads structure import --from-structure JSON` accepts only the structure schema produced for this workflow.

- `--from-structure` accepts inline JSON, `@file.json`, or `@-` for stdin.
- `schemaVersion` must currently be `1`.
- `type` must be `"structure"`.
- `scope` must be `"campaigns"` or `"adgroups"`.
- `scope: "campaigns"` creates campaigns, then campaign negatives, then ad groups, then ad group negatives, then keywords.
- `scope: "adgroups"` creates ad groups, then ad group negatives, then keywords, inside an existing destination campaign.
- `--campaign-id` is required for `scope: "adgroups"` and invalid for `scope: "campaigns"`.
- `--no-adgroups` is valid only for `scope: "campaigns"` and skips ad group creation entirely.
- `--no-negatives` skips campaign and ad group negative keyword creation.
- `--no-keywords` skips keyword creation.
- `--no-adgroups` also skips all ad group negatives and keywords because no ad groups are created.
- With `--no-adgroups`, campaign negatives are still created unless `--no-negatives` is also set.
- Import precedence is: CLI override > transformed name/template value > JSON value > command default.
- When normalized export omitted create-time defaults, import must restore those defaults before constructing campaign/ad group create requests.
- `--campaigns-from-json` and `--adgroups-from-json` apply the same override object to every created campaign/ad group of that type.
- Override JSON flags accept inline JSON, `@file.json`, or `@-` for stdin.
- `--budget-amount` is a deprecated campaign override flag for structure imports.
- Override JSON must not include read-only or relationship fields; those are rejected as invalid input.
- Imported structure data may include read-only and relationship fields; import ignores those fields before request construction.
- Relationship fields are always re-bound during import. For example, imported `campaignId` / `adGroupId` values do not control the destination hierarchy.
- Keyword import omits or clears `bidAmount` when it matches the resolved destination ad group `defaultBidAmount`.
- `--check` performs the same planning and validation path as a live import, including collision checks, but sends no mutating API requests.
- `--check` uses sequential mock numeric IDs in the mapping output, matching the JSON type used by the real entities handled by this workflow.
- `--check` must detect both within-batch name collisions and existing destination collisions before failing.
- Existing collision checks are case-insensitive exact-name matches.
- Name-template collisions are validation failures.

### Import naming

- `--campaigns-name` and `--adgroups-name` override the destination names using templates.
- Template expansion applies only to names.
- Imported `campaign.name` and `adgroup.name` values from the structure JSON are also treated as templates during import.
- `--campaigns-name-pattern` and `--adgroups-name-pattern` are basic regular expressions used to capture parts of the source name for `%1`, `%2`, etc.
- If a name pattern does not match, import fails unless `--allow-unmatched-name-pattern` is set.
- Document API-style variables first, but accept both API-style and CLI/output-style aliases in templates.
- Show `%(VAR)` in help/examples; `%VAR` is also accepted.
- Variables resolve from the item being created.
- Whenever import resolves app-derived variables from `adamId`, both `%(appName)` and alias `%(appNameShort)` are available.
- For `--adgroups-name`, destination campaign variables may also be referenced with the `CAMPAIGN_` prefix, for example `%(CAMPAIGN_name)`.
- `--adgroups-name` also exposes synthetic `%(campaignName)` plus app variables derived from the destination/source campaign, including `%(appName)` and `%(appNameShort)`.
- Arrays rendered into templates use comma-separated values with no brackets. For example, `countriesOrRegions=["DE","FR"]` becomes `DE,FR`.
- Imported `campaign.startTime`, `campaign.endTime`, `adgroup.startTime`, and `adgroup.endTime` accept the same datetime forms as the override flags, including absolute timestamps, `now`, and signed relative expressions like `+3mo`; they are resolved during import planning before request construction.
- When structure-imported ad groups omit `startTime`, import restores it to `now` before constructing the create request.
- If a required campaign or ad group field is missing from the structure JSON and that field can be supplied by an import override flag, the validation error must name the corresponding flag, for example `adamId` -> `--adam-id` and `defaultBidAmount` -> `--default-bid`.

### Import output

- Import emits JSON with the same hierarchy shape as the source structure and root `type: "mapping"`.
- Mapping root fields are `schemaVersion`, `type`, `scope`, and `creationTime`.
- Mapping `campaign` / `adgroup` entries contain `source` and `created` objects with IDs and names when available.
- Mapping keyword/negative entries contain `source` and `created` objects with IDs and text when available.
- Live import output adds per-item `status` with values `created`, `failed`, or `not_attempted`.
- Live import output adds per-item `error` when `status` is `failed`.
- Live import omits `created` for `not_attempted` items; failed items only include `created` when the API already returned created identity such as IDs.
- Without `--output-mapping`, mapping JSON is written to stdout.
- `--output-mapping FILE` writes mapping JSON only to `FILE`; stdout does not emit the mapping.
- `--output-mapping -` is an explicit stdout alias.
- `--pretty` pretty-prints mapping JSON when it is written to stdout, even when stdout is not a TTY.
- If a live import fails after partial creation, the partial mapping must still be emitted to the selected mapping destination before the command exits non-zero.

## Ad Group Update Behavior

When updating ad group targeting dimensions, the Apple Ads API requires the complete `targetingDimensions` object — omitting a dimension removes it. The CLI provides two modes:

- `aads adgroups update --campaign-id 123 --adgroup-id 456 --from-json dimensions.json` — sends the file contents as the full targeting object
- `aads adgroups update --campaign-id 123 --adgroup-id 456 --merge --from-json overlay.json` — fetches the current ad group, overlays the changes, PUTs the merged result
- `aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid 2.50` — shortcut for common field changes without JSON
- `aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid +10% --update-inherited-bids` — also updates keyword-level bids that currently match the ad group's previous default bid

The `--merge` flag makes partial updates safe by default.

## Campaign Update Envelope

Campaign PUT requests require wrapping the payload in `{"campaign": {...}}`. Other resources do not. The CLI handles this transparently.

## Retry and Rate Limiting

### Auth retry

On 401 (Unauthorized), the CLI:
1. Invalidates the cached access token
2. Requests a new token via the OAuth2 flow
3. Retries the original request once

### Rate limit retry

On 429 (Too Many Requests) or 5xx, the CLI uses exponential backoff:
- Attempt 1: wait 2s
- Attempt 2: wait 4s
- Attempt 3: wait 8s
- Attempt 4: wait 16s
- Attempt 5: fail

Jitter of +/-25% is added to each delay. `Retry-After` headers are respected when present.

### Mutating request throttling

POST/PUT/DELETE requests are limited to 8 concurrent in-flight requests to avoid hitting API quotas.

## Embedded Schema

The CLI embeds a compact API schema index at build time via `//go:embed`. This powers:

```
aads schema campaigns           # list all campaign endpoints
aads schema --type Campaign     # show Campaign type fields
aads schema --method post       # list endpoints for a method
aads schema keyword             # fuzzy search across endpoints and types
```

Agents can introspect the API surface without network access.

## Distribution

- **Install script**: `curl -fsSL https://raw.githubusercontent.com/imesart/apple-ads-cli/main/install.sh | sh`
- **Homebrew**: `brew install imesart/tap/aads`
- **Binary releases**: GitHub Releases with signed macOS archives (`darwin/arm64`, `darwin/amd64`), Linux archives (`linux/arm64`, `linux/amd64`), and Windows zip files (`windows/arm64`, `windows/amd64`)
- **From source**: `go install github.com/imesart/apple-ads-cli@latest`
- **Release automation**: GitHub Actions builds `release/` archives via `make release-artifacts`, uploads them to GitHub Releases, and updates the Homebrew tap

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | General error |
| 2 | Usage error (bad flags, missing required args) |
| 3 | Authentication error |
| 4 | API error (4xx) |
| 5 | Network / server error (5xx, timeout) |
| 6 | Safety limit exceeded (use --force to override) |

## UX Requirements

### Interactive use

- Concise help text with examples in command help
- Sensible defaults for common operations
- Readable table output

### Help text

Flag help text must encode required/default status directly in the flag description.

- For flags that map to API fields:
  - required fields from the API schema must be marked with `(required)` when they do not have an API default.
  - if the API schema defines a non-empty, non-null default value, help text must show that default and the CLI must apply that default itself instead of relying on the API.
  - optional API fields with no default, or with default `null` / `""`, must not show a default in help text.
- Enum-style flags with defaults must mark the default inline next to the value, for example `SEARCH (default) | DISPLAY`.
- Boolean flags are a special case: keep the default with the flag form itself, for example `--check=false`.
- Non-enum flags with meaningful non-boolean defaults should use a parenthesized suffix such as `(default 1000)` when a default should be shown.
- When a flag description includes both behavioral notes and required/default markers, keep the behavioral notes first and keep `(required)` as the final suffix. Example: `Campaign name (accepts template variables like %(fieldName) or %(FIELD_NAME)) (required)`.
- Shorthand flags are a special case: describe the full flag name, do not repeat accepted values, and end the help text with `(shorthand)`. For example, `-f` should describe `--format (shorthand)`.
- `--from-json` is a special case: do not mark it with `(required)` in flag help. For commands that support both shortcut flags and `--from-json`, the help description must explicitly say that users can use shortcut flags or `--from-json`.
- Help descriptions must not repeat flag semantics that are already present in flag help.
- Put flag-specific guidance in the flag help text whenever it fits cleanly there.
- Keep flag-specific guidance in command descriptions only when it is needed to explain multi-flag interactions, merge/fetch behavior, or other details that do not fit well in a single flag help string.
- Required/default status must not be explained only in surrounding prose when it can be attached directly to the flag help text.

### Scripting

- Stable JSON output structure
- No noise on stdout in machine mode
- Progress messages and warnings only on stderr
- Non-zero exit codes on failure (see Exit Codes)

### Safety

- Destructive commands (delete) require `--confirm`
- Budget/bid safety thresholds via config (see Safety Limits)
- Secrets (tokens, private keys) must never appear in logs or error output

## Error Handling Requirements

The CLI must classify errors into distinct categories:

- Invalid input (bad flags, missing required args)
- Missing auth (no credentials configured)
- Unauthorized (401)
- Forbidden (403)
- Not found (404)
- Rate limited (429)
- API / server failure (5xx)
- Internal / config failure

Error output should be short, specific, and actionable. Each category maps to a distinct exit code (see Exit Codes).

## Performance Requirements

- Startup should feel immediate (no unnecessary initialization)
- Token generation and refresh must be cached
- Auto-pagination should avoid unnecessary duplicate requests
- Retries must use bounded exponential backoff

## Testing Requirements

### Unit tests

- Auth: JWT signing, token exchange mock, cache expiry
- Client: middleware chain, request building, response parsing
- Types: JSON serialization roundtrip for all types
- Config: loading, profile resolution, env var precedence
- Safety: budget/bid limit validation

### Snapshot tests

One golden file per endpoint. Tests build a request from Go types and assert the full HTTP request matches the golden file. Catches:

- URL path construction errors
- Missing or wrong headers
- Request body serialization bugs
- HTTP method mismatches

Initial golden files derived from mihai8804858's Swift snapshot tests.

### Mocked-response client tests

HTTP client tests with mocked Apple Ads API responses. These test the response side: status code handling, JSON decoding, error parsing, pagination assembly, and retry behavior. Snapshot tests verify request construction; mocked-response tests verify response handling.

### Command-level tests

End-to-end tests for major command groups that exercise the full CLI path: flag parsing, command routing, API interaction (with mocked HTTP), and output formatting. These catch integration issues that unit tests miss.

Run the mocked command-integration suite with:

```bash
make test-integration
```

Local-only commands should also be covered here. Profile command tests must use
`AADS_CONFIG_DIR` (or an equivalent in-process config-dir override) to isolate
their config and token files from the user's default `~/.aads` directory.

### Example data rules

- Do not commit production Apple Ads identifiers in tests, help text, examples, fixtures, snapshots, or generated docs.
- IDs used in mocked tests and examples must be clearly synthetic placeholders, not values copied from real API traffic.
- Live integration tests may use real IDs at runtime, but they must discover them by calling the API during the test run rather than hardcoding them in the repository.
- Keyword and search-term examples in help text, docs, fixtures, and tests should stay within a clearly fake fitness-app domain.
- Prefer obviously synthetic examples such as `FitTrack`, `fitness coach`, `home workout planner`, and placeholder IDs like `900001`, `101`, or `201`.

### Integration tests (optional, requires credentials)

Gated behind `AADS_INTEGRATION_TEST=1`. Run against the real API with a test organization. Verify end-to-end: auth -> create campaign -> list -> update -> delete.

Run the live command-integration suite with:

```bash
make test-live
```

Not run in CI (requires Apple Ads credentials). Run manually before releases.
