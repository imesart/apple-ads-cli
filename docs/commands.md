# aads Command Reference

> Generated with `make commands-doc`. Do not edit manually.

Fast, lightweight CLI for the Apple Ads Campaign Management API.

```
USAGE
  aads <subcommand> [flags]

SUBCOMMANDS
  campaigns         Manage campaigns.
  adgroups          Manage ad groups.
  keywords          Manage targeting keywords.
  negatives         Manage negative keywords (campaign and ad group level).
  ads               Manage ads.
  creatives         Manage creatives.
  budgetorders      Manage budget orders.
  product-pages     Manage custom product pages.
  ad-rejections     Manage ad creative rejection reasons.
  reports           Generate reports.
  impression-share  Manage impression share reports.
  apps              Search and inspect apps.
  geo               Search geolocations.
  orgs              Manage organizations and user context (Apple Ads ACLs).
  structure         Export and import campaign/ad group structures as JSON.
  profiles          Manage configuration profiles.
  version           Print the CLI version and target Apple Ads API version.
  schema            Query embedded API schema information.
  completion        Generate shell completion scripts.
  help              Show help for a command.
```

## Global Flags

```
FLAGS
  -config-dir string  Configuration directory
  -currency string    Override currency for money fields
  -no-color=false     Disable color output
  -org-id string      Override org ID from config
  -p string           --profile (shorthand)
  -profile string     Config profile name
  -v=false            --verbose (shorthand)
  -verbose=false      Verbose output, show HTTP requests/responses
  -version=false      Print version and exit
```

---

## campaigns

Manage campaigns.

```
USAGE
  aads campaigns <subcommand>

SUBCOMMANDS
  list    List campaigns.
  get     Get a campaign by ID.
  create  Create a campaign.
  update  Update a campaign.
  delete  Delete a campaign.
```

### campaigns list

List campaigns.

```
USAGE
  aads campaigns list [flags]

List campaigns, with optional filtering.

Use --filter / --sort flags, or --selector for JSON selector.

Filter operators:
  EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

Searchable and filterable fields:
  id, orgId, name, adamId, status, servingStatus, displayStatus,
  adChannelType, supplySources, billingEvent, paymentModel,
  countriesOrRegions, budgetAmount, dailyBudgetAmount

Selector JSON keys: conditions, fields, orderBy, pagination.

Examples:
  aads campaigns list
  aads campaigns list --filter "status=ENABLED"
  aads campaigns list --filter "name STARTSWITH MyApp" --filter "adamId IN [123, 456]" --filter "dailyBudgetAmount BETWEEN [10, 100]" --sort "name:asc" --sort "id:desc"
```

```
FLAGS
  -f json           --format (shorthand)
  -fields string    Comma-separated output fields to include
  -filter value     Filter: "field=value" or "field OPERATOR value" (repeatable)
  -format json      Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -limit 0          Maximum results; 0 fetches all pages
  -offset 0         Starting offset
  -pretty=false     Pretty-print JSON even when stdout is not a TTY
  -selector string  Selector input: inline JSON, @file.json, or @- for stdin
  -sort value       Sort: "field:asc" or "field:desc" (repeatable)
```

### campaigns get

Get a campaign by ID.

```
USAGE
  aads campaigns get --campaign-id ID
```

```
FLAGS
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
```

### campaigns create

Create a campaign.

```
USAGE
  aads campaigns create --name NAME --adam-id ID --daily-budget-amount AMT --countries-or-regions CC [flags]

Use shortcut flags or --from-json for the full JSON body.

JSON keys (for --from-json):
  adamId              integer   (required) App Store app ID
  name                string    (required) Campaign name
  adChannelType       string    SEARCH (default) | DISPLAY
  supplySources       [string]  APPSTORE_SEARCH_RESULTS (default) |
                                APPSTORE_SEARCH_TAB |
                                APPSTORE_PRODUCT_PAGES_BROWSE | APPSTORE_TODAY_TAB
  countriesOrRegions  [string]  (required) ISO alpha-2 country codes
  billingEvent        string    TAPS (default) | IMPRESSIONS
  dailyBudgetAmount   Money     (required) Daily budget cap
  budgetAmount        Money     DEPRECATED: Total (lifetime) budget cap
  locInvoiceDetails   object    {billingContactEmail, buyerEmail, buyerName,
                                 clientName, orderNumber}
  status              string    ENABLED (default) | PAUSED
  startTime           string    ISO 8601 datetime
  endTime             string    ISO 8601 datetime

Money object: {"amount": "10.00", "currency": "USD"}

Examples:
  aads campaigns create --name "FitTrack US Search" --adam-id 900001 --daily-budget-amount 50 --countries-or-regions US
  aads campaigns create --name "FitTrack EU Search" --adam-id 900001 --daily-budget-amount "50 EUR" --countries-or-regions "GB,DE,FR"
  aads campaigns create --name "FitTrack %(COUNTRIES_OR_REGIONS) %(adChannelType)" --adam-id 900001 --daily-budget-amount 50 --countries-or-regions "DE,FR"
  aads campaigns create --name "FitTrack LOC" --adam-id 900001 --daily-budget-amount 50 --countries-or-regions US --loc-invoice-details '{"orderNumber":"PO-123"}'
  aads campaigns create --from-json campaign.json
```

```
FLAGS
  -ad-channel-type SEARCH                  SEARCH (default) | DISPLAY
  -adam-id string                          App Store app ID (required)
  -billing-event TAPS                      TAPS (default) | IMPRESSIONS
  -budget-amount string                    DEPRECATED: Total budget (AMOUNT or "AMOUNT CURRENCY"; bare amount uses default currency)
  -check=false                             Validate and summarize without sending the request
  -countries-or-regions string             Comma-separated country codes (required)
  -daily-budget-amount string              Daily budget (AMOUNT or "AMOUNT CURRENCY"; bare amount uses default currency) (required)
  -end-time string                         End time (UTC; accepts ISO 8601/RFC3339 datetime, YYYY-MM-DD, now, or signed offset like +5d)
  -f json                                  --format (shorthand)
  -fields string                           Comma-separated output fields to include
  -force=false                             Skip safety checks
  -format json                             Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string                        Path to JSON body file (or - for stdin)
  -loc-invoice-details string              LOC invoice details JSON: inline JSON, @file.json, or @- for stdin
  -name string                             Campaign name (accepts template variables like %(fieldName) or %(FIELD_NAME)) (required)
  -pretty=false                            Pretty-print JSON even when stdout is not a TTY
  -start-time string                       Start time (UTC; accepts ISO 8601/RFC3339 datetime, YYYY-MM-DD, now, or signed offset like +5d)
  -status string                           ENABLED (default) | PAUSED (also 1/0, enable/pause)
  -supply-sources APPSTORE_SEARCH_RESULTS  APPSTORE_SEARCH_RESULTS (default) | APPSTORE_SEARCH_TAB | APPSTORE_PRODUCT_PAGES_BROWSE | APPSTORE_TODAY_TAB
```

### campaigns update

Update a campaign.

```
USAGE
  aads campaigns update --campaign-id ID [flags]

Update a campaign. The API accepts partial updates.
The CLI wraps the body in a {"campaign": ...} envelope automatically.

Use shortcut flags for common changes, or --from-json for arbitrary JSON.
Shortcut flags can be combined with each other.

JSON keys (all optional):
  name                string    Campaign name
  status              string    ENABLED | PAUSED
  budgetAmount        Money     DEPRECATED: Total (lifetime) budget cap
  dailyBudgetAmount   Money     Daily budget cap
  countriesOrRegions  [string]  ISO alpha-2 country codes
  budgetOrders        [integer] Budget order IDs (LOC only)
  locInvoiceDetails   object    {billingContactEmail, buyerEmail, buyerName,
                                clientName, orderNumber}
  startTime           string    ISO 8601 datetime
  endTime             string    ISO 8601 datetime

Money object: {"amount": "10.00", "currency": "USD"}

Examples:
  aads campaigns update --campaign-id 123 --status 0
  aads campaigns update --campaign-id 123 --daily-budget-amount 75.00
  aads campaigns update --campaign-id 123 --daily-budget-amount "75.00 EUR"
  aads campaigns update --campaign-id 123 --name "New Name" --status ENABLED
  aads campaigns update --campaign-id 123 --loc-invoice-details '{"orderNumber":"PO-123"}'
  aads campaigns update --campaign-id 123 --countries-or-regions "US,GB,CA"
  aads campaigns update --campaign-id 123 --from-json changes.json
```

```
FLAGS
  -budget-amount string         DEPRECATED: Total budget (AMOUNT or "AMOUNT CURRENCY"; bare amount uses default currency)
  -campaign-id string           Campaign ID (or - to read IDs from stdin) (required)
  -check=false                  Validate and summarize without sending the request
  -countries-or-regions string  Comma-separated country codes (e.g. US,GB)
  -daily-budget-amount string   Daily budget (AMOUNT or "AMOUNT CURRENCY"; bare amount uses default currency)
  -f json                       --format (shorthand)
  -fields string                Comma-separated output fields to include
  -force=false                  Skip safety checks
  -format json                  Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string             JSON body input: inline JSON, @file.json, or @- for stdin
  -loc-invoice-details string   LOC invoice details JSON: inline JSON, @file.json, or @- for stdin
  -name string                  Campaign name
  -pretty=false                 Pretty-print JSON even when stdout is not a TTY
  -status string                ENABLED | PAUSED (also 1/0, enable/pause)
```

### campaigns delete

Delete a campaign.

```
USAGE
  aads campaigns delete --campaign-id ID --confirm
```

```
FLAGS
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -check=false         Validate and summarize without sending the request
  -confirm=false       Confirm deletion
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
```

---

## adgroups

Manage ad groups.

```
USAGE
  aads adgroups <subcommand>

SUBCOMMANDS
  list    List ad groups.
  get     Get an ad group by ID.
  create  Create an ad group.
  update  Update an ad group.
  delete  Delete an ad group.
```

### adgroups list

List ad groups.

```
USAGE
  aads adgroups list [--campaign-id ID] [flags]

List ad groups, with optional filtering.

With --campaign-id, lists ad groups in that campaign.
Without --campaign-id, searches across all campaigns.

Use --filter / --sort flags, or --selector for JSON selector.

Filter operators:
  EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

Searchable and filterable fields:
  id, campaignId, name, status, servingStatus, displayStatus,
  pricingModel, defaultCpcBid, startTime, endTime

Selector JSON keys: conditions, fields, orderBy, pagination.

Examples:
  aads adgroups list
  aads adgroups list --campaign-id 123
  aads adgroups list --campaign-id 123 --filter "status=ENABLED" --sort "name:asc"
  aads adgroups list --filter "name STARTSWITH Search"
```

