package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
)

type FormatFunc func(any) string

type ColumnDef struct {
	Header string
	Key    string
	Format FormatFunc
}

var commandColumns = map[string][]ColumnDef{
	"campaigns.list": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
		{Header: "STATUS", Key: "status"},
		{Header: "CHANNEL", Key: "channel"},
		{Header: "CREATED", Key: "created_at"},
	},
	"flows.list": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
		{Header: "STATUS", Key: "status"},
		{Header: "CREATED", Key: "created"},
	},
	"segments.list": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
		{Header: "CREATED", Key: "created"},
	},
	"metrics.list": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
		{Header: "INTEGRATION", Key: "integration"},
	},
	"events.list": {
		{Header: "ID", Key: "id"},
		{Header: "METRIC", Key: "metric_id"},
		{Header: "PROFILE", Key: "profile_id"},
		{Header: "DATETIME", Key: "datetime"},
	},
	"profiles.list": {
		{Header: "ID", Key: "id"},
		{Header: "EMAIL", Key: "email"},
		{Header: "FIRST NAME", Key: "first_name"},
		{Header: "LAST NAME", Key: "last_name"},
	},
	"lists.list": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
		{Header: "CREATED", Key: "created"},
	},
	"tags.list": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
	},
	"templates.list": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
		{Header: "EDITOR TYPE", Key: "editor_type"},
	},
	"webhooks.list": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
		{Header: "ENDPOINT URL", Key: "endpoint_url"},
	},
	"catalog.items.list": {
		{Header: "ID", Key: "id"},
		{Header: "TITLE", Key: "title"},
		{Header: "URL", Key: "url"},
	},
	"coupons.list": {
		{Header: "ID", Key: "id"},
		{Header: "EXTERNAL ID", Key: "external_id"},
		{Header: "DESCRIPTION", Key: "description"},
	},
	"coupons.codes.list": {
		{Header: "ID", Key: "id"},
		{Header: "UNIQUE CODE", Key: "unique_code"},
		{Header: "STATUS", Key: "status"},
		{Header: "EXPIRES AT", Key: "expires_at"},
	},
	"images.list": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
		{Header: "FORMAT", Key: "format"},
		{Header: "URL", Key: "image_url"},
	},
	"catalog.variants.list": {
		{Header: "ID", Key: "id"},
		{Header: "TITLE", Key: "title"},
		{Header: "PRICE", Key: "price"},
	},
	"config.list": {
		{Header: "NAME", Key: "name"},
		{Header: "API KEY", Key: "api_key"},
		{Header: "REVISION", Key: "revision"},
		{Header: "DEFAULT", Key: "default"},
	},
}

func printTable(command string, data json.RawMessage) error {
	columns, ok := commandColumns[command]
	if !ok {
		return fmt.Errorf("no table definition for %s", command)
	}

	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err != nil {
		var single map[string]any
		if err2 := json.Unmarshal(data, &single); err2 != nil {
			return fmt.Errorf("cannot render as table")
		}
		rows = []map[string]any{single}
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	if term.IsTerminal(int(os.Stdout.Fd())) {
		t.SetStyle(table.StyleLight)
	} else {
		t.SetStyle(table.StyleDefault)
	}

	header := make(table.Row, len(columns))
	for i, col := range columns {
		header[i] = col.Header
	}
	t.AppendHeader(header)

	for _, row := range rows {
		r := make(table.Row, len(columns))
		for i, col := range columns {
			if col.Format != nil {
				r[i] = col.Format(row[col.Key])
			} else {
				r[i] = formatValue(row[col.Key])
			}
		}
		t.AppendRow(r)
	}

	t.Render()
	return nil
}

func formatValue(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []any:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(parts, ", ")
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}
