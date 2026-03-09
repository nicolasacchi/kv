package client

import (
	"encoding/json"
)

// Resource represents a JSON:API resource object.
type Resource struct {
	Type          string                       `json:"type"`
	ID            string                       `json:"id"`
	Attributes    map[string]json.RawMessage   `json:"attributes,omitempty"`
	Relationships map[string]json.RawMessage   `json:"relationships,omitempty"`
	Links         map[string]json.RawMessage   `json:"links,omitempty"`
}

// FlattenResponse takes a JSON:API response and flattens data objects.
// It merges id and type into the attributes for a cleaner structure.
// If raw is true, returns the original data unchanged.
func FlattenResponse(resp *JSONAPIResponse, raw bool) json.RawMessage {
	if resp == nil || resp.Data == nil {
		return json.RawMessage("null")
	}
	if raw {
		return resp.Data
	}

	// Try as array first
	var resources []Resource
	if err := json.Unmarshal(resp.Data, &resources); err == nil {
		return flattenArray(resources)
	}

	// Try as single object
	var resource Resource
	if err := json.Unmarshal(resp.Data, &resource); err == nil {
		return flattenSingle(resource)
	}

	// Can't flatten — return as-is
	return resp.Data
}

func flattenSingle(r Resource) json.RawMessage {
	flat := make(map[string]json.RawMessage)

	// Start with id and type
	idBytes, _ := json.Marshal(r.ID)
	typeBytes, _ := json.Marshal(r.Type)
	flat["id"] = idBytes
	flat["type"] = typeBytes

	// Merge attributes
	for k, v := range r.Attributes {
		flat[k] = v
	}

	// Add relationship IDs as flat references
	for name, relData := range r.Relationships {
		var rel struct {
			Data json.RawMessage `json:"data"`
		}
		if json.Unmarshal(relData, &rel) == nil && rel.Data != nil {
			// Check if it's a single relationship or array
			var single struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			}
			if json.Unmarshal(rel.Data, &single) == nil && single.ID != "" {
				idVal, _ := json.Marshal(single.ID)
				flat[name+"_id"] = idVal
			} else {
				var multi []struct {
					Type string `json:"type"`
					ID   string `json:"id"`
				}
				if json.Unmarshal(rel.Data, &multi) == nil {
					ids := make([]string, len(multi))
					for i, m := range multi {
						ids[i] = m.ID
					}
					idsVal, _ := json.Marshal(ids)
					flat[name+"_ids"] = idsVal
				}
			}
		}
	}

	result, _ := json.Marshal(flat)
	return result
}

func flattenArray(resources []Resource) json.RawMessage {
	if len(resources) == 0 {
		return json.RawMessage("[]")
	}

	items := make([]json.RawMessage, len(resources))
	for i, r := range resources {
		items[i] = flattenSingle(r)
	}

	result := []byte("[")
	for i, item := range items {
		if i > 0 {
			result = append(result, ',')
		}
		result = append(result, item...)
	}
	result = append(result, ']')
	return result
}

// FlattenRaw flattens a raw json.RawMessage that contains JSON:API resource objects.
// Works for both arrays and single objects. If raw is true, returns unchanged.
func FlattenRaw(data json.RawMessage, raw bool) json.RawMessage {
	if data == nil {
		return json.RawMessage("null")
	}
	if raw {
		return data
	}

	// Try as array first
	var resources []Resource
	if err := json.Unmarshal(data, &resources); err == nil {
		return flattenArray(resources)
	}

	// Try as single object
	var resource Resource
	if err := json.Unmarshal(data, &resource); err == nil {
		return flattenSingle(resource)
	}

	return data
}

// FlattenAggregates handles the special response format from /api/metric-aggregates/.
// The response has data.attributes.data[] with measurement results.
func FlattenAggregates(resp *JSONAPIResponse) json.RawMessage {
	if resp == nil || resp.Data == nil {
		return json.RawMessage("null")
	}

	var resource struct {
		Attributes struct {
			Data []json.RawMessage `json:"data"`
		} `json:"attributes"`
	}
	if err := json.Unmarshal(resp.Data, &resource); err != nil {
		return resp.Data
	}

	if len(resource.Attributes.Data) == 0 {
		return json.RawMessage("[]")
	}

	result := []byte("[")
	for i, item := range resource.Attributes.Data {
		if i > 0 {
			result = append(result, ',')
		}
		result = append(result, item...)
	}
	result = append(result, ']')
	return result
}
