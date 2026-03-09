package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBaseURL  = "https://a.klaviyo.com/api"
	DefaultRevision = "2024-10-15"
	MaxRetries      = 3
	MaxPages        = 50
	Timeout         = 30 * time.Second
	ContentType     = "application/vnd.api+json"
)

type Client struct {
	http       *http.Client
	apiKey     string
	baseURL    string
	revision   string
	verbose    bool
	maxRetries int
}

// JSONAPIResponse represents a Klaviyo JSON:API response.
type JSONAPIResponse struct {
	Data     json.RawMessage `json:"data"`
	Included json.RawMessage `json:"included,omitempty"`
	Links    *Links          `json:"links,omitempty"`
	Errors   []ErrorEntry    `json:"errors,omitempty"`
}

type Links struct {
	Self string `json:"self,omitempty"`
	Next string `json:"next,omitempty"`
	Prev string `json:"prev,omitempty"`
}

type ErrorEntry struct {
	ID     string `json:"id"`
	Status any    `json:"status"`
	Code   string `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func New(apiKey, baseURL, revision string, verbose bool) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if revision == "" {
		revision = DefaultRevision
	}
	return &Client{
		http:       &http.Client{Timeout: Timeout},
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(baseURL, "/"),
		revision:   revision,
		verbose:    verbose,
		maxRetries: MaxRetries,
	}
}

// Get performs a single-page GET request.
func (c *Client) Get(ctx context.Context, path string, params url.Values) (*JSONAPIResponse, error) {
	u := c.buildURL(path, params)
	return c.doRequest(ctx, http.MethodGet, u, nil)
}

// GetAll performs an auto-paginating GET, following links.next.
// maxResults of 0 means unlimited.
func (c *Client) GetAll(ctx context.Context, path string, params url.Values, maxResults int) (json.RawMessage, error) {
	u := c.buildURL(path, params)

	var allItems []json.RawMessage

	for page := 0; page < MaxPages; page++ {
		resp, err := c.doRequest(ctx, http.MethodGet, u, nil)
		if err != nil {
			if len(allItems) > 0 {
				return mergeItems(allItems), err
			}
			return nil, err
		}

		// Parse data — could be array or single object
		var items []json.RawMessage
		if err := json.Unmarshal(resp.Data, &items); err != nil {
			// Single object, return as-is
			return resp.Data, nil
		}
		allItems = append(allItems, items...)

		// Check max results cap
		if maxResults > 0 && len(allItems) >= maxResults {
			allItems = allItems[:maxResults]
			break
		}

		// Check for next page
		if resp.Links == nil || resp.Links.Next == "" {
			break
		}
		u = resp.Links.Next
	}

	return mergeItems(allItems), nil
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body any) (*JSONAPIResponse, error) {
	u := c.buildURL(path, nil)
	return c.doJSON(ctx, http.MethodPost, u, body)
}

// Patch performs a PATCH request with a JSON body.
func (c *Client) Patch(ctx context.Context, path string, body any) (*JSONAPIResponse, error) {
	u := c.buildURL(path, nil)
	return c.doJSON(ctx, http.MethodPatch, u, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	u := c.buildURL(path, nil)
	_, err := c.doRequest(ctx, http.MethodDelete, u, nil)
	return err
}

func (c *Client) buildURL(path string, params url.Values) string {
	// If path is already a full URL (pagination links), use directly
	if strings.HasPrefix(path, "http") {
		return path
	}
	u := c.baseURL + "/" + strings.TrimLeft(path, "/")
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	return u
}

func (c *Client) doJSON(ctx context.Context, method, rawURL string, body any) (*JSONAPIResponse, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}
	return c.doRequest(ctx, method, rawURL, &buf)
}

func (c *Client) doRequest(ctx context.Context, method, rawURL string, body io.Reader) (*JSONAPIResponse, error) {
	// Read body into buffer if present, so we can retry
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("read request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	if c.verbose {
		fmt.Fprintf(os.Stderr, "> %s %s\n", method, rawURL)
		if len(bodyBytes) > 0 {
			fmt.Fprintf(os.Stderr, "> Body: %s\n", string(bodyBytes))
		}
	}

	start := time.Now()
	resp, err := c.doWithRetry(ctx, method, rawURL, bodyBytes)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "< %d %s (%s, %s)\n", resp.StatusCode, http.StatusText(resp.StatusCode), elapsed.Round(time.Millisecond), humanBytes(len(respBody)))
	}

	// Handle DELETE with no content
	if resp.StatusCode == http.StatusNoContent {
		return &JSONAPIResponse{}, nil
	}

	// Handle error responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if c.verbose {
			fmt.Fprintf(os.Stderr, "< Body: %s\n", string(respBody))
		}
		return nil, c.parseError(respBody, resp.StatusCode, rawURL)
	}

	var apiResp JSONAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unexpected response format: %w", err)
	}
	return &apiResp, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Klaviyo-API-Key "+c.apiKey)
	req.Header.Set("Revision", c.revision)
	req.Header.Set("Accept", ContentType)
	if req.Method == http.MethodPost || req.Method == http.MethodPatch {
		req.Header.Set("Content-Type", ContentType)
	}
}

func (c *Client) doWithRetry(ctx context.Context, method, rawURL string, bodyBytes []byte) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		var bodyReader io.Reader
		if len(bodyBytes) > 0 {
			bodyReader = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		c.setHeaders(req)

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			if attempt == c.maxRetries {
				return nil, lastErr
			}
			delay := retryDelay(attempt, "")
			if c.verbose {
				fmt.Fprintf(os.Stderr, "! request error, retrying in %s (attempt %d/%d)\n", delay, attempt+1, c.maxRetries)
			}
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// Retry on 429 and 5xx
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			resp.Body.Close()
			if attempt == c.maxRetries {
				return nil, &APIError{
					StatusCode: resp.StatusCode,
					Title:      http.StatusText(resp.StatusCode),
					Detail:     fmt.Sprintf("failed after %d retries", c.maxRetries),
				}
			}
			delay := retryDelay(attempt, resp.Header.Get("Retry-After"))
			if c.verbose {
				fmt.Fprintf(os.Stderr, "! %d %s, retrying in %s (attempt %d/%d)\n",
					resp.StatusCode, http.StatusText(resp.StatusCode), delay, attempt+1, c.maxRetries)
			}
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		return resp, nil
	}
	return nil, lastErr
}

func (c *Client) parseError(body []byte, statusCode int, rawURL string) *APIError {
	apiErr := &APIError{StatusCode: statusCode}

	// Try Klaviyo's JSON:API error format
	var errResp struct {
		Errors []ErrorEntry `json:"errors"`
	}
	if json.Unmarshal(body, &errResp) == nil && len(errResp.Errors) > 0 {
		e := errResp.Errors[0]
		apiErr.Code = e.Code
		apiErr.Title = e.Title
		apiErr.Detail = e.Detail
	}

	if apiErr.Detail == "" && apiErr.Title == "" {
		switch statusCode {
		case 401:
			apiErr.Detail = "authentication failed — check KLAVIYO_API_KEY env var or --api-key flag"
		case 403:
			apiErr.Detail = "forbidden — your API key may not have access to this resource"
		case 404:
			apiErr.Detail = "resource not found"
		default:
			apiErr.Detail = http.StatusText(statusCode)
		}
	}

	apiErr.Hint = hintForError(statusCode)
	return apiErr
}

func retryDelay(attempt int, retryAfter string) time.Duration {
	if retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	base := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	jitter := time.Duration(rand.IntN(500)) * time.Millisecond
	return base + jitter
}

func hintForError(statusCode int) string {
	switch statusCode {
	case 401:
		return "check your API key: --api-key flag, KLAVIYO_API_KEY env, or 'kv config list'"
	case 403:
		return "your API key may lack the required scopes for this endpoint"
	case 429:
		return "rate limited — reduce request frequency or wait a moment"
	}
	return ""
}

func humanBytes(b int) string {
	if b < 1024 {
		return fmt.Sprintf("%dB", b)
	}
	kb := float64(b) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.1fKB", kb)
	}
	return fmt.Sprintf("%.1fMB", kb/1024)
}

func mergeItems(items []json.RawMessage) json.RawMessage {
	if len(items) == 0 {
		return json.RawMessage("[]")
	}
	result := []byte("[")
	for i, item := range items {
		if i > 0 {
			result = append(result, ',')
		}
		result = append(result, item...)
	}
	result = append(result, ']')
	return json.RawMessage(result)
}