```
FLAGS
  -campaign-id string  Campaign ID (omit to search all campaigns) (or - to read IDs from stdin)
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -filter value        Filter: "field=value" or "field OPERATOR value" (repeatable)
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -limit 0             Maximum results; 0 fetches all pages
  -offset 0            Starting offset
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
  -selector string     Selector input: inline JSON, @file.json, or @- for stdin
  -sort value          Sort: "field:asc" or "field:desc" (repeatable)
```

### adgroups get

Get an ad group by ID.

```
USAGE
  aads adgroups get --campaign-id CID --adgroup-id ID
```

```
FLAGS
  -adgroup-id string   Ad Group ID (or - to read IDs from stdin) (required)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
```

### adgroups create

Create an ad group.

```
USAGE
  aads adgroups create --campaign-id CID --name NAME --default-bid BID [flags]

Use shortcut flags or --from-json for the full JSON body.

JSON keys (for --from-json):
  name                    string  (required) Ad group name
  pricingModel            string  CPC (default)
  defaultBidAmount        Money   (required) Default bid
  status                  string  ENABLED (default) | PAUSED
  cpaGoal                 Money   Target cost per acquisition (Search only)
  automatedKeywordsOptIn  bool    Enable automated keywords
  startTime               string  ISO 8601 datetime
  endTime                 string  ISO 8601 datetime
  targetingDimensions     object  Audience targeting

Money object: {"amount": "1.50", "currency": "USD"}

Examples:
  aads adgroups create --campaign-id 123 --name "Search Group" --default-bid 1.50
  aads adgroups create --campaign-id 123 --name "Search Group" --default-bid "1.50 USD" --status PAUSED
  aads adgroups create --campaign-id 123 --name "Search Group %(PRICING_MODEL)" --default-bid 1.50 --cpa-goal 2.00 --age 18-65 --gender M,F
  aads adgroups create --campaign-id 123 --from-json adgroup.json
```

```
FLAGS
  -admin-area string                Admin area targeting (comma-separated: US|CA,US|NY)
  -age string                       Age range (e.g. "18-65")
  -automated-keywords-opt-in=false  Enable automated keywords
  -campaign-id string               Campaign ID (or - to read IDs from stdin) (required)
  -check=false                      Validate and summarize without sending the request
  -country-code string              Country targeting (comma-separated: US,GB)
  -cpa-goal string                  CPA goal amount (AMOUNT or "AMOUNT CURRENCY"; bare amount uses default currency); Search campaigns only
  -default-bid string               Default CPC bid amount (AMOUNT or "AMOUNT CURRENCY"; bare amount uses default currency) (required)
  -device-class string              Device class (comma-separated: IPHONE,IPAD)
  -end-time string                  End time (UTC; accepts ISO 8601/RFC3339 datetime, YYYY-MM-DD, now, or signed offset like +5d)
  -f json                           --format (shorthand)
  -fields string                    Comma-separated output fields to include
  -force=false                      Skip safety checks
  -format json                      Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string                 JSON body input: inline JSON, @file.json, or @- for stdin
  -gender string                    Gender targeting (comma-separated: M,F)
  -locality string                  Locality targeting (comma-separated: US|CA|Los Angeles)
  -name string                      Ad group name (accepts template variables like %(fieldName) or %(FIELD_NAME)) (required)
  -pretty=false                     Pretty-print JSON even when stdout is not a TTY
  -start-time string                Start time (UTC; accepts ISO 8601/RFC3339 datetime, YYYY-MM-DD, now, or signed offset like +5d)
  -status string                    ENABLED (default) | PAUSED (also 1/0, enable/pause)
```

### adgroups update

Update an ad group.

```
USAGE
  aads adgroups update --campaign-id CID --adgroup-id ID [flags]

Update an ad group. The API accepts partial updates.

Use shortcut flags for common changes, or --from-json for arbitrary JSON fields,
or --merge to fetch the current state first (needed for targeting dimensions).
Shortcut flags can be combined with each other.

Bid/CPA flags also accept math expressions:
  Delta:       +1, -0.50, "+1 USD" (adjusts current value)
  Multiplier:  x1.1, x0.9 (multiplies current value)
  Percent:     +10%, -15% (percent adjustment)
Relative expressions fetch the current value from the API first.

JSON keys (all optional):
  name                    string  Ad group name
  status                  string  ENABLED | PAUSED
  defaultBidAmount        Money   Default bid
  cpaGoal                 Money   Target cost per acquisition
  automatedKeywordsOptIn  bool    Enable automated keywords
  startTime               string  ISO 8601 datetime
  endTime                 string  ISO 8601 datetime
  targetingDimensions     object  Audience targeting (see adgroups create --help)

Money object: {"amount": "1.50", "currency": "USD"}

The --cpa-goal flag requires the campaign's adChannelType to be SEARCH;
the campaign is fetched automatically to verify this.

Use --update-inherited-bids with --default-bid to also update keyword-level
bid overrides that currently match the ad group's old default bid.

Examples:
  aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid 1.50
  aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid "1.50 EUR"
  aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid +0.50
  aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid x1.1
  aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid +10%
  aads adgroups update --campaign-id 123 --adgroup-id 456 --default-bid +10% --update-inherited-bids
  aads adgroups update --campaign-id 123 --adgroup-id 456 --cpa-goal 2.00
  aads adgroups update --campaign-id 123 --adgroup-id 456 --status 0
  aads adgroups update --campaign-id 123 --adgroup-id 456 --name "New Name" --status ENABLED
  aads adgroups update --campaign-id 123 --adgroup-id 456 --start-time 2025-06-01T00:00:00.000
  aads adgroups update --campaign-id 123 --adgroup-id 456 --from-json changes.json
  aads adgroups update --campaign-id 123 --adgroup-id 456 --merge --from-json targeting.json
```

```
FLAGS
  -adgroup-id string            Ad Group ID (or - to read IDs from stdin) (required)
  -campaign-id string           Campaign ID (or - to read IDs from stdin) (required)
  -check=false                  Validate and summarize without sending the request
  -cpa-goal string              CPA goal (AMOUNT or "AMOUNT CURRENCY"; bare amount uses default currency), delta (+1), multiplier (x1.1), or percent (+10%); Search only
  -default-bid string           Default CPC bid (AMOUNT or "AMOUNT CURRENCY"; bare amount uses default currency), delta (+1), multiplier (x1.1), or percent (+10%)
  -end-time string              End time (UTC; accepts ISO 8601/RFC3339 datetime, YYYY-MM-DD, now, or signed offset like +5d)
  -f json                       --format (shorthand)
  -fields string                Comma-separated output fields to include
  -force=false                  Skip safety checks
  -format json                  Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string             JSON body input: inline JSON, @file.json, or @- for stdin
  -merge=false                  Fetch current ad group and merge changes first (for targeting dimensions)
  -name string                  Ad group name
  -pretty=false                 Pretty-print JSON even when stdout is not a TTY
  -start-time string            Start time (UTC; accepts ISO 8601/RFC3339 datetime, YYYY-MM-DD, now, or signed offset like +5d)
  -status string                ENABLED | PAUSED (also 1/0, enable/pause)
  -update-inherited-bids=false  Also update keyword bids that currently inherit this ad group's default bid
```

### adgroups delete

Delete an ad group.

```
USAGE
  aads adgroups delete --campaign-id CID --adgroup-id ID --confirm
```

```
FLAGS
  -adgroup-id string   Ad Group ID (or - to read IDs from stdin) (required)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -check=false         Validate and summarize without sending the request
  -confirm=false       Confirm deletion
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
```

---

## keywords

Manage targeting keywords.

```
USAGE
  aads keywords <subcommand>

SUBCOMMANDS
  list    List targeting keywords.
  get     Get a targeting keyword.
  create  Create targeting keywords.
  update  Update a targeting keyword.
  delete  Delete targeting keywords.
```

### keywords list

List targeting keywords.

```
USAGE
  aads keywords list --campaign-id CID --adgroup-id AGID [flags]

List targeting keywords, with optional filtering.

Lists keywords in a specific ad group.

Use --filter / --sort flags, or --selector for JSON selector.

Filter operators:
  EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

Searchable and filterable fields:
  id, campaignId, adGroupId, text, matchType, status, bidAmount

Selector JSON keys: conditions, fields, orderBy, pagination.

Examples:
  aads keywords list --campaign-id 123 --adgroup-id 456
  aads keywords list --campaign-id 123 --adgroup-id 456 --filter "matchType=EXACT"
  aads keywords list --campaign-id 123 --adgroup-id 456 --filter "status IN [ACTIVE, PAUSED]"
    --sort "text:asc"
```

```
FLAGS
  -adgroup-id string   Ad Group ID (or - to read IDs from stdin) (required)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -filter value        Filter: "field=value" or "field OPERATOR value" (repeatable)
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -limit 0             Maximum results; 0 fetches all pages
  -offset 0            Starting offset
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
  -selector string     Selector input: inline JSON, @file.json, or @- for stdin
  -sort value          Sort: "field:asc" or "field:desc" (repeatable)
```

### keywords get

Get a targeting keyword.

```
USAGE
  aads keywords get --campaign-id CID --adgroup-id AGID --keyword-id ID
```

```
FLAGS
  -adgroup-id string   Ad Group ID (or - to read IDs from stdin) (required)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -keyword-id string   Keyword ID (or - to read IDs from stdin) (required)
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
```

### keywords create

Create targeting keywords.

```
USAGE
  aads keywords create --campaign-id CID --adgroup-id AGID --text TEXT [flags]

Create targeting keywords using shortcut flags or --from-json for full JSON.

With --text, creates one keyword per comma-separated item. All share
the same --match-type, --bid, and --status. Quote items with commas.

JSON keys per keyword (for --from-json):
  text       string  (required) Keyword text
  matchType  string  BROAD | EXACT
  status     string  ACTIVE (default) | PAUSED
  bidAmount  Money   Keyword-level bid override

Examples:
  aads keywords create --campaign-id 101 --adgroup-id 201 --text "fitness coach"
  aads keywords create --campaign-id 101 --adgroup-id 201 --text "fitness coach,home workout planner" --match-type BROAD
  aads keywords create --campaign-id 101 --adgroup-id 201 --text "fitness coach" --bid 1.50
  aads keywords create --campaign-id 1 --adgroup-id 2 --from-json keywords.json
```

