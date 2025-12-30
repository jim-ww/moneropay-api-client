package moneropayclient

import "net/http"

type Config struct {
	Endpoint string `json:"endpoint" yaml:"endpoint" toml:"endpoint" env:"MONERO_PAY_ENDPOINT"`
}

// TODO: implement MoneroPayAPI interface
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
