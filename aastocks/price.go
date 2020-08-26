package aastocks

import (
	"context"
	"time"
)

// Prices continously fetching latest price from AAStocks
// It will start goroutine to fetch prices
func (q *Quote) Prices(ctx context.Context, delay time.Duration) (<-chan float64, <-chan error) {
	prices := make(chan float64)
	errors := make(chan error)

	go func() {
		var priceChan chan<- float64
		var errChan chan<- error
		qq := q.clone()

		var err error
		timeout := time.After(0)
		for {
			select {
			case <-ctx.Done():
				return
			case errChan <- err:
				errChan = nil
			case priceChan <- qq.Price:
				priceChan = nil
			case <-timeout:
				err = qq.getQuoteDetail()
				if err != nil {
					errChan = errors
				} else {
					priceChan = prices
				}
				timeout = time.After(delay)
			}
		}
	}()
	return prices, errors
}
