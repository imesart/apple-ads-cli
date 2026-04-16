# API Endpoint Coverage

Tracks which Apple Ads Campaign Management API v5 endpoints are implemented.

Source artifacts: [`docs/apple_ads/openapi-latest.json`](apple_ads/openapi-latest.json) and [`docs/apple_ads/paths.txt`](apple_ads/paths.txt).
All request files live under `internal/api/requests/`.

## Summary

| Family | Documented | Covered | Additional | Notes |
|---|---|---|---|---|
| campaigns | 6 | 6 | | |
| adgroups | 7 | 7 | | |
| keywords | 7 | 7 | | |
| campaign-negatives | 6 | 6 | | |
| adgroup-negatives | 6 | 6 | | |
| ads | 7 | 7 | | campaign-scoped find path differs in implementation |
| creatives | 4 | 4 | | |
| budgetorders | 4 | 4 | | |
| product-pages | 5 | 5 | | |
| ad-rejection | 3 | 3 | | |
| apps | 2 | 2 | 2 | app detail endpoints are implemented but absent from generated docs |
| geo | 2 | 2 | | |
| reports | 7 | 7 | | keyword/searchterm reports handle both scopes |
| impression-share | 3 | 3 | | |
| orgs (Apple ACLs) | (not in generated docs) | | 2 | GET /acls, GET /me |
| **Total** | **69** | **69** | **4** | All documented endpoints are covered. |

## Detailed Coverage

### campaigns

| Method | Path | Request File |
|---|---|---|
| GET | `/campaigns` | `campaigns/list.go` |
| GET | `/campaigns/{campaignId}` | `campaigns/get.go` |
| POST | `/campaigns` | `campaigns/create.go` |
| PUT | `/campaigns/{campaignId}` | `campaigns/update.go` |
| DELETE | `/campaigns/{campaignId}` | `campaigns/delete.go` |
| POST | `/campaigns/find` | `campaigns/find.go` |

### adgroups

| Method | Path | Request File |
|---|---|---|
| GET | `/campaigns/{cId}/adgroups` | `adgroups/list.go` |
| GET | `/campaigns/{cId}/adgroups/{agId}` | `adgroups/get.go` |
| POST | `/campaigns/{cId}/adgroups` | `adgroups/create.go` |
| PUT | `/campaigns/{cId}/adgroups/{agId}` | `adgroups/update.go` |
| DELETE | `/campaigns/{cId}/adgroups/{agId}` | `adgroups/delete.go` |
| POST | `/campaigns/{cId}/adgroups/find` | `adgroups/find.go` |
| POST | `/adgroups/find` | `adgroups/find_all.go` |

### keywords

| Method | Path | Request File | Notes |
|---|---|---|---|
| GET | `.../{agId}/targetingkeywords` | `keywords/list.go` | |
| GET | `.../{agId}/targetingkeywords/{kwId}` | `keywords/get.go` | |
| POST | `.../{agId}/targetingkeywords/bulk` | `keywords/create.go` | |
| PUT | `.../{agId}/targetingkeywords/bulk` | `keywords/update.go` | |
| POST | `.../{agId}/targetingkeywords/delete/bulk` | `keywords/delete_bulk.go` | |
| DELETE | `.../{agId}/targetingkeywords/{kwId}` | `keywords/delete_one.go` | |
| POST | `/campaigns/{cId}/adgroups/targetingkeywords/find` | `keywords/find.go` | campaign-level find across ad groups |

All paths above prefixed with `/campaigns/{cId}/adgroups` unless noted.

### campaign-negatives

| Method | Path | Request File |
|---|---|---|
| GET | `/campaigns/{cId}/negativekeywords` | `negatives_campaign/list.go` |
| GET | `/campaigns/{cId}/negativekeywords/{kwId}` | `negatives_campaign/get.go` |
| POST | `/campaigns/{cId}/negativekeywords/bulk` | `negatives_campaign/create.go` |
| PUT | `/campaigns/{cId}/negativekeywords/bulk` | `negatives_campaign/update.go` |
| POST | `/campaigns/{cId}/negativekeywords/delete/bulk` | `negatives_campaign/delete.go` |
| POST | `/campaigns/{cId}/negativekeywords/find` | `negatives_campaign/find.go` |

### adgroup-negatives

| Method | Path | Request File |
|---|---|---|
| GET | `.../{agId}/negativekeywords` | `negatives_adgroup/list.go` |
| GET | `.../{agId}/negativekeywords/{kwId}` | `negatives_adgroup/get.go` |
| POST | `.../{agId}/negativekeywords/bulk` | `negatives_adgroup/create.go` |
| PUT | `.../{agId}/negativekeywords/bulk` | `negatives_adgroup/update.go` |
| POST | `.../{agId}/negativekeywords/delete/bulk` | `negatives_adgroup/delete.go` |
| POST | `/campaigns/{cId}/adgroups/negativekeywords/find` | `negatives_adgroup/find.go` |

All paths above prefixed with `/campaigns/{cId}/adgroups`.

### ads