```
FLAGS
  -adgroup-id string   Ad Group ID (or - to read IDs from stdin) (required)
  -bid string          Bid amount (AMOUNT or "AMOUNT CURRENCY"; bare amount uses default currency)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -check=false         Validate and summarize without sending the request
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -force=false         Skip safety checks
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string    JSON body input: inline JSON, @file.json, or @- for stdin
  -match-type EXACT    BROAD | EXACT (default)
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
  -status string       ACTIVE (default) | PAUSED (also 1/0, enable/pause)
  -text string         Keyword text (comma-separated for multiple; quote individual keywords)
```

### keywords update

Update a targeting keyword.

```
USAGE
  aads keywords update --campaign-id CID --adgroup-id AGID --keyword-id ID [flags]

Update a targeting keyword. The API accepts partial updates.

Use shortcut flags for common changes, or --from-json for arbitrary JSON.
Shortcut flags can be combined with each other.

The --bid flag also accepts math expressions:
  Delta:       +1, -0.50, "+1 USD" (adjusts current value)
  Multiplier:  x1.1, x0.9 (multiplies current value)
  Percent:     +10%, -15% (percent adjustment)
Relative expressions fetch the current keyword from the API first.

With --from-json, the body is a JSON array of keyword update objects.

JSON keys per keyword (all optional except id):
  id         integer  (required) Keyword ID to update
  text       string   Keyword text
  matchType  string   BROAD | EXACT
  status     string   ACTIVE | PAUSED
  bidAmount  Money    Keyword-level bid override

Money object: {"amount": "1.50", "currency": "USD"}

Examples:
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --bid 1.50
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --bid "1.50 EUR"
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --bid +0.25
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --bid x1.1
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --bid -10%
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --status 0
  aads keywords update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --from-json updates.json
```

```
FLAGS
  -adgroup-id string   Ad Group ID (or - to read IDs from stdin) (required)
  -bid string          Keyword bid (AMOUNT or "AMOUNT CURRENCY"; bare amount uses default currency), delta (+1), multiplier (x1.1), or percent (+10%)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -check=false         Validate and summarize without sending the request
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -force=false         Skip safety checks
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string    JSON body input: inline JSON, @file.json, or @- for stdin
  -keyword-id string   Keyword ID (or - to read IDs from stdin) (required)
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
  -status string       ACTIVE | PAUSED (also 1/0, enable/pause)
```

### keywords delete

Delete targeting keywords.

```
USAGE
  aads keywords delete --campaign-id CID --adgroup-id AGID (--keyword-id ID | --from-json FILE) --confirm

Delete targeting keywords by ID or from a JSON array of keyword IDs.

Use --keyword-id for one ID or a comma-separated list of IDs.
Use --from-json for inline JSON, @file.json, or @- for stdin. The body is a JSON array of keyword IDs to delete.

Requires --confirm to execute. Without --confirm, the command prints a check summary and exits with an error.
```

```
FLAGS
  -adgroup-id string   Ad Group ID (or - to read IDs from stdin) (required)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -check=false         Validate and summarize without sending the request
  -confirm=false       Confirm deletion
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string    JSON body input: inline JSON, @file.json, or @- for stdin
  -keyword-id string   Keyword ID, or comma-separated keyword IDs (or - to read IDs from stdin)
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
```

---

## negatives

Manage negative keywords (campaign and ad group level).

```
USAGE
  aads negatives <subcommand>

SUBCOMMANDS
  list    List negative keywords.
  get     Get a negative keyword.
  create  Create negative keywords.
  update  Update negative keywords.
  delete  Delete negative keywords.
```

### negatives list

List negative keywords.

```
USAGE
  aads negatives list --campaign-id CID [--adgroup-id AGID] [flags]

List negative keywords, with optional filtering.

Without --adgroup-id, lists campaign-level negative keywords.
With --adgroup-id, lists ad group-level negative keywords.

Use --filter / --sort flags, or --selector for JSON selector.

Filter operators:
  EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

Searchable and filterable fields:
  id, campaignId, adGroupId, text, matchType, status

Selector JSON keys: conditions, fields, orderBy, pagination.

Examples:
  aads negatives list --campaign-id 123
  aads negatives list --campaign-id 123 --adgroup-id 456
  aads negatives list --campaign-id 123 --filter "text CONTAINS free" --sort "text:asc"
  aads negatives list --campaign-id 123 --adgroup-id 456 --filter "matchType=EXACT"
```

```
FLAGS
  -adgroup-id string   Ad Group ID; omit for campaign-level (or - to read IDs from stdin)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -filter value        Filter: "field=value" or "field OPERATOR value" (repeatable)
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -limit 0             Maximum results; 0 fetches all pages
  -offset 0            Starting offset
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
  -selector string     Selector input: inline JSON, @file.json, or @- for stdin
  -sort value          Sort: "field:asc" or "field:desc" (repeatable)
```

### negatives get

Get a negative keyword.

```
USAGE
  aads negatives get --campaign-id CID [--adgroup-id AGID] --keyword-id ID

Get a negative keyword by ID.

Without --adgroup-id, fetches a campaign-level negative keyword.
With --adgroup-id, fetches an ad group-level negative keyword.
```

```
FLAGS
  -adgroup-id string   Ad Group ID; omit for campaign-level (or - to read IDs from stdin)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -keyword-id string   Negative Keyword ID (or - to read IDs from stdin) (required)
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
```

### negatives create

Create negative keywords.

```
USAGE
  aads negatives create --campaign-id CID [--adgroup-id AGID] --text TEXT [flags]

Create negative keywords using shortcut flags or --from-json for full JSON.

Without --adgroup-id, creates campaign-level negative keywords.
With --adgroup-id, creates ad group-level negative keywords.

With --text, creates one negative keyword per comma-separated item.
All share the same --match-type and --status. Quote items with commas.

JSON keys per negative keyword (for --from-json):
  text       string  (required) Keyword text to exclude
  matchType  string  BROAD | EXACT

Examples:
  aads negatives create --campaign-id 101 --text "free workout,fitness wallpaper"
  aads negatives create --campaign-id 101 --adgroup-id 201 --text "yoga mat,protein powder" --match-type BROAD
  aads negatives create --campaign-id 1 --from-json negatives.json
```

```
FLAGS
  -adgroup-id string   Ad Group ID; omit for campaign-level (or - to read IDs from stdin)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -check=false         Validate and summarize without sending the request
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string    JSON body input: inline JSON, @file.json, or @- for stdin
  -match-type EXACT    BROAD | EXACT (default)
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
  -status string       ACTIVE (default) | PAUSED (also 1/0, enable/pause)
  -text string         Keyword text (comma-separated for multiple; quote individual keywords)
```

### negatives update

Update negative keywords.

```
USAGE
  aads negatives update --campaign-id CID [--adgroup-id AGID] --keyword-id ID [flags]

Update negative keywords.

Without --adgroup-id, updates campaign-level negative keywords.
With --adgroup-id, updates ad group-level negative keywords.

Use shortcut flags for quick status changes, or --from-json for arbitrary JSON array updates.

JSON keys per negative keyword (all optional except id):
  id         integer  (required) Negative keyword ID to update
  text       string   Keyword text
  matchType  string   BROAD | EXACT
  status     string   ACTIVE | PAUSED

Examples:
  aads negatives update --campaign-id 1 --keyword-id 2 --status 0
  aads negatives update --campaign-id 1 --adgroup-id 2 --keyword-id 3 --status PAUSED
  aads negatives update --campaign-id 1 --keyword-id 2 --from-json updates.json
```

```
FLAGS
  -adgroup-id string   Ad Group ID; omit for campaign-level (or - to read IDs from stdin)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -check=false         Validate and summarize without sending the request
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string    JSON body input: inline JSON, @file.json, or @- for stdin
  -keyword-id string   Negative Keyword ID (or - to read IDs from stdin) (required)
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
  -status string       ACTIVE | PAUSED (also 1/0, enable/pause)
```

### negatives delete

Delete negative keywords.

```
USAGE
  aads negatives delete --campaign-id CID [--adgroup-id AGID] (--keyword-id ID | --from-json FILE) --confirm

Delete negative keywords by ID or from a JSON array of keyword IDs.

Without --adgroup-id, deletes campaign-level negative keywords.
With --adgroup-id, deletes ad group-level negative keywords.
Use --keyword-id for one ID or a comma-separated list of IDs.
Use --from-json for inline JSON, @file.json, or @- for stdin. The body is a JSON array of keyword IDs to delete.

Requires --confirm to execute. Without --confirm, the command prints a check summary and exits with an error.
```

```
FLAGS
  -adgroup-id string   Ad Group ID; omit for campaign-level (or - to read IDs from stdin)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -check=false         Validate and summarize without sending the request
  -confirm=false       Confirm deletion
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string    JSON body input: inline JSON, @file.json, or @- for stdin
  -keyword-id string   Negative Keyword ID, or comma-separated negative keyword IDs (or - to read IDs from stdin)
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
```

---

## ads

Manage ads.

```
USAGE
  aads ads <subcommand>

SUBCOMMANDS
  list    List ads.
  get     Get an ad by ID.
  create  Create an ad.
  update  Update an ad.
  delete  Delete an ad.
```

### ads list

List ads.

```
USAGE
  aads ads list [--campaign-id CID --adgroup-id AGID] [flags]

List ads, with optional filtering.

With --campaign-id and --adgroup-id, lists ads in that ad group.
Without them, searches across all campaigns.

Use --filter / --sort flags, or --selector for JSON selector.

Filter operators:
  EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

Searchable and filterable fields:
  id, orgId, campaignId, adGroupId, name, status, servingStatus,
  creativeType

Selector JSON keys: conditions, fields, orderBy, pagination.

Examples:
  aads ads list
  aads ads list --campaign-id 123 --adgroup-id 456
  aads ads list --filter "status=ENABLED"
  aads ads list --campaign-id 123 --adgroup-id 456 --filter "name CONTAINS promo" --filter "creativeType=CUSTOM_PRODUCT_PAGE" --sort "name:asc"
```

```
FLAGS
  -adgroup-id string   Ad Group ID (omit both to search all) (or - to read IDs from stdin)
  -campaign-id string  Campaign ID (omit both to search all) (or - to read IDs from stdin)
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -filter value        Filter: "field=value" or "field OPERATOR value" (repeatable)
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -limit 0             Maximum results; 0 fetches all pages
  -offset 0            Starting offset
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
  -selector string     Selector input: inline JSON, @file.json, or @- for stdin
  -sort value          Sort: "field:asc" or "field:desc" (repeatable)
```

### ads get

Get an ad by ID.

```
USAGE
  aads ads get --campaign-id CID --adgroup-id AGID --ad-id ID
```

```
FLAGS
  -ad-id string        Ad ID (or - to read IDs from stdin) (required)
  -adgroup-id string   Ad Group ID (or - to read IDs from stdin) (required)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
```

### ads create

Create an ad.

