package aastocks

import (
	"net/http"
)

// Quote of AAStocks data
type Quote struct {
	symbol  string
	client  *http.Client
	name    string
	price   float64
	yield   float64
	peRatio float64
	pbRatio float64
	lots    int
	eps     float64
}

// Get quote from AAStocks with symbol
func Get(symbol string, opts ...Option) (*Quote, error) {
	client, err := defaultClient()
	if err != nil {
		return nil, err
	}
	q := &Quote{
		symbol: symbol,
		client: client,
	}
	for _, opt := range opts {
		opt(q)
	}
	return q, q.getQuoteDetail()
}

func defaultClient() (*http.Client, error) {
	t := &transport{r: http.DefaultTransport}
	c := &http.Client{
		Transport: t,
	}
	return c, nil
}

type transport struct {
	r http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Referer", req.URL.String())
	return t.r.RoundTrip(req)
}