| Method | Path | Request File | Notes |
|---|---|---|---|
| GET | `.../{agId}/ads` | `ads/list.go` | |
| GET | `.../{agId}/ads/{adId}` | `ads/get.go` | |
| POST | `.../{agId}/ads` | `ads/create.go` | |
| PUT | `.../{agId}/ads/{adId}` | `ads/update.go` | |
| DELETE | `.../{agId}/ads/{adId}` | `ads/delete.go` | |
| POST | `/campaigns/{cId}/ads/find` | `ads/find.go` | impl path: `/campaigns/{cId}/adgroups/{agId}/ads/find` (adgroup-scoped) |
| POST | `/ads/find` | `ads/find_all.go` | |

Paths prefixed with `/campaigns/{cId}/adgroups` unless noted.

### creatives

| Method | Path | Request File |
|---|---|---|
| GET | `/creatives` | `creatives/list.go` |
| GET | `/creatives/{creativeId}` | `creatives/get.go` |
| POST | `/creatives` | `creatives/create.go` |
| POST | `/creatives/find` | `creatives/find.go` |

### budgetorders

| Method | Path | Request File |
|---|---|---|
| GET | `/budgetorders` | `budgetorders/list.go` |
| GET | `/budgetorders/{boId}` | `budgetorders/get.go` |
| POST | `/budgetorders` | `budgetorders/create.go` |
| PUT | `/budgetorders/{boId}` | `budgetorders/update.go` |

### product-pages

| Method | Path | Request File |
|---|---|---|
| GET | `/apps/{adamId}/product-pages` | `product_pages/list.go` |
| GET | `/apps/{adamId}/product-pages/{ppId}` | `product_pages/get.go` |
| GET | `/apps/{adamId}/product-pages/{ppId}/locale-details` | `product_pages/locales.go` |
| GET | `/countries-or-regions` | `product_pages/countries.go` |
| GET | `/creativeappmappings/devices` | `product_pages/devices.go` |

### ad-rejection

| Method | Path | Request File |
|---|---|---|
| POST | `/apps/{adamId}/assets/find` | `ad_rejections/find_assets.go` |
| GET | `/product-page-reasons/{pprId}` | `ad_rejections/get.go` |
| POST | `/product-page-reasons/find` | `ad_rejections/find.go` |

### apps

| Method | Path | Request File | Notes |
|---|---|---|---|
| GET | `/search/apps` | `apps/search.go` | |
| POST | `/apps/{adamId}/eligibilities/find` | `apps/eligibility.go` | |
| GET | `/apps/{adamId}` | `apps/details.go` | additional; not in generated docs |
| GET | `/apps/{adamId}/locale-details` | `apps/localized.go` | additional; not in generated docs |

### geo

| Method | Path | Request File | Notes |
|---|---|---|---|
| GET | `/search/geo` | `geo/search.go` | query parameters select entity variants |
| POST | `/search/geo` | `geo/get.go` | georequest body with `entity` and `id` |

### reports

| Method | Path | Request File |
|---|---|---|
| POST | `/reports/campaigns` | `reports/campaigns.go` |
| POST | `/reports/campaigns/{cId}/adgroups` | `reports/adgroups.go` |
| POST | `/reports/campaigns/{cId}/keywords` | `reports/keywords.go` |
| POST | `/reports/campaigns/{cId}/adgroups/{agId}/keywords` | `reports/keywords.go` |
| POST | `/reports/campaigns/{cId}/searchterms` | `reports/searchterms.go` |
| POST | `/reports/campaigns/{cId}/adgroups/{agId}/searchterms` | `reports/searchterms.go` |
| POST | `/reports/campaigns/{cId}/ads` | `reports/ads.go` |

The keywords and searchterms request types select the path based on whether `AdGroupID` is set.

### impression-share

| Method | Path | Request File |
|---|---|---|
| POST | `/custom-reports` | `impression_share/create.go` |
| GET | `/custom-reports` | `impression_share/list.go` |
| GET | `/custom-reports/{reportId}` | `impression_share/get.go` |

### orgs (Apple ACLs; not in generated docs)

These endpoints are implemented in the CLI under `aads orgs ...` but are absent from the generated Apple Ads docs artifacts.

| Method | Path | Request File |
|---|---|---|
| GET | `/acls` | `acls/list.go` |
| GET | `/me` | `acls/me.go` |

## Path Discrepancies

Cases where the implementation uses a different API path than the generated docs list:

| Family | Generated Path | Implementation Path | Rationale |
|---|---|---|---|
| ads | `POST /campaigns/{cId}/ads/find` | `POST /campaigns/{cId}/adgroups/{agId}/ads/find` | Ad group-scoped find; cross-campaign find via `find_all.go` |
These may indicate the generated docs diverge from the current Apple API or from behavior observed in practice. Verify against [Apple's documentation](https://developer.apple.com/documentation/apple_ads) when in doubt.

## Updating This File

Update this file when:
- A new request file is added or removed under `internal/api/requests/`
- The OpenAPI and path list are regenerated (`make openapi` or `python3 scripts/generate_openapi.py docs/apple_ads/openapi.json --output-paths docs/apple_ads/paths.txt`)
- Apple publishes new API endpoints

To check for drift, compare the request file count against this document:
```
find internal/api/requests -name '*.go' ! -name '*_test.go' | wc -l
```
