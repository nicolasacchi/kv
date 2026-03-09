package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGet_AuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Klaviyo-API-Key test-key" {
			t.Errorf("expected Authorization: Klaviyo-API-Key test-key, got %q", auth)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, "2024-10-15", false)
	_, err := c.Get(context.Background(), "campaigns", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGet_RevisionHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rev := r.Header.Get("Revision")
		if rev != "2024-10-15" {
			t.Errorf("expected Revision: 2024-10-15, got %q", rev)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, "2024-10-15", false)
	_, err := c.Get(context.Background(), "campaigns", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGet_AcceptHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		if accept != "application/vnd.api+json" {
			t.Errorf("expected Accept: application/vnd.api+json, got %q", accept)
		}
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, "2024-10-15", false)
	_, err := c.Get(context.Background(), "campaigns", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGet_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]string{
				{"title": "Not Found", "detail": "Campaign not found"},
			},
		})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, "2024-10-15", false)
	_, err := c.Get(context.Background(), "campaigns/nonexistent", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
	if apiErr.Detail != "Campaign not found" {
		t.Errorf("expected detail 'Campaign not found', got %q", apiErr.Detail)
	}
}

func TestGet_401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]string{
				{"title": "Unauthorized", "detail": "Invalid API key"},
			},
		})
	}))
	defer srv.Close()

	c := New("bad-key", srv.URL, "2024-10-15", false)
	_, err := c.Get(context.Background(), "campaigns", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
	if apiErr.ExitCode() != 3 {
		t.Errorf("expected exit code 3, got %d", apiErr.ExitCode())
	}
}

func TestGet_429Retry(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"type": "campaign", "id": "abc", "attributes": map[string]any{"name": "test"}},
		})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, "2024-10-15", false)
	c.maxRetries = 3
	resp, err := c.Get(context.Background(), "campaigns/abc", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestGetAll_Pagination(t *testing.T) {
	page := 0
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		resp := map[string]any{
			"data": []map[string]any{
				{"type": "campaign", "id": fmt.Sprintf("id-%d", page), "attributes": map[string]any{}},
			},
		}
		if page < 3 {
			nextURL := srvURL + "/campaigns?page[cursor]=page" + fmt.Sprintf("%d", page+1)
			resp["links"] = map[string]string{"next": nextURL}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := New("test-key", srv.URL, "2024-10-15", false)
	data, err := c.GetAll(context.Background(), "campaigns", nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items across 3 pages, got %d", len(items))
	}
}

func TestGetAll_MaxResults(t *testing.T) {
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": []map[string]any{
				{"type": "campaign", "id": "1", "attributes": map[string]any{}},
				{"type": "campaign", "id": "2", "attributes": map[string]any{}},
				{"type": "campaign", "id": "3", "attributes": map[string]any{}},
			},
			"links": map[string]string{"next": srvURL + "/campaigns?page[cursor]=next"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := New("test-key", srv.URL, "2024-10-15", false)
	data, err := c.GetAll(context.Background(), "campaigns", nil, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected max 2 items, got %d", len(items))
	}
}

func TestGet_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Shouldn't reach here
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, "2024-10-15", false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := c.Get(ctx, "campaigns", nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestPost_ContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/vnd.api+json" {
			t.Errorf("expected Content-Type: application/vnd.api+json, got %q", ct)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"type": "campaign-values-report", "id": "rpt-1", "attributes": map[string]any{}},
		})
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, "2024-10-15", false)
	_, err := c.Post(context.Background(), "campaign-values-reports", map[string]any{"data": map[string]any{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildURL(t *testing.T) {
	c := New("key", "https://a.klaviyo.com/api", "2024-10-15", false)

	// Relative path
	u := c.buildURL("campaigns", nil)
	if u != "https://a.klaviyo.com/api/campaigns" {
		t.Errorf("expected https://a.klaviyo.com/api/campaigns, got %s", u)
	}

	// With params
	params := url.Values{"filter": {"equals(status,\"Draft\")"}}
	u = c.buildURL("campaigns", params)
	if !strings.Contains(u, "filter=") {
		t.Errorf("expected URL to contain filter param, got %s", u)
	}

	// Absolute URL (pagination link)
	u = c.buildURL("https://a.klaviyo.com/api/campaigns?page[cursor]=abc", nil)
	if u != "https://a.klaviyo.com/api/campaigns?page[cursor]=abc" {
		t.Errorf("expected absolute URL preserved, got %s", u)
	}
}

func TestRetryDelay(t *testing.T) {
	// With Retry-After header
	d := retryDelay(0, "5")
	if d != 5*1e9 {
		t.Errorf("expected 5s from Retry-After, got %s", d)
	}

	// Without Retry-After, exponential backoff
	d = retryDelay(0, "")
	if d < 1e9 || d > 2e9 {
		t.Errorf("expected ~1s for attempt 0, got %s", d)
	}

	d = retryDelay(1, "")
	if d < 2e9 || d > 3e9 {
		t.Errorf("expected ~2s for attempt 1, got %s", d)
	}
}

func TestDeleteNoContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, "2024-10-15", false)
	err := c.Delete(context.Background(), "webhooks/123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