```
USAGE
  aads ads create --campaign-id CID --adgroup-id AGID --name NAME --creative-id ID [flags]

Create an ad using shortcut flags or --from-json for full JSON.

JSON keys (for --from-json):
  creativeId    integer  (required) Creative ID to use for this ad
  name          string   (required) Ad name
  status        string   ENABLED (default) | PAUSED
  creativeType  string   CUSTOM_PRODUCT_PAGE | CREATIVE_SET | DEFAULT_PRODUCT_PAGE

Examples:
  aads ads create --campaign-id 1 --adgroup-id 2 --name "My Ad" --creative-id 123456
  aads ads create --campaign-id 1 --adgroup-id 2 --name "My Ad" --creative-id 123456 --status 0
  aads ads create --campaign-id 1 --adgroup-id 2 --from-json ad.json
```

```
FLAGS
  -adgroup-id string     Ad Group ID (or - to read IDs from stdin) (required)
  -campaign-id string    Campaign ID (or - to read IDs from stdin) (required)
  -check=false           Validate and summarize without sending the request
  -creative-id string    Creative ID (required)
  -f json                --format (shorthand)
  -fields string         Comma-separated output fields to include
  -format json           Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string      JSON body input: inline JSON, @file.json, or @- for stdin
  -metadata-json string  Inline creative metadata JSON
  -name string           Ad name (required)
  -pretty=false          Pretty-print JSON even when stdout is not a TTY
  -status string         ENABLED (default) | PAUSED (also 1/0, enable/pause)
```

### ads update

Update an ad.

```
USAGE
  aads ads update --campaign-id CID --adgroup-id AGID --ad-id ID [flags]

Update an ad. The API accepts partial updates.

Use shortcut flags for quick status changes, or --from-json for arbitrary JSON.

JSON keys (all optional):
  name          string   Ad name
  status        string   ENABLED | PAUSED
  creativeId    integer  Creative ID
  creativeType  string   CUSTOM_PRODUCT_PAGE | CREATIVE_SET | DEFAULT_PRODUCT_PAGE

Examples:
  aads ads update --campaign-id 1 --adgroup-id 2 --ad-id 3 --status 0
  aads ads update --campaign-id 1 --adgroup-id 2 --ad-id 3 --status PAUSED
  aads ads update --campaign-id 1 --adgroup-id 2 --ad-id 3 --from-json changes.json
```

```
FLAGS
  -ad-id string        Ad ID (or - to read IDs from stdin) (required)
  -adgroup-id string   Ad Group ID (or - to read IDs from stdin) (required)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -check=false         Validate and summarize without sending the request
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string    JSON body input: inline JSON, @file.json, or @- for stdin
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
  -status string       ENABLED | PAUSED (also 1/0, enable/pause)
```

### ads delete

Delete an ad.

```
USAGE
  aads ads delete --campaign-id CID --adgroup-id AGID --ad-id ID --confirm
```

```
FLAGS
  -ad-id string        Ad ID (or - to read IDs from stdin) (required)
  -adgroup-id string   Ad Group ID (or - to read IDs from stdin) (required)
  -campaign-id string  Campaign ID (or - to read IDs from stdin) (required)
  -check=false         Validate and summarize without sending the request
  -confirm=false       Confirm deletion
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
```

---

## creatives

Manage creatives.

```
USAGE
  aads creatives <subcommand>

SUBCOMMANDS
  list    List creatives.
  get     Get a creative by ID.
  create  Create a creative.
```

### creatives list

List creatives.

```
USAGE
  aads creatives list [flags]

List creatives, with optional filtering.

Filter examples (repeatable):
  --filter "type=CUSTOM_PRODUCT_PAGE"
  --filter "name CONTAINS holiday"
  --filter "state=VALID"
  --sort "name:asc"

Filter operators: EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

		Searchable and filterable fields:
  id, orgId, adamId, name, type, state

Advanced: use --selector for inline JSON.
```

```
FLAGS
  -f json           --format (shorthand)
  -fields string    Comma-separated output fields to include
  -filter value     Filter: "field=value" or "field OPERATOR value" (repeatable)
  -format json      Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -limit 0          Maximum results; 0 fetches all pages
  -offset 0         Starting offset
  -pretty=false     Pretty-print JSON even when stdout is not a TTY
  -selector string  Selector input: inline JSON, @file.json, or @- for stdin
  -sort value       Sort: "field:asc" or "field:desc" (repeatable)
```

### creatives get

Get a creative by ID.

```
USAGE
  aads creatives get --creative-id ID
```

```
FLAGS
  -creative-id string                         Creative ID (or - to read IDs from stdin) (required)
  -f json                                     --format (shorthand)
  -fields string                              Comma-separated output fields to include
  -format json                                Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -include-deleted-creative-set-assets=false  Include deleted creative set assets
  -pretty=false                               Pretty-print JSON even when stdout is not a TTY
```

### creatives create

Create a creative.

```
USAGE
  aads creatives create --adam-id ID --name NAME --type TYPE [flags]

Create a creative using shortcut flags or --from-json for full JSON.

JSON keys (for --from-json):
  adamId         integer  (required) App Store app ID
  name           string   (required) Creative name
  type           string   (required) CUSTOM_PRODUCT_PAGE | CREATIVE_SET | DEFAULT_PRODUCT_PAGE
  productPageId  string   Product page ID (for CUSTOM_PRODUCT_PAGE type)

Examples:
  aads creatives create --adam-id 900001 --name "FitTrack Strength Page" --type CUSTOM_PRODUCT_PAGE
  aads creatives create --adam-id 900001 --name "FitTrack Strength Page" --type CUSTOM_PRODUCT_PAGE --product-page-id cpp-fitness-strength
  aads creatives create --from-json creative.json
```

```
FLAGS
  -adam-id string          App Store app ID (required)
  -check=false             Validate and summarize without sending the request
  -f json                  --format (shorthand)
  -fields string           Comma-separated output fields to include
  -format json             Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string        JSON body input: inline JSON, @file.json, or @- for stdin
  -name string             Creative name (required)
  -pretty=false            Pretty-print JSON even when stdout is not a TTY
  -product-page-id string  Product page ID
  -type string             CUSTOM_PRODUCT_PAGE | CREATIVE_SET | DEFAULT_PRODUCT_PAGE (required)
```

---

## budgetorders

Manage budget orders.

```
USAGE
  aads budgetorders <subcommand>

SUBCOMMANDS
  list    List all budget orders.
  get     Get a budget order.
  create  Create a budget order.
  update  Update a budget order.
```

### budgetorders list

List all budget orders.

```
USAGE
  aads budgetorders list
```

```
FLAGS
  -f json         --format (shorthand)
  -fields string  Comma-separated output fields to include
  -format json    Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -limit 0        Maximum results; 0 fetches all
  -offset 0       Starting offset
  -pretty=false   Pretty-print JSON even when stdout is not a TTY
  -sort value     Sort: "field:asc" or "field:desc" (repeatable)
```

### budgetorders get

Get a budget order.

```
USAGE
  aads budgetorders get --budget-order-id ID
```

```
FLAGS
  -budget-order-id string  Budget Order ID (or - to read IDs from stdin) (required)
  -f json                  --format (shorthand)
  -fields string           Comma-separated output fields to include
  -format json             Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false            Pretty-print JSON even when stdout is not a TTY
```

### budgetorders create

Create a budget order.

```
USAGE
  aads budgetorders create --name NAME [flags]

Create a budget order using shortcut flags or --from-json for full JSON.

Shortcut flags are wrapped in {"orgIds": [ORG], "bo": {...}} using --org-id.

JSON keys (for --from-json):
  orgIds  [integer]  (required) Organization IDs to associate
  bo      object     (required) Budget order details:
    name               string  Budget order name
    budget             Money   Total budget amount
    startDate          string  Start date (YYYY-MM-DD)
    endDate            string  End date (YYYY-MM-DD)
    primaryBuyerEmail  string  Primary buyer email
    primaryBuyerName   string  Primary buyer name
    billingEmail       string  Billing contact email
    clientName         string  Client name
    orderNumber        string  Purchase order number

Money object: {"amount": "10000.00", "currency": "USD"}

Examples:
  aads budgetorders create --name "Q1 Budget" --budget-amount "10000 USD" --start-date 2025-01-01 --end-date 2025-03-31
  aads budgetorders create --name "Q1 Budget" --primary-buyer-email buyer@example.com --primary-buyer-name "Jane Doe"
  aads budgetorders create --from-json budgetorder.json
```

```
FLAGS
  -billing-email string        Billing contact email
  -budget-amount string        Total budget (AMOUNT or "AMOUNT CURRENCY"; bare amount uses default currency)
  -check=false                 Validate and summarize without sending the request
  -client-name string          Client name
  -end-date string             End date (YYYY-MM-DD, now, or signed offset like -5d)
  -f json                      --format (shorthand)
  -fields string               Comma-separated output fields to include
  -format json                 Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string            Path to JSON body file (or - for stdin)
  -name string                 Budget order name (required)
  -order-number string         Purchase order number
  -pretty=false                Pretty-print JSON even when stdout is not a TTY
  -primary-buyer-email string  Primary buyer email
  -primary-buyer-name string   Primary buyer name
  -start-date string           Start date (YYYY-MM-DD, now, or signed offset like -5d)
```

### budgetorders update

Update a budget order.

```
USAGE
  aads budgetorders update --budget-order-id ID [flags]

Update a budget order. Use shortcut flags or --from-json for the full body.

Shortcut flags are wrapped in {"orgIds": [ORG], "bo": {...}} using --org-id.

JSON keys (for --from-json):
  orgIds  [integer]  (required) Organization IDs
  bo      object     (required) Fields to update:
    name               string  Budget order name
    budget             Money   Total budget amount
    startDate          string  Start date (YYYY-MM-DD)
    endDate            string  End date (YYYY-MM-DD)
    primaryBuyerEmail  string  Primary buyer email
    primaryBuyerName   string  Primary buyer name
    billingEmail       string  Billing contact email
    clientName         string  Client name
    orderNumber        string  Purchase order number

Money object: {"amount": "10000.00", "currency": "USD"}

Examples:
  aads budgetorders update --budget-order-id 123 --name "Q2 Budget"
  aads budgetorders update --budget-order-id 123 --budget-amount "15000 USD"
  aads budgetorders update --budget-order-id 123 --end-date 2025-06-30
  aads budgetorders update --budget-order-id 123 --from-json changes.json
```

