# aads - Apple Ads CLI

[![Latest release](https://img.shields.io/github/v/release/imesart/apple-ads-cli)](https://github.com/imesart/apple-ads-cli/releases)
[![API version](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fimesart.com%2Fip%2Faads-openapi-latest.php&query=%24.info.version&prefix=v&label=API&color=cyan)](docs/apple_ads/openapi-latest.json)
[![Go version](https://img.shields.io/github/go-mod/go-version/imesart/apple-ads-cli?logo=go)](go.mod)
[![License](https://img.shields.io/github/license/imesart/apple-ads-cli)](LICENSE)

A command-line interface for the Apple Ads (formerly Apple Search Ads) Campaign Management API, with **composable commands** to simplify common workflows, and **shareable campaign/ad group/keyword JSON** for export and import. Made for humans, usable by AI agents. Independent project, not affiliated with Apple.

```sh
# Find ad groups with less than 10 impressions in the last week, and increase
# their default bid by 10%, also increasing bids for keyword with default bid.
# Remove "--check" to actually increase the bids.
aads campaigns list --filter "status = ENABLED" \
    | aads reports adgroups --campaign-id - --start -7d --end now \
        --filter "impressions < 10" \
    | aads adgroups update --campaign-id - --adgroup-id - \
        --update-inherited-bids --default-bid +10% --check
```

## Install

With [Homebrew](https://brew.sh/):

```sh
brew install imesart/tap/aads
```

With the install script:

```sh
curl -fsSL https://raw.githubusercontent.com/imesart/apple-ads-cli/main/install.sh | sh
```

Check the installed binary:

```sh
aads version
```

## Configure Authentication

Set up Apple Ads API access. `aads` uses Apple Ads OAuth credentials to obtain and refresh API access tokens automatically.

```sh
aads profiles create --interactive
```

Or read the detailed manual setup guide in [docs/auth.md](docs/auth.md).

## Examples

Find all active ad groups:

```sh
aads campaigns list --filter "status = ENABLED" \
    | aads adgroups list --campaign-id - \
        --fields campaignId,campaignName,id,name,defaultBidAmount \
        --sort "campaignName:asc" --sort "name:asc"
```

Find the bid for ad groups with less than 10 impressions in the last week:

```sh
aads campaigns list --filter "status = ENABLED" \
    | aads reports adgroups --campaign-id - --start -7d --end now \
        --filter "impressions < 10" -f pipe \
    | aads adgroups get --campaign-id - --adgroup-id - \
        --fields campaignId,campaignName,id,name,defaultBidAmount
```

Find spend for active ad groups this past week:

```sh
aads campaigns list --filter "status = ENABLED" \
    | aads adgroups list --campaign-id - -f pipe \
    | aads reports adgroups --campaign-id - --adgroup-id - --start -7d --end now \
        --filter "localSpend > 0" --sort "localSpend:desc" \
        --fields campaignId,campaignName,adgroupId,adgroupName,impressions,localSpend,totalAvgCPI
```

Find 1 ad group campaigns that max their daily budget, or almost:

```sh
aads campaigns list --filter "status = ENABLED" \
    | aads reports adgroups --campaign-id - --start -1d --end now \
        --filter "localSpend > dailyBudgetAmount" --sort "localSpend:desc" \
        --fields campaignId,campaignName,adgroupId,adgroupName,impressions,localSpend,totalAvgCPI
```

Copy and adapt US campaigns, ad groups, and keywords for the UK. Remove `--check` to actually create:

```sh
aads structure export --scope campaigns --redact-names \
        --campaigns-filter "name STARTSWITH FitTrack - US" \
        --keywords-filter "status = ACTIVE" \
    | aads structure import --from-structure @- \
        --countries-or-regions GB \
        --bid "" --default-bid 1.3 --check
```

Assuming existing campaigns named `FitTrack - US - Discovery`, `FitTrack - US - Brand`, `FitTrack - US - Competitor`, and `FitTrack - US - Category`, this creates `FitTrack - GB - Discovery`, `FitTrack - GB - Brand`, `FitTrack - GB - Competitor`, and `FitTrack - GB - Category` with the same keywords and negative keywords as the US campaigns, while adapting the default bid and any keyword bid that had the default bid.

Find potential new search terms to move to a new campaign:

```sh
aads campaigns list --filter "status = ENABLED" \
    | aads adgroups list --campaign-id - -f pipe \
    | aads reports searchterms --campaign-id - --adgroup-id - --start -7d --end now \
        --filter "localSpend > 5" \
        --filter "impressions > 10" \
        --filter "totalAvgCPI < 3.5" \
        --filter "searchTermText != null" \
        --fields campaignId,campaignName,adgroupId,adgroupName,searchTermText,impressions,localSpend,totalAvgCPI
```

Copy negative keywords from one campaign to another:

```sh
aads negatives list --campaign-id 123 --fields text,matchType --format json \
    | aads negatives create --campaign-id 456 --from-json @- --check
```

## Commands

Use `aads --help`, `aads <group> --help`, and `aads <group> <command> --help` for current help. The generated full command reference is [docs/commands.md](docs/commands.md).

### campaigns

Manage campaigns.

```sh
aads campaigns list --filter "status = ENABLED"
aads campaigns get --campaign-id 123
aads campaigns create --name "FitTrack US Search" --adam-id 900001 --daily-budget-amount 50 --countries-or-regions US --check
aads campaigns update --campaign-id 123 --daily-budget-amount 75 --check
aads campaigns delete --campaign-id 123 --confirm --check
```

### adgroups

Manage ad groups inside campaigns.

```sh
aads adgroups list --campaign-id 123 --filter "status = ENABLED"
aads adgroups get --campaign-id 123 --adgroup-id 456
aads adgroups create --campaign-id 123 --name "FitTrack Brand" --default-bid 1.50 --check
aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid +10% --check
aads adgroups delete --campaign-id 123 --adgroup-id 456 --confirm --check
```

### keywords

Manage targeting keywords.

```sh
aads keywords list --campaign-id 123 --adgroup-id 456 --filter "status = ACTIVE"
aads keywords get --campaign-id 123 --adgroup-id 456 --keyword-id 789
aads keywords create --campaign-id 123 --adgroup-id 456 --text "fitness tracker,step counter" --match-type EXACT --bid 1.25 --check
aads keywords update --campaign-id 123 --adgroup-id 456 --keyword-id 789 --bid +10% --check
aads keywords delete --campaign-id 123 --adgroup-id 456 --from-json '[789,123]' --confirm --check
aads keywords delete --campaign-id 123 --adgroup-id 456 --keyword-id 789 --confirm --check
```

### negatives

Manage campaign-level and ad-group-level negative keywords.

```sh
aads negatives list --campaign-id 123
aads negatives list --campaign-id 123 --adgroup-id 456
aads negatives get --campaign-id 123 --keyword-id 789
aads negatives create --campaign-id 123 --text "free workout,fitness wallpaper" --match-type EXACT --check
aads negatives update --campaign-id 123 --keyword-id 789 --status PAUSED --check
aads negatives delete --campaign-id 123 --from-json @negatives.json --confirm --check
```

### reports

Generate campaign, ad group, keyword, search term, and ad reports.

```sh
aads reports campaigns --start -7d --end now --fields campaignId,campaignName,impressions,localSpend
aads reports adgroups --campaign-id 123 --start -7d --end now --filter "impressions > 10"
aads reports keywords --campaign-id 123 --adgroup-id 456 --start -7d --end now
aads reports searchterms --campaign-id 123 --adgroup-id 456 --start -7d --end now
aads reports ads --campaign-id 123 --start -7d --end now
```

### impression-share

Create and retrieve custom impression share reports.

```sh
aads impression-share create --name "FitTrack weekly share" --granularity WEEKLY --dateRange LAST_WEEK
aads impression-share list
aads impression-share get --report-id 123
```

### structure

Export and import campaign/ad group structures as JSON.

```sh
aads structure export --scope campaigns --campaigns-filter "name STARTSWITH FitTrack - US" --pretty
aads structure import --from-structure @structure.json --countries-or-regions GB --default-bid 1.30 --check
```

### Other Commands

Additional command groups cover ads, creatives, budget orders, product pages, ad rejections, apps, geolocation, organizations/user context (`orgs`, backed by Apple Ads ACLs), profiles, schema lookup, version output, and shell completion.

```sh
aads apps search --query "FitTrack" --only-owned-apps
aads geo search --query "Luxembourg"
aads profiles list
aads schema campaigns
aads completion zsh
```

## Common Flags

### ID Flags

Resource commands use explicit ID flags such as `--campaign-id`, `--adgroup-id`, `--keyword-id`, `--ad-id`, `--creative-id`, and `--budget-order-id`. Negative keyword commands also use `--keyword-id`, with the help text identifying it as a negative keyword ID.

Many ID flags accept `-` to read IDs from stdin, which makes command pipelines possible. For example, `--campaign-id -` reads campaign IDs from the previous command. Use `--format ids` or `-f ids` when you only want to pass IDs onward. Use `--format pipe` or `-f pipe` when you want to pass IDs plus all available fields and carried stdin context to the next `aads` command.

### Output Format

Most commands support:

```sh
--format json
--format table
--format yaml
--format markdown
--format ids
--format pipe
```

`-f` is shorthand for `--format`. JSON is the default when stdout is piped; table output is the default for an interactive terminal. Use `--pretty` to pretty-print JSON.

- `json`: machine-readable output; default when stdout is piped
- `table`: human-readable table; default in a TTY
- `yaml`: YAML output
- `markdown`: markdown table output
- `ids`: tab-separated IDs with a header row, useful when downstream commands only need IDs
- `pipe`: tab-separated records with a header row, useful when downstream commands need IDs plus carried fields such as `campaignName`

### Fields

Use `--fields` to choose output fields:

```sh
aads campaigns list --fields id,name,status,dailyBudgetAmount
aads reports adgroups --campaign-id 123 --start -7d --end now --fields campaignId,adGroupId,impressions,localSpend
```

Field names accept API-style JSON keys and table-style column names. For example, `campaignId` and `CAMPAIGN_ID` refer to the same field; `adGroupId`, `AD_GROUP_ID`, and compact input aliases like `ADGROUP_ID` are accepted where relevant.

Commands that output one entity type also accept useful aliases. For example, ad group output accepts `adGroupId` as an alias for `id`, and `adGroupName` as an alias for `name`.

Fields from stdin can be carried through pipelines. `-f pipe` is the short form when you want the current command to emit a downstream-friendly tab-separated row set without manually listing intermediate `--fields`. Downstream filters and output can use carried fields such as `campaignName`, `dailyBudgetAmount`, `adGroupName`, and `defaultBidAmount`.

### Filter, Selector, And Sort

Use repeatable `--filter` flags on list and report commands:

```sh
aads campaigns list --filter "status = ENABLED" --filter "name STARTSWITH FitTrack"
aads reports adgroups --campaign-id 123 --start -7d --end now --filter "localSpend > 10"
```

Available filter operators:

| Operator | Meaning |
|---|---|
| `=` / `EQUALS` | Equal |
| `!=` / `NOT_EQUALS` | Not equal |
| `<` / `LESS_THAN` | Less than |
| `>` / `GREATER_THAN` | Greater than |
| `CONTAINS` | Contains text or collection value |
| `STARTSWITH` | Starts with text |
| `ENDSWITH` | Ends with text |
| `IN` | Value is in a list |
| `BETWEEN` | Value is between two values |
| `CONTAINS_ALL` | Collection contains all values |
| `CONTAINS_ANY` | Collection contains any value |

Use `--selector` when you need to provide the Apple Ads selector JSON directly:

```sh
aads campaigns list --selector @selector.json
```

Use repeatable `--sort` flags on commands that support sorting:

```sh
aads campaigns list --sort "name:asc" --sort "id:desc"
```

Like `--fields`, `--filter` and `--sort` accept JSON keys and table column names.

### Date And Time Flags

Report commands use date flags such as `--start` and `--end`:

```sh
aads reports campaigns --start -7d --end now
aads reports campaigns --start 2026-04-01 --end 2026-04-15
```

Mutation commands use time flags such as `--start-time` and `--end-time`:

```sh
aads campaigns create --name "FitTrack US" --adam-id 900001 --daily-budget-amount 50 --countries-or-regions US --start-time now --check
aads adgroups update --campaign-id 123 --adgroup-id 456 --end-time +30d --check
```

Date and time flags accept `YYYY-MM-DD`, `now`, and signed relative expressions like `-7d`, `+2w`, `+1mo`, and `+1y`. Time flags also accept [ISO/RFC3339](https://datatracker.ietf.org/doc/html/rfc3339) datetime values.

### JSON Input

Commands with `--from-json` accept inline JSON, file input with `@file.json`, and stdin with `@-`:

```sh
aads campaigns create --from-json @campaign.json --check
aads keywords create --campaign-id 123 --adgroup-id 456 --from-json @-
aads ads update --campaign-id 123 --adgroup-id 456 --ad-id 789 --from-json '{"status":"PAUSED"}' --check
```

Structure import uses `--from-structure` with the same input convention:

```sh
aads structure import --from-structure @structure.json --check
```

### Global Flags

Useful global flags include:

```sh
--profile NAME
--org-id ID
--config-dir DIR
--currency EUR
--verbose
--no-color
```

`--profile` chooses a named configuration profile. `--org-id` overrides the configured Apple Ads organization ID for the command.

## Safety

`aads` is designed for day-to-day campaign operations, so mutating commands include guardrails:

- `--check` validates and summarizes without sending mutating API requests.
- Delete commands require `--confirm`.
- Profile safety limits can cap daily budgets, total budgets, bids, and CPA goals.
- `--force` skips configured safety limit checks.

Set safety limits when creating or updating a profile:

```sh
aads profiles update \
    --name default \
    --default-currency USD \
    --max-daily-budget 1000 \
    --max-budget 50000 \
    --max-bid 10 \
    --max-cpa-goal 25
```

Limits are stored as decimal text in the profile's `default_currency`. Use `0` or an empty value to disable a limit.

## Contributing

Issues and pull requests are welcome. Before making substantial changes, read [CONTRIBUTING.md](CONTRIBUTING.md), [docs/requirements.md](docs/requirements.md), and [docs/architecture.md](docs/architecture.md).

## License

Licensed under the GNU Affero General Public License v3.0. See [LICENSE](LICENSE).

Apple and Apple Ads are trademarks of Apple Inc. This is an independent project and is not affiliated with, endorsed by, or sponsored by Apple.

## Author

Copyright (c) 2026 Imesart S.a.r.l.
