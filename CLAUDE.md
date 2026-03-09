# CLAUDE.md — kv (Klaviyo CLI)

Agent reference for Claude Code. `kv` is a Go CLI covering the full Klaviyo REST API. Single binary, JSON output, automatic pagination.

## Auth Setup

```bash
# Option 1: Environment variable (recommended for agents)
export KLAVIYO_API_KEY=pk_your_private_key

# Option 2: Config file
kv config add production --api-key pk_your_private_key
kv config use production

# Option 3: Per-command flag
kv campaigns list --api-key pk_your_private_key
```

Resolution order: `--api-key` flag > `KLAVIYO_API_KEY` env > `KV_API_KEY` env > `~/.config/kv/config.toml`

## Output Contract

- Default (piped/non-TTY): **flattened JSON** — `id` and `type` merged into attributes, no nesting
- `--json`: force JSON on TTY
- `--raw`: original JSON:API format (nested `data.attributes`)
- `--jq <path>`: gjson filter (NOT jq syntax). Examples: `#.name`, `#.{id:id,name:name}`, `0.email`
- Errors: `{"error": "401: message", "status": 401}` to stdout, human text to stderr

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | API error (400, 404, 429, 5xx) |
| 2 | CLI usage error |
| 3 | Auth error (401, 403) |

## Global Flags

```
--api-key STRING      API key override
--project STRING      Named project from config
--json                Force JSON output
--jq STRING           gjson path filter (implies --json)
--revision STRING     API revision header (default: 2024-10-15)
--max-results INT     Cap total results (default: 0 = unlimited)
--no-paginate         First page only
--raw                 Skip JSON:API flattening
--verbose             Log request/response to stderr
--quiet               Suppress non-error output
--output-dir STRING   Directory for exports
```

## Gotchas

- `--jq` uses **gjson** syntax, NOT jq. Array iteration: `#.field`. Object extraction: `#.{a:a,b:b}`
- `campaigns list` defaults `--channel` to `email` (Klaviyo requires a channel filter)
- Report commands (`campaigns report`, `flows report`, `segments report`) **require** `--conversion-metric-id`
- `metrics aggregates` **requires** `--start` and `--end`
- `webhooks` endpoints require Klaviyo Advanced KDP plan (403 without it)
- Pagination is automatic. Use `--max-results N` to cap or `--no-paginate` for first page only
- Rate limits (429) are auto-retried with `Retry-After` header, up to 3 retries

## Command Reference

### config

```bash
kv config add <name> --api-key <key> [--revision <rev>]
kv config remove <name>
kv config list
kv config use <name>
kv config current
```

### campaigns

```bash
kv campaigns list [--status draft|scheduled|sent] [--channel email|sms]
# Output: [{id, type, name, status, created_at, send_time, ...}]

kv campaigns get <ID>
# Output: {id, type, name, status, ...}

kv campaigns create --name "Name" [--channel email]
kv campaigns put <file.json>

kv campaigns report <ID> --conversion-metric-id <METRIC_ID> \
  [--timeframe last_30_days] \
  [--start ISO --end ISO] \
  [--stats opens,clicks,open_rate]
# Output: {id, type, results: [{groupings: {...}, statistics: {...}}]}
```

### flows

```bash
kv flows list [--status draft|live|manual]
# Output: [{id, type, name, status, trigger_type, created, updated}]

kv flows get <ID>
kv flows update <ID> --status draft|live|manual

kv flows report <ID> --conversion-metric-id <METRIC_ID> \
  [--timeframe last_30_days] [--stats opens,clicks]
```

### segments

```bash
kv segments list
# Output: [{id, type, name, created, updated, is_active}]

kv segments get <ID>

kv segments report <ID> --conversion-metric-id <METRIC_ID> \
  [--timeframe last_30_days] [--stats opens,clicks]
```

### metrics

```bash
kv metrics list [--integration <name>]
# Output: [{id, type, name, integration: {name, category}, created}]

kv metrics get <ID>

kv metrics aggregates <METRIC_ID> \
  --start 2024-06-01 --end 2024-06-30 \
  [--measurements count,unique,sum_value] \
  [--interval day|week|month] \
  [--by '$flow,$campaign'] \
  [--filter 'equals(field,"value")'] \
  [--timezone UTC]
# Output: [{dimensions: [...], measurements: {count: [N, N, ...]}}]
```

### events

```bash
kv events list [--metric-id <ID>] [--profile-id <ID>] [--since ISO] [--until ISO]
# Output: [{id, type, datetime, event_properties: {...}}]

kv events get <ID>

kv events create --metric-name "Name" --profile-email user@example.com \
  [--properties '{"key":"val"}']
```