```
FLAGS
  -budget-amount string    Total budget (AMOUNT or "AMOUNT CURRENCY"; bare amount uses default currency)
  -budget-order-id string  Budget Order ID (or - to read IDs from stdin) (required)
  -check=false             Validate and summarize without sending the request
  -end-date string         End date (YYYY-MM-DD, now, or signed offset like -5d)
  -f json                  --format (shorthand)
  -fields string           Comma-separated output fields to include
  -format json             Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string        JSON body input: inline JSON, @file.json, or @- for stdin
  -name string             Budget order name
  -pretty=false            Pretty-print JSON even when stdout is not a TTY
  -start-date string       Start date (YYYY-MM-DD, now, or signed offset like -5d)
```

---

## product-pages

Manage custom product pages.

```
USAGE
  aads product-pages <subcommand>

SUBCOMMANDS
  list       List custom product pages.
  get        Get a product page.
  locales    Get product page locales.
  countries  Get supported countries/regions.
  devices    Get app preview device sizes.
```

### product-pages list

List custom product pages.

```
USAGE
  aads product-pages list --adam-id ID
```

```
FLAGS
  -adam-id string  App Adam ID (or - to read IDs from stdin) (required)
  -f json          --format (shorthand)
  -fields string   Comma-separated output fields to include
  -filter value    Filter query: "name=value" or "state=value" (repeatable)
  -format json     Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -limit 0         Maximum results; 0 fetches all
  -offset 0        Starting offset
  -pretty=false    Pretty-print JSON even when stdout is not a TTY
  -sort value      Sort: "field:asc" or "field:desc" (repeatable)
```

### product-pages get

Get a product page.

```
USAGE
  aads product-pages get --adam-id ID --product-page-id PPID
```

```
FLAGS
  -adam-id string          App Adam ID
  -f json                  --format (shorthand)
  -fields string           Comma-separated output fields to include
  -format json             Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false            Pretty-print JSON even when stdout is not a TTY
  -product-page-id string  Product Page ID
```

### product-pages locales

Get product page locales.

```
USAGE
  aads product-pages locales --adam-id ID --product-page-id PPID
```

```
FLAGS
  -adam-id string          App Adam ID (or - to read IDs from stdin) (required)
  -f json                  --format (shorthand)
  -fields string           Comma-separated output fields to include
  -format json             Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false            Pretty-print JSON even when stdout is not a TTY
  -product-page-id string  Product Page ID (or - to read IDs from stdin) (required)
  -sort value              Sort: "field:asc" or "field:desc" (repeatable)
```

### product-pages countries

Get supported countries/regions.

```
USAGE
  aads product-pages countries
```

```
FLAGS
  -countries-or-regions string  Comma-separated ISO alpha-2 country or region codes
  -f json                       --format (shorthand)
  -fields string                Comma-separated output fields to include
  -format json                  Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false                 Pretty-print JSON even when stdout is not a TTY
  -sort value                   Sort: "field:asc" or "field:desc" (repeatable)
```

### product-pages devices

Get app preview device sizes.

```
USAGE
  aads product-pages devices
```

```
FLAGS
  -f json         --format (shorthand)
  -fields string  Comma-separated output fields to include
  -format json    Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false   Pretty-print JSON even when stdout is not a TTY
  -sort value     Sort: "field:asc" or "field:desc" (repeatable)
```

---

## ad-rejections

Manage ad creative rejection reasons.

```
USAGE
  aads ad-rejections <subcommand>

SUBCOMMANDS
  list    List ad creative rejection reasons.
  get     Get rejection reasons by product page reason ID.
  assets  List app assets.
```

### ad-rejections list

List ad creative rejection reasons.

```
USAGE
  aads ad-rejections list [flags]

List ad creative rejection reasons, with optional filtering.

Filter examples (repeatable):
  --filter "adamId=900001"
  --filter "countryOrRegion=US"
  --filter "reasonLevel IN [CUSTOM_PRODUCT_PAGE, DEFAULT_PRODUCT_PAGE]"
  --sort "id:desc"

Filter operators: EQUALS, NOT_EQUALS (local), CONTAINS, STARTSWITH, ENDSWITH, IN,
  LESS_THAN, GREATER_THAN, BETWEEN, CONTAINS_ALL, CONTAINS_ANY

Searchable and filterable fields:
  id, adamId, productPageId, assetGenId, languageCode, reasonCode,
  reasonType, reasonLevel, supplySource, countryOrRegion

Advanced: use --selector for inline JSON.
```

```
FLAGS
  -f json           --format (shorthand)
  -fields string    Comma-separated output fields to include
  -filter value     Filter: "field=value" or "field OPERATOR value" (repeatable)
  -format json      Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -limit 0          Maximum results; 0 fetches all pages
  -offset 0         Starting offset
  -pretty=false     Pretty-print JSON even when stdout is not a TTY
  -selector string  Selector input: inline JSON, @file.json, or @- for stdin
  -sort value       Sort: "field:asc" or "field:desc" (repeatable)
```

### ad-rejections get

Get rejection reasons by product page reason ID.

```
USAGE
  aads ad-rejections get --reason-id ID
```

```
FLAGS
  -f json            --format (shorthand)
  -fields string     Comma-separated output fields to include
  -format json       Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false      Pretty-print JSON even when stdout is not a TTY
  -reason-id string  Product Page Reason ID (or - to read IDs from stdin) (required)
```

### ad-rejections assets

List app assets.

```
USAGE
  aads ad-rejections assets --adam-id ID
```

```
FLAGS
  -adam-id string  App Adam ID
  -f json          --format (shorthand)
  -fields string   Comma-separated output fields to include
  -format json     Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false    Pretty-print JSON even when stdout is not a TTY
```

---

## reports

Generate reports.

```
USAGE
  aads reports <subcommand>

SUBCOMMANDS
  campaigns    Campaign-level reports.
  adgroups     Ad group-level reports.
  keywords     Keyword-level reports.
  searchterms  Search term-level reports.
  ads          Ad-level reports.
```

### reports campaigns

Campaign-level reports.

```
USAGE
  aads reports campaigns --start DATE --end DATE --sort FIELD:ORDER

Sortable fields:
  impressions, taps, localSpend, tapInstalls, tapNewDownloads,
  tapRedownloads, tapInstallCPI, avgCPT, avgCPM, ttr,
  tapInstallRate, totalInstalls, totalNewDownloads, totalRedownloads,
  totalAvgCPI, totalInstallRate

Additional filterable fields:
  campaignId, orgId, campaignName, campaignStatus, servingStatus,
  displayStatus, adChannelType, deleted, appAdamId, appAppName,
  countriesOrRegions, deviceClass, gender, ageRange, countryOrRegion,
  adminArea, locality, dailyBudget.amount, dailyBudget.currency,
  totalBudget.amount, totalBudget.currency

Local filter examples (repeatable, combined with AND):
  --filter "campaignId=123"
  --filter "localSpend > 10"
  --filter "localSpend BETWEEN [10, 50]"
```

```
FLAGS
  -condition string       Filter condition: field=operator=value
  -end string             End date (YYYY-MM-DD, now, or signed offset like -5d; profile timezone ignored unless effective report timezone is UTC) (required)
  -f json                 --format (shorthand)
  -fields string          Comma-separated output fields to include
  -filter value           Local post-fetch filter: "field=value" or "field OPERATOR value" (repeatable)
  -format json            Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -grand-totals=true      Include grand totals
  -granularity string     HOURLY | DAILY | WEEKLY | MONTHLY
  -group-by string        Group by dimension
  -no-metrics=true        Include records with no metrics
  -pretty=false           Pretty-print JSON even when stdout is not a TTY
  -row-totals=true        Include row totals
  -selector string        Selector input: inline JSON, @file.json, or @- for stdin; overrides flags
  -sort impressions:desc  Sort field:order, e.g. impressions:desc (default impressions:desc)
  -start string           Start date (YYYY-MM-DD, now, or signed offset like -5d; profile timezone ignored unless effective report timezone is UTC) (required)
  -timezone UTC           UTC (default) | ORTZ
```

### reports adgroups

Ad group-level reports.

```
USAGE
  aads reports adgroups --campaign-id ID [--adgroup-id AGID] --start DATE --end DATE --sort FIELD:ORDER

Sortable fields:
  impressions, taps, localSpend, tapInstalls, tapNewDownloads,
  tapRedownloads, tapInstallCPI, avgCPT, avgCPM, ttr,
  tapInstallRate, totalInstalls, totalNewDownloads, totalRedownloads,
  totalAvgCPI, totalInstallRate

Additional filterable fields:
  campaignId, adGroupId, orgId, adGroupName, adGroupStatus,
  adGroupServingStatus, adGroupDisplayStatus, deleted,
  automatedKeywordsOptIn, defaultBidAmount.amount, defaultBidAmount.currency,
  cpaGoal.amount, cpaGoal.currency, startTime, endTime, modificationTime,
  deviceClass, gender, ageRange, countryOrRegion, adminArea, locality

Local filter examples (repeatable, combined with AND):
  --filter "adGroupId=5001"
  --filter "localSpend > 10"
  --filter "adGroupName STARTSWITH Brand"

Selector shortcut:
  --adgroup-id 5001
    Adds selector condition: adGroupId EQUALS 5001
    Supports --adgroup-id - for stdin pipelines.
```

```
FLAGS
  -adgroup-id string      Ad Group ID filter shortcut (or - to read IDs from stdin)
  -campaign-id string     Campaign ID (or - to read IDs from stdin) (required)
  -condition string       Filter condition: field=operator=value
  -end string             End date (YYYY-MM-DD, now, or signed offset like -5d; profile timezone ignored unless effective report timezone is UTC) (required)
  -f json                 --format (shorthand)
  -fields string          Comma-separated output fields to include
  -filter value           Local post-fetch filter: "field=value" or "field OPERATOR value" (repeatable)
  -format json            Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -grand-totals=true      Include grand totals
  -granularity string     HOURLY | DAILY | WEEKLY | MONTHLY
  -group-by string        Group by dimension
  -no-metrics=true        Include records with no metrics
  -pretty=false           Pretty-print JSON even when stdout is not a TTY
  -row-totals=true        Include row totals
  -selector string        Selector input: inline JSON, @file.json, or @- for stdin; overrides flags
  -sort impressions:desc  Sort field:order, e.g. impressions:desc (default impressions:desc)
  -start string           Start date (YYYY-MM-DD, now, or signed offset like -5d; profile timezone ignored unless effective report timezone is UTC) (required)
  -timezone UTC           UTC (default) | ORTZ
```

### reports keywords

Keyword-level reports.

