package aastocks

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestServePrices(t *testing.T) {
	testCases := []struct {
		desc     string
		symbol   string
		requests map[string]http.HandlerFunc
		err      error
		prices   []PriceResult
	}{
		{
			desc:   "WithContextCancellation",
			symbol: "00006",
			requests: map[string]http.HandlerFunc{
				"GET-http://www.aastocks.com/en/stocks/quote/detail-quote.aspx?symbol=00006": serveAll(
					serveFile("testdata/detail_quote.html"), // First called with getting quote

					serveFile("testdata/detail_quote.html"),
					serveFile("testdata/detail_quote_00006_2.html"),
				),
			},
			prices: []PriceResult{
				{
					Symbol: "00006",
					Price:  44.65,
					Time:   time.Date(2020, time.August, 25, 21, 18, 38, 0, time.UTC),
				},
				{
					Symbol: "00006",
					Price:  44.4,
					Time:   time.Date(2020, time.August, 29, 00, 55, 31, 0, time.UTC),
				},
			},
		},
		{
			desc:   "WithErrorChannel",
			symbol: "00006",
			requests: map[string]http.HandlerFunc{
				"GET-http://www.aastocks.com/en/stocks/quote/detail-quote.aspx?symbol=00006": serveAll(
					serveFile("testdata/detail_quote.html"), // First called with getting quote

					serveFile("testdata/detail_quote.html"),
					serveError(fmt.Errorf("testing error")),
				),
			},
			err: fmt.Errorf("Failed to fetch quote details"),
			prices: []PriceResult{
				{
					Symbol: "00006",
					Price:  44.65,
					Time:   time.Date(2020, time.August, 25, 21, 18, 38, 0, time.UTC),
				},
			},
		},
	}

	mock := mockClient()
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mock.set(tC.requests)

			checkErrorFunc(t, tC.err, func() error {
				quote, err := Get(tC.symbol, WithClient(mock.client))
				if err != nil {
					return err
				}

				ctx := context.Background()
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()

				pricesChan, errChan := quote.ServePrices(ctx, 100*time.Millisecond)

				prices := make([]PriceResult, 0)
				timeout := time.After(2 * time.Second)

			Loop:
				for {
					select {
					case p := <-pricesChan:
						prices = append(prices, p)
						if len(prices) == 2 {
							cancel()
							break Loop
						}
					case err = <-errChan:
						cancel()
						break Loop
					case <-timeout:
						return fmt.Errorf("Timeout is triggered, expect case to terminate normally")
					}
				}

				diff := cmp.Diff(tC.prices, prices)
				if diff != "" {
					t.Fatalf(diff)
				}
				return err
			})
		})
	}
}