### profiles

```bash
kv profiles list [--email <email>] [--phone <phone>]
# Output: [{id, type, email, first_name, last_name, properties: {...}}]

kv profiles get <ID>
kv profiles create --email user@example.com [--first-name Jane] [--last-name Doe] \
  [--phone +1234567890] [--properties '{"key":"val"}']
kv profiles update <ID> [--first-name Jane] [--properties '{"key":"val"}']
kv profiles suppress <ID>
```

### lists

```bash
kv lists list
# Output: [{id, type, name, created, updated}]

kv lists get <ID>
kv lists create --name "Name"
kv lists members <LIST_ID>
kv lists add-member <LIST_ID> --profile <PROFILE_ID>
kv lists remove-member <LIST_ID> --profile <PROFILE_ID>
```

### catalog

```bash
kv catalog items list
kv catalog items get <ID>
kv catalog items create --payload item.json
kv catalog variants list <ITEM_ID>
kv catalog variants get <ID>
```

### tags

```bash
kv tags list
# Output: [{id, type, name, tag-group_id}]

kv tags get <ID>
kv tags create --name "Name" [--group-id <ID>]
kv tags assign <TAG_ID> --resource-type campaign --resource-id <ID>
kv tags remove <TAG_ID> --resource-type campaign --resource-id <ID>
```

### templates

```bash
kv templates list
kv templates get <ID>
kv templates render <ID> [--context '{"name":"Jane"}']
```

### webhooks

Requires Advanced KDP plan.

```bash
kv webhooks list
kv webhooks get <ID>
kv webhooks create --url https://example.com/hook --events event1,event2 [--secret KEY]
kv webhooks delete <ID>
```

### privacy

```bash
kv privacy request-deletion --email user@example.com [--phone +1234567890]
kv privacy status <REQUEST_ID>
```

## Common Patterns

```bash
# Find a metric ID for "Placed Order" (common conversion metric)
kv metrics list --jq '#(name=="Placed Order").id'

# Campaign performance summary
kv campaigns list --jq '#.{id:id,name:name,status:status}' --max-results 10

# Get daily order counts for last week
kv metrics aggregates METRIC_ID --measurements count --interval day \
  --start 2024-06-01 --end 2024-06-08

# Campaign report with conversion data
CONVERSION_METRIC=$(kv metrics list --jq '#(name=="Placed Order").id' | tr -d '"')
kv campaigns report CAMPAIGN_ID --conversion-metric-id $CONVERSION_METRIC \
  --stats opens,click_rate,conversion_rate,conversions

# Profile lookup by email
kv profiles list --email user@example.com --jq '0'

# Flow performance grouped by message
kv flows report FLOW_ID --conversion-metric-id METRIC_ID \
  --timeframe last_30_days --stats recipients,open_rate,click_rate
```

## Report Statistics

Available for `campaigns report`, `flows report`, `segments report`:

```
opens, opens_unique, open_rate, clicks, clicks_unique, click_rate,
click_to_open_rate, recipients, delivered, delivery_rate, bounced,
bounce_rate, bounced_or_failed, bounced_or_failed_rate, unsubscribes,
unsubscribe_rate, unsubscribe_uniques, spam_complaints, spam_complaint_rate,
conversion_rate, conversions, conversion_uniques, conversion_value,
average_order_value, revenue_per_recipient, failed, failed_rate
```

## Timeframe Keys

For report `--timeframe` flag:

```
today, yesterday, last_7_days, last_30_days, last_90_days,
last_3_months, last_12_months, last_365_days, last_year,
last_week, last_month, this_week, this_month, this_year
```

## Filter Syntax

Klaviyo filter expressions used in list commands:

```
equals(field,"value")
greater-or-equal(field,value)
less-or-equal(field,value)
and(expr1,expr2)
```

## Project Structure

```
cmd/kv/main.go              Entry point, exit code mapping
internal/client/client.go    HTTP client, retry, pagination
internal/client/jsonapi.go   JSON:API flattening
internal/client/errors.go    APIError type
internal/commands/*.go       Cobra command definitions
internal/config/config.go    TOML config, key resolution
internal/output/output.go    TTY detection, format dispatch
internal/output/table.go     go-pretty table rendering
internal/output/filter.go    gjson filtering
```

## Building & Testing

```bash
make build     # ./bin/kv
make install   # $GOPATH/bin/kv
make test      # go test ./...
```
