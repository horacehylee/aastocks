package aastocks

import (
	"net/http"
)

// Option for getting symbol from AAStocks
type Option func(q *Quote)

// WithClient to customize the HTTP client used for AAStocks
func WithClient(client *http.Client) Option {
	return func(q *Quote) {
		q.client = client
	}
}
