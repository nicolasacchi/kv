package output

import (
	"encoding/json"
	"testing"
)

func TestApplyFilter_SimplePath(t *testing.T) {
	data := json.RawMessage(`{"name": "Test", "status": "Draft"}`)
	result, err := ApplyFilter(data, "name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != `"Test"` {
		t.Errorf("expected \"Test\", got %s", result)
	}
}

func TestApplyFilter_NestedPath(t *testing.T) {
	data := json.RawMessage(`[
		{"id": "1", "name": "A"},
		{"id": "2", "name": "B"}
	]`)
	result, err := ApplyFilter(data, "#.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != `["A","B"]` {
		t.Errorf("expected [\"A\",\"B\"], got %s", result)
	}
}

func TestApplyFilter_ObjectExtraction(t *testing.T) {
	data := json.RawMessage(`[
		{"id": "1", "name": "A", "status": "Draft"},
		{"id": "2", "name": "B", "status": "Sent"}
	]`)
	result, err := ApplyFilter(data, `#.{id:id,name:name}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []map[string]string
	if err := json.Unmarshal(result, &items); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestApplyFilter_NonexistentPath(t *testing.T) {
	data := json.RawMessage(`{"name": "Test"}`)
	result, err := ApplyFilter(data, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "null" {
		t.Errorf("expected null for missing path, got %s", result)
	}
}

func TestApplyFilter_EmptyPath(t *testing.T) {
	data := json.RawMessage(`{"name": "Test"}`)
	result, err := ApplyFilter(data, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != string(data) {
		t.Errorf("empty path should return data unchanged")
	}
}