```
USAGE
  aads reports keywords --campaign-id ID [--adgroup-id AGID] --start DATE --end DATE --sort FIELD:ORDER

Sortable fields:
  impressions, taps, localSpend, tapInstalls, tapNewDownloads,
  tapRedownloads, tapInstallCPI, avgCPT, avgCPM, ttr,
  tapInstallRate, totalInstalls, totalNewDownloads, totalRedownloads,
  totalAvgCPI, totalInstallRate

Additional filterable fields:
  campaignId, adGroupId, keywordId, orgId, adGroupName, keyword,
  keywordStatus, keywordDisplayStatus, matchType, deleted, adGroupDeleted,
  bidAmount.amount, bidAmount.currency, modificationTime,
  deviceClass, gender, ageRange, countryOrRegion, adminArea, locality

Selector shortcut:
  --adgroup-id 5001
    Adds selector condition: adGroupId EQUALS 5001
    Supports --adgroup-id - for stdin pipelines.
```

```
FLAGS
  -adgroup-id string      Ad Group ID (optional, changes report scope) (or - to read IDs from stdin)
  -campaign-id string     Campaign ID (or - to read IDs from stdin) (required)
  -condition string       Filter condition: field=operator=value
  -end string             End date (YYYY-MM-DD, now, or signed offset like -5d; profile timezone ignored unless effective report timezone is UTC) (required)
  -f json                 --format (shorthand)
  -fields string          Comma-separated output fields to include
  -filter value           Local post-fetch filter: "field=value" or "field OPERATOR value" (repeatable)
  -format json            Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -grand-totals=true      Include grand totals
  -granularity string     HOURLY | DAILY | WEEKLY | MONTHLY
  -group-by string        Group by dimension
  -no-metrics=false       Include records with no metrics
  -pretty=false           Pretty-print JSON even when stdout is not a TTY
  -row-totals=true        Include row totals
  -selector string        Selector input: inline JSON, @file.json, or @- for stdin; overrides flags
  -sort impressions:desc  Sort field:order, e.g. impressions:desc (default impressions:desc)
  -start string           Start date (YYYY-MM-DD, now, or signed offset like -5d; profile timezone ignored unless effective report timezone is UTC) (required)
  -timezone UTC           UTC (default) | ORTZ
```

### reports searchterms

Search term-level reports.

```
USAGE
  aads reports searchterms --campaign-id ID [--adgroup-id AGID] --start DATE --end DATE --sort FIELD:ORDER

Sortable fields:
  impressions, taps, localSpend, tapInstalls, tapNewDownloads,
  tapRedownloads, tapInstallCPI, avgCPT, avgCPM, ttr,
  tapInstallRate, totalInstalls, totalNewDownloads, totalRedownloads,
  totalAvgCPI, totalInstallRate

Additional filterable fields:
  campaignId, adGroupId, keywordId, orgId, adGroupName, keyword,
  keywordStatus, keywordDisplayStatus, matchType, deleted, adGroupDeleted,
  bidAmount.amount, bidAmount.currency, searchTermSource, searchTermText,
  modificationTime, deviceClass, gender, ageRange, countryOrRegion,
  adminArea, locality

Selector shortcut:
  --adgroup-id 5001
    Adds selector condition: adGroupId EQUALS 5001
    Supports --adgroup-id - for stdin pipelines.
```

```
FLAGS
  -adgroup-id string      Ad Group ID (optional, changes report scope) (or - to read IDs from stdin)
  -campaign-id string     Campaign ID (or - to read IDs from stdin) (required)
  -condition string       Filter condition: field=operator=value
  -end string             End date (YYYY-MM-DD, now, or signed offset like -5d; profile timezone ignored unless effective report timezone is UTC) (required)
  -f json                 --format (shorthand)
  -fields string          Comma-separated output fields to include
  -filter value           Local post-fetch filter: "field=value" or "field OPERATOR value" (repeatable)
  -format json            Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -grand-totals=true      Include grand totals
  -granularity string     HOURLY | DAILY | WEEKLY | MONTHLY
  -group-by string        Group by dimension
  -no-metrics=false       Include records with no metrics
  -pretty=false           Pretty-print JSON even when stdout is not a TTY
  -row-totals=true        Include row totals
  -selector string        Selector input: inline JSON, @file.json, or @- for stdin; overrides flags
  -sort impressions:desc  Sort field:order, e.g. impressions:desc (default impressions:desc)
  -start string           Start date (YYYY-MM-DD, now, or signed offset like -5d; profile timezone ignored unless effective report timezone is UTC) (required)
  -timezone UTC           UTC (default) | ORTZ
```

### reports ads

Ad-level reports.

```
USAGE
  aads reports ads --campaign-id ID --start DATE --end DATE --sort FIELD:ORDER

Sortable fields:
  impressions, taps, localSpend, tapInstalls, tapNewDownloads,
  tapRedownloads, tapInstallCPI, avgCPT, avgCPM, ttr,
  tapInstallRate, totalInstalls, totalNewDownloads, totalRedownloads,
  totalAvgCPI, totalInstallRate

Additional filterable fields:
  campaignId, adGroupId, adId, creativeId, orgId, productPageId,
  adName, creativeType, status, adDisplayStatus, adServingStateReasons,
  language, deleted, creationTime, modificationTime, deviceClass,
  gender, ageRange, countryOrRegion, adminArea, locality
```

```
FLAGS
  -campaign-id string     Campaign ID (or - to read IDs from stdin) (required)
  -condition string       Filter condition: field=operator=value
  -end string             End date (YYYY-MM-DD, now, or signed offset like -5d; profile timezone ignored unless effective report timezone is UTC) (required)
  -f json                 --format (shorthand)
  -fields string          Comma-separated output fields to include
  -filter value           Local post-fetch filter: "field=value" or "field OPERATOR value" (repeatable)
  -format json            Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -grand-totals=true      Include grand totals
  -granularity string     HOURLY | DAILY | WEEKLY | MONTHLY
  -group-by string        Group by dimension
  -no-metrics=false       Include records with no metrics
  -pretty=false           Pretty-print JSON even when stdout is not a TTY
  -row-totals=true        Include row totals
  -selector string        Selector input: inline JSON, @file.json, or @- for stdin; overrides flags
  -sort impressions:desc  Sort field:order, e.g. impressions:desc (default impressions:desc)
  -start string           Start date (YYYY-MM-DD, now, or signed offset like -5d; profile timezone ignored unless effective report timezone is UTC) (required)
  -timezone UTC           UTC (default) | ORTZ
```

---

## impression-share

Manage impression share reports.

```
USAGE
  aads impression-share <subcommand>

SUBCOMMANDS
  create  Create an impression share report.
  get     Get a single impression share report.
  list    List all impression share reports.
```

### impression-share create

Create an impression share report.

```
USAGE
  aads impression-share create --name NAME [flags]

Use shortcut flags or --from-json for the full JSON body.

Conditional shortcut flags:
  --startTime and --endTime are required when --dateRange is CUSTOM
  providing only one of --startTime/--endTime is allowed; the other default still applies
  --dateRange is not allowed with DAILY granularity

Shortcut flags:
  --name         Report name (required)
  --dateRange    LAST_WEEK | LAST_2_WEEKS | LAST_4_WEEKS | CUSTOM
  --startTime    Start date (required when --dateRange is CUSTOM; default -7d when no date flags are provided)
  --endTime      End date (required when --dateRange is CUSTOM; default now when no date flags are provided)
  --granularity  DAILY (default) | WEEKLY

JSON keys (for --from-json):
  name         string    (required) Report name
  dateRange    string    LAST_WEEK | LAST_2_WEEKS | LAST_4_WEEKS | CUSTOM
  startTime    string    Start date (required if dateRange is CUSTOM)
  endTime      string    End date (required if dateRange is CUSTOM)
  granularity  string    DAILY | WEEKLY
  selector     object    Filter conditions:
    conditions  [object]  Each: {field, operator, values}
                          Supported operator: IN
                          Supported fields: adamId, countryOrRegion

Report metrics returned: lowImpressionShare, highImpressionShare, rank, searchPopularity.
Report dimensions: adamId, appName, countryOrRegion, searchTerm.

Examples:
  aads impression-share create --name "Weekly Share Report"
  aads impression-share create --name "Weekly Share Report" --granularity WEEKLY --dateRange LAST_WEEK
  aads impression-share create --name "Custom Share Report" --dateRange CUSTOM --startTime 2026-03-01 --endTime 2026-03-07
  aads impression-share create --from-json '{"name":"Weekly Share Report","dateRange":"LAST_WEEK","granularity":"WEEKLY"}'
```

```
FLAGS
  -check=false         Validate and summarize without sending the request
  -dateRange string    LAST_WEEK | LAST_2_WEEKS | LAST_4_WEEKS | CUSTOM
  -endTime string      End date (YYYY-MM-DD, now, or signed offset like -5d) (required when --dateRange=CUSTOM)
  -f json              --format (shorthand)
  -fields string       Comma-separated output fields to include
  -format json         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string    JSON body input: inline JSON, @file.json, or @- for stdin
  -granularity string  DAILY (default) | WEEKLY
  -name string         Report name (required)
  -pretty=false        Pretty-print JSON even when stdout is not a TTY
  -startTime string    Start date (YYYY-MM-DD, now, or signed offset like -5d) (required when --dateRange=CUSTOM)
```

### impression-share get

Get a single impression share report.

```
USAGE
  aads impression-share get --report-id ID [--download FILE]
```

```
FLAGS
  -download string   Write the file at downloadUri to this path; use - for stdout
  -f json            --format (shorthand)
  -fields string     Comma-separated output fields to include
  -format json       Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false      Pretty-print JSON even when stdout is not a TTY
  -report-id string  Custom Report ID
```

### impression-share list

List all impression share reports.

```
USAGE
  aads impression-share list
```

```
FLAGS
  -f json         --format (shorthand)
  -fields string  Comma-separated output fields to include
  -format json    Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -limit 0        Maximum results; 0 fetches all
  -offset 0       Starting offset
  -pretty=false   Pretty-print JSON even when stdout is not a TTY
  -sort string    Sort query: "field:asc" or "field:desc"
```

---

## apps

Search and inspect apps.

```
USAGE
  aads apps <subcommand>

SUBCOMMANDS
  search       Search for iOS apps.
  eligibility  Check app eligibility.
  details      Get app details.
  localized    Get localized app details.
```

### apps search

Search for iOS apps.

```
USAGE
  aads apps search [--query TEXT] [--only-owned-apps] [flags]

Search for iOS apps.

Requires at least one of --query or --only-owned-apps.
```

```
FLAGS
  -f json                 --format (shorthand)
  -fields string          Comma-separated output fields to include
  -format json            Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -limit 0                Maximum results; 0 fetches all pages
  -offset 0               Starting offset
  -only-owned-apps=false  Only return apps owned by the current organization
  -pretty=false           Pretty-print JSON even when stdout is not a TTY
  -query string           Search query string
  -sort value             Sort: "field:asc" or "field:desc" (repeatable)
```

