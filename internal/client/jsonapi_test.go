package client

import (
	"encoding/json"
	"testing"
)

func TestFlattenSingleObject(t *testing.T) {
	resp := &JSONAPIResponse{
		Data: json.RawMessage(`{
			"type": "campaign",
			"id": "abc123",
			"attributes": {
				"name": "Welcome Email",
				"status": "Draft"
			}
		}`),
	}

	result := FlattenResponse(resp, false)
	var flat map[string]any
	if err := json.Unmarshal(result, &flat); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if flat["id"] != "abc123" {
		t.Errorf("expected id=abc123, got %v", flat["id"])
	}
	if flat["type"] != "campaign" {
		t.Errorf("expected type=campaign, got %v", flat["type"])
	}
	if flat["name"] != "Welcome Email" {
		t.Errorf("expected name=Welcome Email, got %v", flat["name"])
	}
	if flat["status"] != "Draft" {
		t.Errorf("expected status=Draft, got %v", flat["status"])
	}
}

func TestFlattenArray(t *testing.T) {
	resp := &JSONAPIResponse{
		Data: json.RawMessage(`[
			{"type": "campaign", "id": "1", "attributes": {"name": "A"}},
			{"type": "campaign", "id": "2", "attributes": {"name": "B"}}
		]`),
	}

	result := FlattenResponse(resp, false)
	var items []map[string]any
	if err := json.Unmarshal(result, &items); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0]["id"] != "1" {
		t.Errorf("expected first id=1, got %v", items[0]["id"])
	}
	if items[1]["name"] != "B" {
		t.Errorf("expected second name=B, got %v", items[1]["name"])
	}
}

func TestFlattenEmptyArray(t *testing.T) {
	resp := &JSONAPIResponse{
		Data: json.RawMessage(`[]`),
	}

	result := FlattenResponse(resp, false)
	if string(result) != "[]" {
		t.Errorf("expected [], got %s", result)
	}
}

func TestFlattenRawMode(t *testing.T) {
	original := json.RawMessage(`{"type":"campaign","id":"abc","attributes":{"name":"test"}}`)
	resp := &JSONAPIResponse{Data: original}

	result := FlattenResponse(resp, true)
	if string(result) != string(original) {
		t.Errorf("raw mode should return original data\ngot: %s\nwant: %s", result, original)
	}
}

func TestFlattenWithRelationship(t *testing.T) {
	resp := &JSONAPIResponse{
		Data: json.RawMessage(`{
			"type": "campaign",
			"id": "camp-1",
			"attributes": {"name": "Sale"},
			"relationships": {
				"campaign-messages": {
					"data": [
						{"type": "campaign-message", "id": "msg-1"},
						{"type": "campaign-message", "id": "msg-2"}
					]
				},
				"tags": {
					"data": [
						{"type": "tag", "id": "tag-1"}
					]
				}
			}
		}`),
	}

	result := FlattenResponse(resp, false)
	var flat map[string]any
	if err := json.Unmarshal(result, &flat); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Array relationships become _ids
	msgIDs, ok := flat["campaign-messages_ids"]
	if !ok {
		t.Fatal("expected campaign-messages_ids in flattened output")
	}
	ids, ok := msgIDs.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", msgIDs)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 message IDs, got %d", len(ids))
	}

	tagIDs, ok := flat["tags_ids"]
	if !ok {
		t.Fatal("expected tags_ids in flattened output")
	}
	tids, ok := tagIDs.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", tagIDs)
	}
	if len(tids) != 1 {
		t.Errorf("expected 1 tag ID, got %d", len(tids))
	}
}

func TestFlattenWithSingleRelationship(t *testing.T) {
	resp := &JSONAPIResponse{
		Data: json.RawMessage(`{
			"type": "event",
			"id": "evt-1",
			"attributes": {"datetime": "2024-01-01"},
			"relationships": {
				"metric": {
					"data": {"type": "metric", "id": "met-1"}
				}
			}
		}`),
	}

	result := FlattenResponse(resp, false)
	var flat map[string]any
	if err := json.Unmarshal(result, &flat); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if flat["metric_id"] != "met-1" {
		t.Errorf("expected metric_id=met-1, got %v", flat["metric_id"])
	}
}

func TestFlattenNilResponse(t *testing.T) {
	result := FlattenResponse(nil, false)
	if string(result) != "null" {
		t.Errorf("expected null, got %s", result)
	}
}

func TestFlattenAggregates(t *testing.T) {
	resp := &JSONAPIResponse{
		Data: json.RawMessage(`{
			"type": "metric-aggregate",
			"attributes": {
				"data": [
					{"dimensions": ["2024-01-01"], "measurements": {"count": 100}},
					{"dimensions": ["2024-01-02"], "measurements": {"count": 150}}
				]
			}
		}`),
	}

	result := FlattenAggregates(resp)
	var items []map[string]any
	if err := json.Unmarshal(result, &items); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 aggregate items, got %d", len(items))
	}
}

func TestFlattenAggregatesEmpty(t *testing.T) {
	resp := &JSONAPIResponse{
		Data: json.RawMessage(`{"type": "metric-aggregate", "attributes": {"data": []}}`),
	}

	result := FlattenAggregates(resp)
	if string(result) != "[]" {
		t.Errorf("expected [], got %s", result)
	}
}
