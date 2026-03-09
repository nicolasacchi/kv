package commands

import (
	"fmt"
	"strings"
)

// buildFilter constructs a Klaviyo filter string from individual filter parts.
// Each part should be a complete filter expression like: equals(status,"Draft")
func buildFilter(parts ...string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	if len(nonEmpty) == 0 {
		return ""
	}
	if len(nonEmpty) == 1 {
		return nonEmpty[0]
	}
	return "and(" + strings.Join(nonEmpty, ",") + ")"
}

// filterEquals creates an equals filter expression.
func filterEquals(field, value string) string {
	if value == "" {
		return ""
	}
	return fmt.Sprintf("equals(%s,\"%s\")", field, value)
}

// filterGreaterOrEqual creates a greater-or-equal filter expression.
func filterGreaterOrEqual(field, value string) string {
	if value == "" {
		return ""
	}
	return fmt.Sprintf("greater-or-equal(%s,%s)", field, value)
}

// filterLessOrEqual creates a less-or-equal filter expression.
func filterLessOrEqual(field, value string) string {
	if value == "" {
		return ""
	}
	return fmt.Sprintf("less-or-equal(%s,%s)", field, value)
}

// parseTimeframe validates and returns the timeframe config.
// Either a predefined key or start+end dates, but not both.
func parseTimeframe(timeframe, start, end string) (map[string]any, error) {
	if timeframe != "" && (start != "" || end != "") {
		return nil, fmt.Errorf("use either --timeframe or --start/--end, not both")
	}
	if timeframe != "" {
		return map[string]any{"key": timeframe}, nil
	}
	if start != "" && end != "" {
		return map[string]any{
			"start": start,
			"end":   end,
		}, nil
	}
	if start != "" || end != "" {
		return nil, fmt.Errorf("both --start and --end are required for custom timeframes")
	}
	// Default timeframe
	return map[string]any{"key": "last_30_days"}, nil
}

// jsonapiBody wraps data in the JSON:API request envelope.
func jsonapiBody(resourceType string, attributes map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"type":       resourceType,
			"attributes": attributes,
		},
	}
}

// jsonapiBodyWithID wraps data in the JSON:API request envelope with an ID.
func jsonapiBodyWithID(resourceType, id string, attributes map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"type":       resourceType,
			"id":         id,
			"attributes": attributes,
		},
	}
}

// jsonapiRelationship creates a JSON:API relationship reference.
func jsonapiRelationship(resourceType, id string) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"type": resourceType,
			"id":   id,
		},
	}
}
