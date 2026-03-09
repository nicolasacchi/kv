# kv — Klaviyo CLI

Agentic-first CLI for the [Klaviyo API](https://developers.klaviyo.com/en/reference/api-overview). Single static binary, JSON output, designed for LLM tools like Claude Code and automation scripts.

## Install

```bash
# From source
go install github.com/nicolasacchi/kv/cmd/kv@latest

# Or download a release binary
# https://github.com/nicolasacchi/kv/releases
```

## Quick Start

```bash
# Configure API key
kv config add production --api-key pk_your_private_key
kv config use production

# Or use environment variable
export KLAVIYO_API_KEY=pk_your_private_key

# List campaigns
kv campaigns list --json

# Get a specific profile
kv profiles get 01ABC123

# Query metric aggregates
kv metrics aggregates METRIC_ID --measurements count --interval day --start 2024-06-01 --end 2024-06-30

# Campaign report (requires a conversion metric, e.g. Placed Order)
kv campaigns report CAMPAIGN_ID --conversion-metric-id METRIC_ID --timeframe last_30_days
```

## Authentication

API key resolution order (first non-empty wins):

1. `--api-key` flag
2. `KLAVIYO_API_KEY` env var
3. `KV_API_KEY` env var
4. `~/.config/kv/config.toml` — project from `--project` flag, then `default_project`

### Multi-project config

```toml
default_project = "production"

[projects.production]
api_key = "pk_abc123"

[projects.staging]
api_key = "pk_xyz789"
revision = "2025-01-15"
```

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--api-key` | — | Klaviyo API key |
| `--project` | — | Named project from config |
| `--json` | false | Force JSON output (auto-enabled when piped) |
| `--jq` | — | gjson path filter (implies `--json`) |
| `--revision` | `2024-10-15` | API revision header |
| `--max-results` | 0 | Cap results (0 = unlimited) |
| `--no-paginate` | false | First page only |
| `--raw` | false | Output raw JSON:API (no flattening) |
| `--verbose` | false | Request details to stderr |
| `--quiet` | false | Suppress non-error output |
| `--output-dir` | `.` | Directory for file exports |

## Output

- **TTY**: Human-readable tables (go-pretty)
- **Piped/non-TTY**: JSON automatically
- **`--json`**: Force JSON even on TTY
- **`--jq`**: Filter JSON with [gjson](https://github.com/tidwall/gjson) path syntax
- **`--raw`**: Output original JSON:API without flattening

```bash
# Get just campaign names
kv campaigns list --jq "#.name"

# Extract specific fields
kv campaigns list --jq '#.{id:id,name:name,status:status}'

# Single field from a get
kv flows get QQeuKd --jq "name"
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | API error (4xx/5xx) |
| 2 | CLI usage error |
| 3 | Auth error (401/403) |

### Error Output

Errors output structured JSON to stdout and human-readable text to stderr:

```json
{"error": "401: Missing or invalid private key.", "status": 401}
```

## Commands

### config

```bash
kv config add <name> --api-key <key> [--revision <rev>]   # Add project
kv config remove <name>                                    # Remove project
kv config list                                             # List all projects
kv config use <name>                                       # Set default
kv config current                                          # Show active project
```

### campaigns

Campaigns list defaults to `--channel email` (required by the Klaviyo API).

```bash
kv campaigns list [--status draft|scheduled|sent] [--channel email|sms]
kv campaigns get <ID>
kv campaigns create --name "My Campaign" [--channel email]
kv campaigns put <file.json>              # Update from JSON file
kv campaigns report <ID> --conversion-metric-id <METRIC_ID> \
  [--timeframe last_30_days] [--stats opens,clicks,open_rate]
```

### flows

```bash
kv flows list [--status draft|live|manual]
kv flows get <ID>
kv flows update <ID> --status live
kv flows report <ID> --conversion-metric-id <METRIC_ID> [--timeframe last_30_days]
```

### segments

```bash
kv segments list
kv segments get <ID>
kv segments report <ID> --conversion-metric-id <METRIC_ID> [--timeframe last_30_days]
```

### metrics

```bash
kv metrics list [--integration <name>]
kv metrics get <ID>
kv metrics aggregates <ID> \
  --start 2024-06-01 --end 2024-06-30 \
  [--measurements count,unique] \
  [--interval day] \
  [--by '$flow,$campaign'] \
  [--filter 'equals(some_field,"value")'] \
  [--timezone UTC]
```

### events

```bash
kv events list [--metric-id <ID>] [--profile-id <ID>] [--since ISO] [--until ISO]
kv events get <ID>
kv events create --metric-name "Custom Event" --profile-email user@example.com \
  [--properties '{"key":"val"}']
```

### profiles

```bash
kv profiles list [--email <email>] [--phone <phone>]
kv profiles get <ID>
kv profiles create --email user@example.com [--first-name Jane] [--last-name Doe] \
  [--phone +1234567890] [--properties '{"key":"val"}']
kv profiles update <ID> [--first-name Jane] [--properties '{"key":"val"}']
kv profiles suppress <ID>
```

### lists

```bash
kv lists list
kv lists get <ID>
kv lists create --name "My List"
kv lists members <ID>
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
kv tags get <ID>
kv tags create --name "Sale" [--group-id <ID>]
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

Requires Klaviyo Advanced KDP plan.

```bash
kv webhooks list
kv webhooks get <ID>
kv webhooks create --url https://example.com/hook --events placed_order,ordered_product
kv webhooks delete <ID>
```

### privacy (GDPR)

```bash
kv privacy request-deletion --email user@example.com
kv privacy status <REQUEST_ID>
```

## Report Statistics

Available statistics for `campaigns report`, `flows report`, and `segments report`:

| Stat | Description |
|------|-------------|
| `opens` | Total opens |
| `opens_unique` | Unique opens |
| `open_rate` | Open rate (0.0–1.0) |
| `clicks` | Total clicks |
| `clicks_unique` | Unique clicks |
| `click_rate` | Click rate (0.0–1.0) |
| `click_to_open_rate` | Click-to-open rate |
| `recipients` | Total recipients |
| `delivered` | Successfully delivered |
| `delivery_rate` | Delivery rate |
| `bounced` | Bounced count |
| `bounce_rate` | Bounce rate |
| `unsubscribes` | Unsubscribes |
| `unsubscribe_rate` | Unsubscribe rate |
| `spam_complaints` | Spam complaints |
| `conversion_rate` | Conversion rate |
| `conversions` | Conversion count |
| `conversion_value` | Conversion value |
| `revenue_per_recipient` | Revenue per recipient |

## JSON:API Flattening

Klaviyo uses JSON:API format. By default, `kv` flattens responses for easier consumption:

```json
// Raw JSON:API (--raw)
{"data": {"type": "flow", "id": "abc", "attributes": {"name": "Welcome", "status": "live"}}}

// Flattened (default)
{"id": "abc", "type": "flow", "name": "Welcome", "status": "live"}
```

Relationships are resolved: single relationships become `<name>_id`, arrays become `<name>_ids`.

Use `--raw` to get the original JSON:API response.

## Building

```bash
make build     # Build to ./bin/kv
make install   # Install to $GOPATH/bin
make test      # Run tests
```

## License

MIT
