package moneropayclient

import (
	"context"
	"time"
)

type MoneroPayAPI interface {
	// Get the entire wallet balance.
	Balance(context.Context) (*BalanceResponse, error)
	// Check if required services are up.
	Health(context.Context) (*HealthResponse, error)
	// Create a subaddress for incoming transfers.
	Receive(context.Context, ReceiveRequest) (*ReceiveResponse, error)
	// View incoming transfers for a subaddress.
	// Optionally filter "transactions" by min and max block height.
	ReceivedPerAddress(ctx context.Context, address string, minHeight, maxHeight *int) (*PaymentResponse, error)
	// Transfer to a single or multiple recipients. If necessary, split the transfer into multiple transactions.
	// NOTE: This transaction uses balance of the wallet's Primary Account.
	Transfer(context.Context) (*TransferResponse, error)
	// Get information about transaction via its hash.
	GetTransferInfo(ctx context.Context, txHash string) (*TransferInfoResponse, error)
}

type BalanceResponse struct {
	Total    int64 `json:"total"`
	Unlocked int64 `json:"unlocked"`
}

type HealthResponse struct {
	Status   int             `json:"status"`
	Services map[string]bool `json:"services"`
}

type ReceiveRequest struct {
	Amount int64 `json:"amount"`
	// optional
	Description string `json:"description"`
	// optional
	// If "callback_url" is set, MoneroPay will send a POST request to URL specified with a payload described here:
	// https://moneropay.eu/api/callback.html
	CallbackURL string `json:"callback_url"`
}

type ReceiveResponse struct {
	Address     string    `json:"address"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// NOTE: is also used for callbacks
type PaymentResponse struct {
	Amount struct {
		Expected int64 `json:"expected"`
		Covered  struct {
			Total    int64 `json:"total"`
			Unlocked int64 `json:"unlocked"`
		} `json:"covered"`
	} `json:"amount"`

	// "complete" will be set to true inside callback and GET /receive/:subaddress payload, when unlocked amount is equal or more to the one specified in "amount".
	Complete     bool          `json:"complete"`
	Description  string        `json:"description"`
	CreatedAt    time.Time     `json:"created_at"` // ISO 8601 timestamp
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Amount          int64  `json:"amount"`
	Confirmations   int64  `json:"confirmations"`
	DoubleSpendSeen bool   `json:"double_spend_seen"`
	Fee             int64  `json:"fee"`
	Height          int64  `json:"height"`
	Timestamp       string `json:"timestamp"` // ISO 8601
	TxHash          string `json:"tx_hash"`
	UnlockTime      int64  `json:"unlock_time"`
	Locked          bool   `json:"locked"`
}

type TransferRequest struct {
	Destinations []Destination `json:"destinations"`
}

type TransferResponse struct {
	Amount int64 `json:"amount"`
	Fee    int64 `json:"fee"`
	// Deprecated: TxHash (tx_hash) field will be removed the next major release (3.0.0). Please use TxHashList (tx_hash_list) instead. See here for more details.
	TxHash       string        `json:"tx_hash"`
	TxHashList   []string      `json:"tx_hash_list"`
	Destinations []Destination `json:"destinations"`
}

type Destination struct {
	Amount  int64  `json:"amount"`
	Address string `json:"address"`
}

type TransferInfoResponse struct {
	Amount          int64       `json:"amount"`
	Fee             int64       `json:"fee"`
	State           string      `json:"state"`
	Transfer        []Recipient `json:"transfer"`
	Confirmations   int64       `json:"confirmations"`
	DoubleSpendSeen bool        `json:"double_spend_seen"`
	Height          int64       `json:"height"`
	Timestamp       string      `json:"timestamp"` // ISO 8601
	UnlockTime      int64       `json:"unlock_time"`
	TxHash          string      `json:"tx_hash"`
}

type Recipient struct {
	Amount  int64  `json:"amount"`
	Address string `json:"address"`
}