### apps eligibility

Check app eligibility.

```
USAGE
  aads apps eligibility --adam-id ID --country-or-region CC [flags]

Use shortcut flags or --from-json for the selector body.

Selector JSON keys (for --from-json):
  conditions       [object] Each: {field, operator, values}
  fields           [string] Fields to return
  orderBy          [object] Each: {field, sortOrder}
  pagination       object   {offset, limit}

Shortcut flags map to selector conditions:
  --country-or-region -> countryOrRegion EQUALS <value>
  --device-class      -> deviceClass EQUALS <value>
  --supply-source     -> supplySource EQUALS <value>
  --min-age           -> minAge EQUALS <value>

Alternate input:
  --from-json also accepts an alternate body shape with keys:
  adamId, countryOrRegion, deviceClass, supplySource, minAge

Response includes eligibility state: ELIGIBLE or INELIGIBLE.

Examples:
  aads apps eligibility --adam-id 900001 --country-or-region US
  aads apps eligibility --adam-id 900001 --country-or-region US --device-class IPHONE --supply-source APPSTORE_SEARCH_RESULTS
  aads apps eligibility --adam-id 900001 --from-json selector.json
```

```
FLAGS
  -adam-id string            App Store app ID (required)
  -check=false               Validate and summarize without sending the request
  -country-or-region string  ISO alpha-2 country code, e.g. US (required)
  -device-class string       IPHONE | IPAD
  -f json                    --format (shorthand)
  -fields string             Comma-separated output fields to include
  -format json               Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-json string          JSON body input: inline JSON, @file.json, or @- for stdin
  -min-age 0                 Minimum age restriction
  -pretty=false              Pretty-print JSON even when stdout is not a TTY
  -supply-source string      APPSTORE_SEARCH_RESULTS | APPSTORE_SEARCH_TAB | APPSTORE_PRODUCT_PAGES_BROWSE | APPSTORE_TODAY_TAB
```

### apps details

Get app details.

```
USAGE
  aads apps details --adam-id ID
```

```
FLAGS
  -adam-id string  App Adam ID (or - to read IDs from stdin) (required)
  -f json          --format (shorthand)
  -fields string   Comma-separated output fields to include
  -format json     Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false    Pretty-print JSON even when stdout is not a TTY
```

### apps localized

Get localized app details.

```
USAGE
  aads apps localized --adam-id ID
```

```
FLAGS
  -adam-id string  App Adam ID (or - to read IDs from stdin) (required)
  -f json          --format (shorthand)
  -fields string   Comma-separated output fields to include
  -format json     Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false    Pretty-print JSON even when stdout is not a TTY
```

---

## geo

Search geolocations.

```
USAGE
  aads geo <subcommand>

SUBCOMMANDS
  search  Search for geolocations.
  get     Get geolocation details.
```

### geo search

Search for geolocations.

```
USAGE
  aads geo search --query TEXT [--entity TYPE] [--country-code CC] [flags]
```

```
FLAGS
  -country-code string  Country code, ISO 3166-1 alpha-2
  -entity string        Entity type: Country, AdminArea, Locality
  -f json               --format (shorthand)
  -fields string        Comma-separated output fields to include
  -format json          Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -limit 0              Maximum results; 0 fetches all pages
  -offset 0             Starting offset
  -pretty=false         Pretty-print JSON even when stdout is not a TTY
  -query string         Search query (required)
  -sort value           Sort: "field:asc" or "field:desc" (repeatable)
```

### geo get

Get geolocation details.

```
USAGE
  aads geo get --entity TYPE --geo-id ID [flags]

Get geolocation details for a specific geo identifier.

Required flags:
  --entity  Country | AdminArea | Locality
  --geo-id  Geo identifier

Examples:
  aads geo get --entity Country --geo-id US
  aads geo get --entity AdminArea --geo-id US|CA
  aads geo get --entity Locality --geo-id US|CA|San Francisco
```

```
FLAGS
  -entity string  Entity type: Country, AdminArea, Locality (required)
  -f json         --format (shorthand)
  -fields string  Comma-separated output fields to include
  -format json    Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -geo-id string  Geo identifier (required)
  -limit 0        Maximum results
  -offset 0       Starting offset
  -pretty=false   Pretty-print JSON even when stdout is not a TTY
```

---

## orgs

Manage organizations and user context (Apple Ads ACLs).

```
USAGE
  aads orgs <subcommand>

SUBCOMMANDS
  list  List organizations (Apple Ads ACLs).
  user  Get current user details from Apple Ads ACL context.
```

### orgs list

List organizations (Apple Ads ACLs).

```
USAGE
  aads orgs list
```

```
FLAGS
  -f json         --format (shorthand)
  -fields string  Comma-separated output fields to include
  -format json    Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false   Pretty-print JSON even when stdout is not a TTY
  -sort value     Sort: "field:asc" or "field:desc" (repeatable)
```

### orgs user

Get current user details from Apple Ads ACL context.

```
USAGE
  aads orgs user
```

```
FLAGS
  -f json         --format (shorthand)
  -fields string  Comma-separated output fields to include
  -format json    Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -pretty=false   Pretty-print JSON even when stdout is not a TTY
```

---

## structure

Export and import campaign/ad group structures as JSON.

```
USAGE
  aads structure <subcommand>

SUBCOMMANDS
  export  Export a structure JSON for campaigns/ad groups, keywords, and negatives.
  import  Import campaigns/ad groups, keywords, and negatives from structure JSON.
```

### structure export

Export a structure JSON for campaigns/ad groups, keywords, and negatives.

```
USAGE
  aads structure export --scope campaigns|adgroups [flags]

Export a structure JSON that can later be consumed by "aads structure import".

Scopes:
  campaigns  Export campaigns, campaign negatives, ad groups, ad group negatives, and keywords.
  adgroups   Export ad groups, ad group negatives, and keywords from matching campaigns.

Selection:
  --campaign-id selects one campaign directly and supports - for stdin pipelines.
  --campaigns-filter / --campaigns-selector choose which campaigns to include.
  Without campaign filters/selectors, all campaigns are included.
  --adgroups-filter / --adgroups-selector choose which ad groups to include within the selected campaigns.
  Without ad group filters/selectors, all ad groups in the selected campaigns are included.
  --keywords-filter / --keywords-selector choose which keywords to include within the selected ad groups.
  Without keyword filters/selectors, all keywords in the selected ad groups are included.

Field export:
  Omit a --*-fields flag to export the normalized creation-oriented field set.
  Pass --*-fields "" to export all fields for that entity type with no default-value omission.
  Pass --*-fields "fieldA,fieldB" to export those fields plus required creation fields.

Example:
  aads structure export --scope campaigns --campaigns-filter "name STARTSWITH FitTrack"

This command always writes JSON to stdout.
```

```
FLAGS
  -adgroups-fields string             Ad group fields to export. Omit for normalized defaults; use "" to export all fields.
  -adgroups-filter value              Filter: "field=value" or "field OPERATOR value" (repeatable)
  -adgroups-negatives-fields string   Ad group negative keyword fields to export. Omit for normalized defaults; use "" to export all fields.
  -adgroups-selector string           Selector input: inline JSON, @file.json, or @- for stdin
  -adgroups-sort value                Sort: "field:asc" or "field:desc" (repeatable)
  -campaign-id string                 Campaign ID (or - to read IDs from stdin)
  -campaigns-fields string            Campaign fields to export. Omit for normalized defaults; use "" to export all fields.
  -campaigns-filter value             Filter: "field=value" or "field OPERATOR value" (repeatable)
  -campaigns-negatives-fields string  Campaign negative keyword fields to export. Omit for normalized defaults; use "" to export all fields.
  -campaigns-selector string          Selector input: inline JSON, @file.json, or @- for stdin
  -campaigns-sort value               Sort: "field:asc" or "field:desc" (repeatable)
  -keywords-fields string             Keyword fields to export. Omit for normalized defaults; use "" to export all fields.
  -keywords-filter value              Filter: "field=value" or "field OPERATOR value" (repeatable)
  -keywords-selector string           Selector input: inline JSON, @file.json, or @- for stdin
  -keywords-sort value                Sort: "field:asc" or "field:desc" (repeatable)
  -no-adam-id=false                   Omit campaign adamId from exported structure JSON
  -no-budgets=false                   Omit budget, bid, CPA, and invoice-related fields unless explicitly requested
  -no-keywords=false                  Skip keyword export
  -no-negatives=false                 Skip campaign and ad group negative keyword export
  -no-times=false                     Omit campaign/ad group startTime and endTime unless explicitly requested
  -pretty=false                       Pretty-print JSON even when stdout is not a TTY
  -redact-names=false                 Redact campaign/ad group names using %(appName), %(appNameShort), %(countriesOrRegions), and %(campaignName)
  -scope string                       Structure scope: campaigns | adgroups
  -shareable=false                    Export a shareable structure preset: omits keywords, negatives, adamId, and times, and redacts names
```

### structure import

Import campaigns/ad groups, keywords, and negatives from structure JSON.

```
USAGE
  aads structure import --from-structure JSON [flags]

Import a structure JSON previously produced by "aads structure export".

Accepted structure:
  schemaVersion  integer  currently 1
  type           string   must be "structure"
  scope          string   "campaigns" or "adgroups"

Output:
  This command emits mapping JSON with type "mapping".
  Without --output-mapping, the mapping is written to stdout.
  With --output-mapping FILE, the mapping is written only to FILE.
  Use --output-mapping - to explicitly target stdout.

Validation:
  --check runs the same planning and validation path as a live import, including
  collision checks, but uses mock sequential IDs instead of sending mutating API requests.

Skip flags:
  --no-adgroups skips ad group creation for campaigns scope imports. This also skips
  ad group negatives and keywords. Campaign negatives are still created unless
  --no-negatives is also set.

Example:
  aads structure import --from-structure @structure.json --campaign-id 500 --adgroups-name "%(name) Copy" --check
```

