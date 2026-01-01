package moneropayclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/moneropay/go-monero/walletrpc"
	"gitlab.com/moneropay/moneropay/v2/pkg/model"
)

// Helper function to create a test server
func createTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *MoneroPayAPIClient) {
	ts := httptest.NewServer(handler)

	cfg := Config{
		Endpoint: ts.URL + "/",
	}

	client := NewMoneroPayAPIClient(cfg)
	return ts, client
}

func TestHealth(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		response := map[string]interface{}{
			"status": 200,
			"services": map[string]bool{
				"walletrpc":  true,
				"postgresql": true,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	ts, client := createTestServer(t, handler)
	defer ts.Close()

	ctx := context.Background()
	resp, err := client.Health(ctx)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.Status)
	assert.True(t, resp.Services.WalletRPC)
	assert.True(t, resp.Services.PostgreSQL)
}

// TODO
/*
func TestHealth_ServiceUnavailable(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status": 503,
			"services": map[string]bool{
				"walletrpc":  false,
				"postgresql": true,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(503)
		json.NewEncoder(w).Encode(response)
	}

	ts, client := createTestServer(t, handler)
	defer ts.Close()

	ctx := context.Background()
	resp, err := client.Health(ctx)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 503, resp.Status)
	assert.False(t, resp.Services.WalletRPC)
	assert.True(t, resp.Services.PostgreSQL)
}
*/

func TestReceive(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/receive", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request body
		var req model.ReceivePostRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Validate request
		assert.Equal(t, uint64(123000000), req.Amount)
		assert.Equal(t, "Server expenses", req.Description)
		assert.Equal(t, "http://merchant/callback", req.CallbackUrl)

		// Create response
		response := map[string]interface{}{
			"address":     "84WsptnLmjTYQjm52SMkhQWsepprkcchNguxdyLkURTSW1WLo3tShTnCRvepijbc2X8GAKPGxJK9hfQhLHzoKSxh7y8Yqrg",
			"amount":      123000000,
			"description": "Server expenses",
			"created_at":  "2022-07-18T11:54:49.780542861Z",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	ts, client := createTestServer(t, handler)
	defer ts.Close()

	ctx := context.Background()
	req := model.ReceivePostRequest{
		Amount:      123000000,
		Description: "Server expenses",
		CallbackUrl: "http://merchant/callback",
	}

	resp, err := client.Receive(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "84WsptnLmjTYQjm52SMkhQWsepprkcchNguxdyLkURTSW1WLo3tShTnCRvepijbc2X8GAKPGxJK9hfQhLHzoKSxh7y8Yqrg", resp.Address)
	assert.Equal(t, uint64(123000000), resp.Amount)
	assert.Equal(t, "Server expenses", resp.Description)
}

func TestReceivedPerAddress(t *testing.T) {
	address := "84WsptnLmjTYQjm52SMkhQWsepprkcchNguxdyLkURTSW1WLo3tShTnCRvepijbc2X8GAKPGxJK9hfQhLHzoKSxh7y8Yqrg"

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/receive/"+address, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		// Check query parameters
		assert.Equal(t, "1000", r.URL.Query().Get("min"))
		assert.Equal(t, "2000", r.URL.Query().Get("max"))

		response := map[string]interface{}{
			"amount": map[string]interface{}{
				"expected": 200000000,
				"covered": map[string]interface{}{
					"total":    200000000,
					"unlocked": 200000000,
				},
			},
			"complete":    true,
			"description": "Donation to Kernal",
			"created_at":  "2022-07-11T19:04:24.574583Z",
			"transactions": []map[string]interface{}{
				{
					"amount":            200000000,
					"confirmations":     10,
					"double_spend_seen": false,
					"fee":               9200000,
					"height":            2402648,
					"timestamp":         "2022-07-11T19:19:05Z",
					"tx_hash":           "0c9a7b40b15596fa9a06ba32463a19d781c075120bb59ab5e4ed2a97ab3b7f33",
					"unlock_time":       0,
					"locked":            false,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	ts, client := createTestServer(t, handler)
	defer ts.Close()

	ctx := context.Background()
	minHeight := 1000
	maxHeight := 2000
	resp, err := client.ReceivedPerAddress(ctx, address, &minHeight, &maxHeight)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint64(200000000), resp.Amount.Expected)
	assert.Equal(t, uint64(200000000), resp.Amount.Covered.Total)
	assert.Equal(t, uint64(200000000), resp.Amount.Covered.Unlocked)
	assert.True(t, resp.Complete)
	assert.Equal(t, "Donation to Kernal", resp.Description)
	require.Len(t, resp.Transactions, 1)
	assert.Equal(t, uint64(200000000), resp.Transactions[0].Amount)
	assert.Equal(t, "0c9a7b40b15596fa9a06ba32463a19d781c075120bb59ab5e4ed2a97ab3b7f33", resp.Transactions[0].TxHash)
}

func TestTransfer(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/transfer", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request body
		var req model.TransferPostRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Validate request
		require.Len(t, req.Destinations, 1)
		assert.Equal(t, "47stn...", req.Destinations[0].Address)
		assert.Equal(t, uint64(1337000000), req.Destinations[0].Amount)

		// Create response
		response := map[string]interface{}{
			"amount":       1337000000,
			"fee":          87438594,
			"tx_hash":      "5ca34...",
			"tx_hash_list": []string{"5ca34...", "cf448..."},
			"destinations": []map[string]interface{}{
				{
					"amount":  1337000000,
					"address": "47stn...",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	ts, client := createTestServer(t, handler)
	defer ts.Close()

	ctx := context.Background()
	req := model.TransferPostRequest{
		Destinations: []walletrpc.Destination{
			{
				Address: "47stn...",
				Amount:  1337000000,
			},
		},
	}

	resp, err := client.Transfer(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint64(1337000000), resp.Amount)
	assert.Equal(t, uint64(87438594), resp.Fee)
	assert.Len(t, resp.TxHashList, 2)
	assert.Equal(t, "5ca34...", resp.TxHashList[0])
	assert.Equal(t, "cf448...", resp.TxHashList[1])
	assert.Len(t, resp.Destinations, 1)
	assert.Equal(t, "47stn...", resp.Destinations[0].Address)
}

func TestTransferInfo(t *testing.T) {
	txHash := "cf448effb86f24f81476c0012a6636700488e13accd91f8f43302ae90fed25ce"

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/transfer/"+txHash, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		response := map[string]interface{}{
			"amount":            79990000,
			"fee":               9110000,
			"state":             "completed",
			"confirmations":     15,
			"double_spend_seen": false,
			"height":            2407445,
			"timestamp":         "2022-07-18T11:37:50Z",
			"unlock_time":       10,
			"tx_hash":           txHash,
			"transfer": []map[string]interface{}{
				{
					"amount":  79990000,
					"address": "453biCQpM6oSSr7jgTwmtC9YfiXUWZY1wEfSZJD4r6rf7mPqPj8NZpp7WYpAHVq7p69SYa1B1zMN6SeRc8exYi1WEenqu2c",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	ts, client := createTestServer(t, handler)
	defer ts.Close()

	ctx := context.Background()
	resp, err := client.TransferInfo(ctx, txHash)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint64(79990000), resp.Amount)
	assert.Equal(t, uint64(9110000), resp.Fee)
	assert.Equal(t, "completed", resp.State)
	assert.Equal(t, uint64(15), resp.Confirmations)
	assert.False(t, resp.DoubleSpendSeen)
	assert.Equal(t, uint64(2407445), resp.Height)
	assert.Equal(t, txHash, resp.TxHash)
	// require.Len(t, resp.Transfer, 1)
	// assert.Equal(t, uint64(79990000), resp.Transfer[0].Amount)
}

func TestErrorHandling(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":  400,
			"message": "Invalid request parameters",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(response)
	}

	ts, client := createTestServer(t, handler)
	defer ts.Close()

	ctx := context.Background()
	req := model.ReceivePostRequest{
		Amount: 0, // Invalid amount
	}

	resp, err := client.Receive(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 400")
	assert.Contains(t, err.Error(), "Invalid request parameters")
	assert.Nil(t, resp)
}

func TestWithCustomHTTPClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	cfg := Config{
		Endpoint: "http://example.com/",
	}

	client := NewMoneroPayAPIClient(cfg, WithCustomHTTPClient(customClient))
	assert.Equal(t, customClient, client.client)
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		path     string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid endpoint with trailing slash",
			endpoint: "https://api.moneropay.eu/",
			path:     "health",
			expected: "https://api.moneropay.eu/health",
			wantErr:  false,
		},
		{
			name:     "valid endpoint without trailing slash",
			endpoint: "https://api.moneropay.eu",
			path:     "health",
			expected: "https://api.moneropay.eu/health",
			wantErr:  false,
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			path:     "health",
			wantErr:  true,
		},
		{
			name:     "invalid URL",
			endpoint: ":invalid:",
			path:     "health",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MoneroPayAPIClient{
				cfg: Config{Endpoint: tt.endpoint},
			}

			url, err := client.buildURL(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, url)
			}
		})
	}
}

func TestContextCancellation(t *testing.T) {
	// Create a server that delays response
	handler := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		response := map[string]any{
			"status": 200,
			"services": map[string]bool{
				"walletrpc": true,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	ts, client := createTestServer(t, handler)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.Health(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
