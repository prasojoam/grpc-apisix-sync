package apisix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/sync/errgroup"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL:    strings.TrimSuffix(baseURL, "/"),
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

type ApisixItem struct {
	Value struct {
		ID json.RawMessage `json:"id"`
	} `json:"value"`
}

type ApisixListResponse struct {
	Total int          `json:"total"`
	List  []ApisixItem `json:"list"`
}

func (c *Client) DeleteAll(resourceType string, prefix string) error {
	url := fmt.Sprintf("%s/apisix/admin/%s", c.BaseURL, resourceType)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("X-API-KEY", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch %s: %d", resourceType, resp.StatusCode)
	}

	var listResp ApisixListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return err
	}

	if listResp.Total == 0 {
		return nil
	}

	// Filter items by prefix
	var targetItems []ApisixItem
	for _, item := range listResp.List {
		id := strings.Trim(string(item.Value.ID), "\"")
		if prefix != "" && !strings.HasPrefix(id, prefix) {
			continue
		}
		targetItems = append(targetItems, item)
	}

	if len(targetItems) == 0 {
		fmt.Printf("No %s found with prefix '%s', skipping cleanup.\n", resourceType, prefix)
		return nil
	}

	fmt.Printf("🗑️ Deleting %d %s...\n", len(targetItems), resourceType)

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(10)

	for _, item := range targetItems {
		id := strings.Trim(string(item.Value.ID), "\"")
		g.Go(func() error {
			deleteUrl := fmt.Sprintf("%s/%s", url, id)
			req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, deleteUrl, nil)
			req.Header.Set("X-API-KEY", c.APIKey)

			dResp, err := c.HTTPClient.Do(req)
			if err != nil {
				return err
			}
			defer dResp.Body.Close()

			if dResp.StatusCode >= 300 {
				body, _ := io.ReadAll(dResp.Body)
				return fmt.Errorf("failed to delete %s %s: %d %s", resourceType, id, dResp.StatusCode, string(body))
			}
			return nil
		})
	}

	return g.Wait()
}

func (c *Client) Put(path string, body interface{}) error {
	url := fmt.Sprintf("%s/apisix/admin%s", c.BaseURL, path)
	jsonData, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("X-API-KEY", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("PUT %s... ", url)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		fmt.Println("❌ Error")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("✅ %d\n", resp.StatusCode)
		return nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("❌ %d: %s\n", resp.StatusCode, string(respBody))
	return fmt.Errorf("request failed with status %d", resp.StatusCode)
}