```
FLAGS
  -ad-channel-type string              Override campaign adChannelType
  -adam-id string                      Override adamId for created campaigns
  -adgroups-end-time string            Override ad group endTime
  -adgroups-from-json string           Ad group override JSON: inline JSON, @file.json, or @- for stdin
  -adgroups-name string                Destination ad group name template; accepts %(fieldName), %(FIELD_NAME), %(1), %1, and %(CAMPAIGN_name)
  -adgroups-name-pattern string        Basic regexp pattern used to capture source ad group name parts
  -adgroups-start-time string          Override ad group startTime
  -adgroups-status string              Override ad group status
  -allow-unmatched-name-pattern=false  Allow name patterns that do not match the source name
  -automated-keywords-opt-in=false     Override automatedKeywordsOptIn to true for all created ad groups
  -bid string                          Override keyword bidAmount; pass "" to clear keyword-level bids
  -billing-event string                Override campaign billingEvent
  -budget-amount string                DEPRECATED: Override budgetAmount for created campaigns
  -campaign-id string                  Destination campaign ID for adgroups scope
  -campaigns-end-time string           Override campaign endTime
  -campaigns-from-json string          Campaign override JSON: inline JSON, @file.json, or @- for stdin
  -campaigns-name string               Destination campaign name template; accepts %(fieldName), %(FIELD_NAME), %(1), and %1
  -campaigns-name-pattern string       Basic regexp pattern used to capture source campaign name parts
  -campaigns-start-time string         Override campaign startTime
  -campaigns-status string             Override campaign status
  -check=false                         Validate and emit mapping JSON without sending mutating API requests
  -countries-or-regions string         Override countriesOrRegions for created campaigns
  -cpa-goal string                     Override ad group cpaGoal
  -daily-budget-amount string          Override dailyBudgetAmount for created campaigns
  -default-bid string                  Override ad group defaultBidAmount
  -f json                              --format (shorthand)
  -fields string                       Comma-separated output fields to include
  -format json                         Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -from-structure string               Structure JSON input: inline JSON, @file.json, or @- for stdin
  -keywords-status string              Override keyword status
  -loc-invoice-details string          Override locInvoiceDetails JSON: inline JSON, @file.json, or @- for stdin
  -match-type string                   Override keyword matchType
  -negatives-status string             Override negative keyword status
  -no-adgroups=false                   Skip ad group creation for campaigns scope imports
  -no-keywords=false                   Skip keyword creation
  -no-negatives=false                  Skip campaign and ad group negative keyword creation
  -output-mapping string               Write mapping JSON to this path; use - for stdout
  -pretty=false                        Pretty-print JSON even when stdout is not a TTY
  -supply-sources string               Override campaign supplySources
```

---

## profiles

Manage configuration profiles.

```
USAGE
  aads profiles <subcommand>

SUBCOMMANDS
  list         List all configured profiles.
  get          Show profile details.
  genkey       Generate a P-256 private key and print its public key.
  create       Create a new profile.
  update       Update an existing profile.
  delete       Delete a profile.
  set-default  Set the default profile.
```

### profiles list

List all configured profiles.

```
USAGE
  aads profiles list [--show-credentials]
```

```
FLAGS
  -show-credentials=false  Show client ID, team ID, key ID, and private key path
```

### profiles get

Show profile details.

```
USAGE
  aads profiles get [--name NAME] [--show-credentials | --show-key]

Show details for a profile. Without --name, shows the current default profile.

Example:
  aads profiles get
  aads profiles get --name work
  aads profiles get --name work --show-key
```

```
FLAGS
  -name string             Profile name (default current default profile)
  -show-credentials=false  Show client ID, team ID, key ID, and private key path
  -show-key=false          Print the profile public key
```

### profiles genkey

Generate a P-256 private key and print its public key.

```
USAGE
  aads profiles genkey --name NAME [--confirm]

Generate a P-256 (ES256) private key using openssl and print the
corresponding public key to stdout. The private key path is always:

  ~/.aads/keys/NAME-private-key.pem

If NAME matches an existing profile, the command updates that profile's
private_key_path after successful generation.

Example:
  aads profiles genkey --name default
  aads profiles genkey --name work --confirm
```

```
FLAGS
  -confirm=false  Overwrite the existing private key file
  -name string    Profile name or key name (required)
```

### profiles create

Create a new profile.

```
USAGE
  aads profiles create [--name NAME] [--org-id ID] [flags]

Create a new configuration profile. Use --interactive to launch a
terminal wizard that prompts for missing values, guides key setup, and
gathers Apple Ads credentials before writing the profile.

If --org-id is omitted, the CLI tries to infer it from Apple Ads by calling
orgs user and using parentOrgId. It then looks up the matching orgs list row
to infer default_currency and default_timezone unless you already provided
those flags. Apple calls this ACL data; the CLI exposes it under the orgs
command group. If the lookup fails, the CLI warns and still creates the
profile as long as org_id was resolved. If this is the first profile, it
becomes the default automatically.

Example:
  aads profiles create --interactive
  aads profiles create --name default --client-id SEARCHADS.abc --team-id SEARCHADS.abc --key-id abc --org-id 123
  aads profiles create --name work --client-id SEARCHADS.def --team-id SEARCHADS.def --key-id def --org-id 456 --private-key-path ~/.aads/keys/work-private-key.pem

Time defaults apply to mutation time flags like --start-time and --end-time.
If default_timezone is empty, the local machine timezone is used. If
default_time_of_day is empty, the current time in the selected timezone is
used. Report day flags do not use default_timezone unless the report timezone
is UTC.

Limit flags are stored in the config file as decimal text in default_currency.
They accept "AMOUNT" or "AMOUNT CURRENCY". If a currency is provided, it must
match default_currency. Use 0 or an empty value to disable a limit.
```

```
FLAGS
  -check=false                 Validate and summarize without writing config
  -client-id string            Apple Ads client ID
  -default-currency string     Default currency, e.g. USD
  -default-time-of-day string  Default time-of-day for date-only time flags: HH:MM or HH:MM:SS (default current time in the selected timezone)
  -default-timezone string     Default timezone for time flags, e.g. Europe/Luxembourg (default local machine timezone)
  -f json                      --format (shorthand)
  -fields string               Comma-separated output fields to include
  -format json                 Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -interactive=false           Prompt for missing profile fields in a terminal wizard
  -key-id string               Apple Ads key ID
  -max-bid string              Safety limit in default currency: "AMOUNT" or "AMOUNT CURRENCY" (0 or empty = disabled)
  -max-budget string           Safety limit in default currency: "AMOUNT" or "AMOUNT CURRENCY" (0 or empty = disabled)
  -max-cpa-goal string         Safety limit in default currency: "AMOUNT" or "AMOUNT CURRENCY" (0 or empty = disabled)
  -max-daily-budget string     Safety limit in default currency: "AMOUNT" or "AMOUNT CURRENCY" (0 or empty = disabled)
  -name string                 Profile name
  -org-id string               Apple Ads organization ID
  -pretty=false                Pretty-print JSON even when stdout is not a TTY
  -private-key-path string     Path to ES256 private key PEM file
  -team-id string              Apple Ads team ID
```

### profiles update

Update an existing profile.

```
USAGE
  aads profiles update --name NAME [flags]

Update fields on an existing profile. Only provided flags are changed.

Example:
  aads profiles update --name default --org-id 456
  aads profiles update --name work --max-daily-budget 500

Time defaults apply to mutation time flags like --start-time and --end-time.
If default_timezone is empty, the local machine timezone is used. If
default_time_of_day is empty, the current time in the selected timezone is
used. Report day flags do not use default_timezone unless the report timezone
is UTC.

Limit flags are stored in the config file as decimal text in default_currency.
They accept "AMOUNT" or "AMOUNT CURRENCY". If a currency is provided, it must
match default_currency. Use 0 or an empty value to disable a limit.
```

```
FLAGS
  -check=false                 Validate and summarize without writing config
  -client-id string            Apple Ads client ID
  -default-currency string     Default currency, e.g. USD
  -default-time-of-day string  Default time-of-day for date-only time flags: HH:MM or HH:MM:SS (default current time in the selected timezone)
  -default-timezone string     Default timezone for time flags, e.g. Europe/Luxembourg (default local machine timezone)
  -f json                      --format (shorthand)
  -fields string               Comma-separated output fields to include
  -format json                 Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -key-id string               Apple Ads key ID
  -max-bid string              Safety limit in default currency: "AMOUNT" or "AMOUNT CURRENCY" (0 or empty = disabled)
  -max-budget string           Safety limit in default currency: "AMOUNT" or "AMOUNT CURRENCY" (0 or empty = disabled)
  -max-cpa-goal string         Safety limit in default currency: "AMOUNT" or "AMOUNT CURRENCY" (0 or empty = disabled)
  -max-daily-budget string     Safety limit in default currency: "AMOUNT" or "AMOUNT CURRENCY" (0 or empty = disabled)
  -name string                 Profile name (required)
  -org-id string               Apple Ads organization ID
  -pretty=false                Pretty-print JSON even when stdout is not a TTY
  -private-key-path string     Path to ES256 private key PEM file
  -team-id string              Apple Ads team ID
```

### profiles delete

Delete a profile.

```
USAGE
  aads profiles delete --name NAME --confirm [--delete-private-key]

Delete a configuration profile. Requires --confirm.
If the deleted profile was the default, the default is cleared.
Use --delete-private-key to also remove the configured private key file.

Example:
  aads profiles delete --name work --confirm
  aads profiles delete --name work --confirm --delete-private-key
```

```
FLAGS
  -check=false               Validate and summarize without writing config
  -confirm=false             Confirm deletion
  -delete-private-key=false  Delete the configured private key file after deleting the profile
  -f json                    --format (shorthand)
  -fields string             Comma-separated output fields to include
  -format json               Output format: json (default when stdout is not a TTY) | table (default for TTY) | yaml | markdown | ids | pipe
  -name string               Profile name to delete (required)
  -pretty=false              Pretty-print JSON even when stdout is not a TTY
```

### profiles set-default

Set the default profile.

```
USAGE
  aads profiles set-default <name>

Set which profile is used by default when --profile is not specified.

Example:
  aads profiles set-default work
```

---

## version

Print the CLI version and target Apple Ads API version.

```
USAGE
  aads version
```

---

## schema

Query embedded API schema information.

```
USAGE
  aads schema [query] [--type TYPE] [--method METHOD]

Query the embedded API schema index.

Examples:
  aads schema campaigns           List all campaign endpoints
  aads schema --type Campaign     Show Campaign type fields
  aads schema --method post       List endpoints for a method
  aads schema keyword             Fuzzy search across endpoints and types
```

```
FLAGS
  -method string  Filter endpoints by HTTP method
  -type string    Show fields for a specific type
```

---

## completion

Generate shell completion scripts.

```
USAGE
  aads completion <bash|zsh|fish>

Supported shells: bash, zsh, fish.

To load completions:

  Bash:
    source <(aads completion bash)
    # or add to ~/.bashrc:
    eval "$(aads completion bash)"

  Zsh:
    source <(aads completion zsh)
    # or save to fpath:
    aads completion zsh > "${fpath[1]}/_aads"

  Fish:
    aads completion fish | source
    # or save permanently:
    aads completion fish > ~/.config/fish/completions/aads.fish
```

---

## help

Show help for a command.

```
USAGE
  aads help <command> [subcommand]
```
