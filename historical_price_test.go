package aastocks

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestHistoricalPrice(t *testing.T) {
	mock := mockClient()
	testCases := []struct {
		desc         string
		symbol       string
		requests     map[string]http.HandlerFunc
		frequency    PriceFrequency
		firstPrice   Price
		pricesLength int
		err          error
	}{
		{
			desc:   "Hourly",
			symbol: "00006",
			requests: map[string]http.HandlerFunc{
				"GET-http://www.aastocks.com/en/stocks/quote/detail-quote.aspx?symbol=00006":                                                             serveFile("testdata/detail_quote.html"),
				"GET-http://chartdata1.internet.aastocks.com/servlet/iDataServlet/getdaily?id=00006.HK&type=24&market=1&level=1&period=23&encoding=utf8": serveFile("testdata/historical_price_00006_hourly.html"),
			},
			frequency:    Hourly,
			pricesLength: 370,
			firstPrice: Price{
				Time:  time.Date(time.Now().Year(), time.July, 31, 10, 0, 0, 0, time.UTC),
				Open:  42.75,
				High:  43.15,
				Low:   42.7,
				Close: 43.05,
			},
		},
		{
			desc:   "Daily",
			symbol: "00006",
			requests: map[string]http.HandlerFunc{
				"GET-http://www.aastocks.com/en/stocks/quote/detail-quote.aspx?symbol=00006":                                                             serveFile("testdata/detail_quote.html"),
				"GET-http://chartdata1.internet.aastocks.com/servlet/iDataServlet/getdaily?id=00006.HK&type=24&market=1&level=1&period=56&encoding=utf8": serveFile("testdata/historical_price_00006_daily.html"),
			},
			frequency:    Daily,
			pricesLength: 1482,
			firstPrice: Price{
				Time:  time.Date(2015, time.August, 26, 0, 0, 0, 0, time.UTC),
				Open:  45.48,
				High:  48.03,
				Low:   45.23,
				Close: 47.23,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mock.set(tC.requests)

			checkErrorFunc(t, tC.err, func() error {
				quote, err := Get(tC.symbol, WithClient(mock.client))
				if err != nil {
					return err
				}
				prices, err := quote.HistoricalPrices(tC.frequency)
				if err != nil {
					return err
				}
				diff := cmp.Diff(tC.pricesLength, len(prices))
				if diff != "" {
					t.Fatalf(diff)
				}
				if tC.pricesLength > 0 {
					diff = cmp.Diff(tC.firstPrice, prices[0])
					if diff != "" {
						t.Fatalf(diff)
					}
				}
				return nil
			})
		})
	}
}
