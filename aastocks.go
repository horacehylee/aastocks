package aastocks

import (
	"net/http"
	"time"
)

// Quote of AAStocks data
type Quote struct {
	Symbol     string
	Name       string
	Price      float64
	Yield      float64
	PeRatio    float64
	PbRatio    float64
	Lots       int
	Eps        float64
	UpdateTime time.Time

	client    *http.Client
	dividends []*Dividend
}

func (q *Quote) clone() *Quote {
	return &Quote{
		Symbol: q.Symbol,
		client: q.client,
	}
}

// Get quote from AAStocks with symbol
func Get(symbol string, opts ...Option) (*Quote, error) {
	client, err := defaultClient()
	if err != nil {
		return nil, err
	}
	q := &Quote{
		Symbol: symbol,
		client: client,
	}
	for _, opt := range opts {
		opt(q)
	}
	return q, q.details()
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
