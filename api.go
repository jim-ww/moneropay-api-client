package moneropayclient

import (
	"context"

	"gitlab.com/moneropay/moneropay/v2/pkg/model"
)

type MoneroPayAPI interface {
	// Get the entire wallet balance.
	Balance(context.Context) (*model.BalanceResponse, error)
	// Check if required services are up.
	Health(context.Context) (*model.HealthResponse, error)
	// Create a subaddress for incoming transfers.
	Receive(context.Context, model.ReceivePostRequest) (*model.ReceivePostResponse, error)
	// View incoming transfers for a subaddress.
	// Optionally filter "transactions" by min and max block height.
	ReceivedPerAddress(ctx context.Context, address string, minHeight, maxHeight *int) (*model.ReceiveGetResponse, error)
	// Transfer to a single or multiple recipients. If necessary, split the transfer into multiple transactions.
	// NOTE: This transaction uses balance of the wallet's Primary Account.
	Transfer(context.Context, model.TransferPostRequest) (*model.TransferPostResponse, error)
	// Get information about transaction via its hash.
	TransferInfo(ctx context.Context, txHash string) (*model.TransferGetResponse, error)
}
