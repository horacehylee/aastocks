package aastocks

import (
	"context"
	"time"
)

// PriceResult is the result of serving real time prices.
type PriceResult struct {
	Price  float64
	Symbol string
	Time   time.Time
}

// ServePrices continously fetching latest price from AAStocks.
// It will start goroutine to fetch real time prices.
func (q *Quote) ServePrices(ctx context.Context, delay time.Duration) (<-chan PriceResult, <-chan error) {
	prices := make(chan PriceResult)
	errors := make(chan error)

	go func() {
		var priceChan chan<- PriceResult
		var errChan chan<- error
		qq := q.clone()

		var err error
		var price PriceResult
		timeout := time.After(0)
		for {
			select {
			case <-ctx.Done():
				return
			case errChan <- err:
				errChan = nil
				err = nil
			case priceChan <- price:
				priceChan = nil
				price = PriceResult{}
			case <-timeout:
				err = qq.details()
				if err != nil {
					errChan = errors
				} else {
					price = PriceResult{
						Price:  qq.Price,
						Symbol: qq.Symbol,
						Time:   qq.UpdateTime,
					}
					priceChan = prices
				}
				timeout = time.After(delay)
			}
		}
	}()
	return prices, errors
}
