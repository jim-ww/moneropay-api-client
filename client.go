package moneropayclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"gitlab.com/moneropay/moneropay/v2/pkg/model"
)

// ensure that MoneroPayAPIClient implements MoneroPayAPI
var _ MoneroPayAPI = (*MoneroPayAPIClient)(nil)

type Config struct {
	Endpoint string `json:"endpoint" yaml:"endpoint" toml:"endpoint" env:"MONERO_PAY_ENDPOINT"`
}

type MoneroPayAPIClient struct {
	cfg    Config
	client *http.Client
}

type Option func(*MoneroPayAPIClient)

func WithCustomHTTPClient(client *http.Client) Option {
	return func(mpa *MoneroPayAPIClient) {
		mpa.client = client
	}
}

func NewMoneroPayAPIClient(cfg Config, opts ...Option) *MoneroPayAPIClient {
	mpa := &MoneroPayAPIClient{cfg: cfg, client: http.DefaultClient}

	for _, opt := range opts {
		opt(mpa)
	}

	return mpa
}

// Helper function to build the full URL
func (c *MoneroPayAPIClient) buildURL(path string) (string, error) {
	endpoint := c.cfg.Endpoint
	if endpoint == "" {
		return "", fmt.Errorf("endpoint is required")
	}

	// Ensure the endpoint ends with a slash for proper path joining
	if endpoint[len(endpoint)-1] != '/' {
		endpoint += "/"
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint URL: %w", err)
	}

	u.Path += path
	return u.String(), nil
}

// Helper function to make HTTP requests
func (c *MoneroPayAPIClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url, err := c.buildURL(path)
	if err != nil {
		return nil, err
	}

	var reqBody []byte
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// Helper function to parse response
func (c *MoneroPayAPIClient) parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Try to parse error response
		var apiErr struct {
			Status  int    `json:"status"`
			Message string `json:"message,omitempty"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Message != "" {
			return fmt.Errorf("API error %d: %s", apiErr.Status, apiErr.Message)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Health implements MoneroPayAPI interface
func (c *MoneroPayAPIClient) Health(ctx context.Context) (*model.HealthResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "health", nil)
	if err != nil {
		return nil, err
	}

	var healthResp model.HealthResponse
	if err := c.parseResponse(resp, &healthResp); err != nil {
		return nil, err
	}

	return &healthResp, nil
}

// Balance implements MoneroPayAPI interface
func (c *MoneroPayAPIClient) Balance(ctx context.Context) (*model.BalanceResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "balance", nil)
	if err != nil {
		return nil, err
	}

	var balanceResp model.BalanceResponse
	if err := c.parseResponse(resp, &balanceResp); err != nil {
		return nil, err
	}

	return &balanceResp, nil
}

// Receive implements MoneroPayAPI interface
func (c *MoneroPayAPIClient) Receive(ctx context.Context, req model.ReceivePostRequest) (*model.ReceivePostResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "receive", req)
	if err != nil {
		return nil, err
	}

	var receiveResp model.ReceivePostResponse
	if err := c.parseResponse(resp, &receiveResp); err != nil {
		return nil, err
	}

	return &receiveResp, nil
}

// ReceivedPerAddress implements MoneroPayAPI interface (GET /receive/:address)
func (c *MoneroPayAPIClient) ReceivedPerAddress(ctx context.Context, address string, minHeight, maxHeight *int) (*model.ReceiveGetResponse, error) {
	// Build query parameters
	params := url.Values{}
	if minHeight != nil {
		params.Add("min", strconv.Itoa(*minHeight))
	}
	if maxHeight != nil {
		params.Add("max", strconv.Itoa(*maxHeight))
	}

	// Get the base URL
	baseURL, err := c.buildURL("receive/" + address)
	if err != nil {
		return nil, err
	}

	// Append query parameters
	fullURL := baseURL
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var receiveResp model.ReceiveGetResponse
	if err := c.parseResponse(resp, &receiveResp); err != nil {
		return nil, err
	}

	return &receiveResp, nil
}

// Transfer implements MoneroPayAPI interface
func (c *MoneroPayAPIClient) Transfer(ctx context.Context, req model.TransferPostRequest) (*model.TransferPostResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "transfer", req)
	if err != nil {
		return nil, err
	}

	var transferResp model.TransferPostResponse
	if err := c.parseResponse(resp, &transferResp); err != nil {
		return nil, err
	}

	return &transferResp, nil
}

// TransferInfo implements MoneroPayAPI interface (GET /transfer/:tx_hash)
func (c *MoneroPayAPIClient) TransferInfo(ctx context.Context, txHash string) (*model.TransferGetResponse, error) {
	// Get the base URL
	baseURL, err := c.buildURL("transfer/" + txHash)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var transferResp model.TransferGetResponse
	if err := c.parseResponse(resp, &transferResp); err != nil {
		return nil, err
	}

	return &transferResp, nil
}
