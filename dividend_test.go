package aastocks

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestDividends(t *testing.T) {
	mock := mockClient()
	testCases := []struct {
		symbol    string
		requests  map[string]http.HandlerFunc
		dividends []Dividend
		err       error
	}{
		{
			symbol: "00006",
			requests: map[string]http.HandlerFunc{
				"GET-http://www.aastocks.com/en/stocks/quote/detail-quote.aspx?symbol=00006": serveFile("testdata/detail_quote.html"),
				"GET-http://www.aastocks.com/en/stocks/analysis/dividend.aspx?symbol=00006":  serveFile("testdata/dividend.html"),
			},
			dividends: []Dividend{
				{
					AnnounceDate: time.Date(2020, time.August, 5, 0, 0, 0, 0, time.UTC),
					YearEnded:    time.Date(2020, time.December, 1, 0, 0, 0, 0, time.UTC),
					Event:        "Interim",
					Particular:   "D:HKD 0.7700",
					Type:         "Cash",
					ExDate:       time.Date(2020, time.September, 3, 0, 0, 0, 0, time.UTC),
					PayableDate:  time.Date(2020, time.September, 15, 0, 0, 0, 0, time.UTC),
				},
				{
					AnnounceDate: time.Date(2020, time.March, 18, 0, 0, 0, 0, time.UTC),
					YearEnded:    time.Date(2019, time.December, 1, 0, 0, 0, 0, time.UTC),
					Event:        "Final",
					Particular:   "D:HKD 2.0300",
					Type:         "Cash",
					ExDate:       time.Date(2020, time.May, 18, 0, 0, 0, 0, time.UTC),
					PayableDate:  time.Date(2020, time.May, 28, 0, 0, 0, 0, time.UTC),
				},
				{
					AnnounceDate: time.Date(2013, time.September, 27, 0, 0, 0, 0, time.UTC),
					YearEnded:    time.Time{},
					Event:        "Special",
					Particular:   "Preferential Offer: 1 HK Electric Investments and HK Electric Investments Limited Share Stapled unit offer price HKD 5.4500 for every 4 Shares held",
					Type:         "-",
					ExDate:       time.Date(2014, time.January, 8, 0, 0, 0, 0, time.UTC),
					PayableDate:  time.Time{},
				},
				{
					AnnounceDate: time.Date(2013, time.July, 24, 0, 0, 0, 0, time.UTC),
					YearEnded:    time.Date(2013, time.December, 1, 0, 0, 0, 0, time.UTC),
					Event:        "Interim",
					Particular:   "D:HKD 0.6500",
					Type:         "Cash",
					ExDate:       time.Date(2013, time.August, 23, 0, 0, 0, 0, time.UTC),
					PayableDate:  time.Date(2013, time.September, 4, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			symbol: "09923",
			requests: map[string]http.HandlerFunc{
				"GET-http://www.aastocks.com/en/stocks/quote/detail-quote.aspx?symbol=09923": serveFile("testdata/detail_quote_new.html"),
				"GET-http://www.aastocks.com/en/stocks/analysis/dividend.aspx?symbol=09923":  serveFile("testdata/divdend_empty.html"),
			},
			dividends: []Dividend{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.symbol, func(t *testing.T) {
			mock.set(tC.requests)

			checkErrorFunc(t, tC.err, func() error {
				quote, err := Get(tC.symbol, WithClient(mock.client))
				if err != nil {
					return err
				}
				dividends, err := quote.Dividends()
				if err != nil {
					return err
				}
				diff := cmp.Diff(tC.dividends, dividends)
				if diff != "" {
					t.Fatalf(diff)
				}
				return nil
			})
		})
	}
}
