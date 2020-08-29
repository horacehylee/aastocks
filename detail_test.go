package aastocks

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestGetQuote(t *testing.T) {
	mock := mockClient()
	testCases := []struct {
		symbol   string
		requests map[string]http.HandlerFunc
		quote    Quote
		err      error
	}{
		{
			symbol: "00006",
			requests: map[string]http.HandlerFunc{
				"GET-http://www.aastocks.com/en/stocks/quote/detail-quote.aspx?symbol=00006": serveFile("testdata/detail_quote.html"),
			},
			quote: Quote{
				Symbol:     "00006",
				Name:       "POWER ASSETS",
				Price:      44.65,
				Yield:      0.06271,
				PeRatio:    13.368,
				PbRatio:    1.115,
				Eps:        3.34,
				Lots:       500,
				UpdateTime: time.Date(2020, time.August, 25, 21, 18, 38, 0, time.UTC),
			},
		},
		{
			symbol: "09923",
			requests: map[string]http.HandlerFunc{
				"GET-http://www.aastocks.com/en/stocks/quote/detail-quote.aspx?symbol=09923": serveFile("testdata/detail_quote_new.html"),
			},
			quote: Quote{
				Symbol:     "09923",
				Name:       "YEAHKA",
				Price:      59.6,
				Yield:      0,
				PeRatio:    0,
				PbRatio:    0,
				Eps:        0,
				Lots:       400,
				UpdateTime: time.Date(2020, time.August, 26, 2, 50, 43, 0, time.UTC),
			},
		},
		{
			symbol: "151511",
			requests: map[string]http.HandlerFunc{
				"GET-http://www.aastocks.com/en/stocks/quote/detail-quote.aspx?symbol=151511": serveFile("testdata/detail_quote_not_found.html"),
			},
			err: errors.New("Symbol cannot be found: 151511"),
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
				diff := cmp.Diff(tC.quote, *quote, cmp.AllowUnexported(Quote{}), cmpopts.IgnoreTypes(&http.Client{}))
				if diff != "" {
					t.Fatalf(diff)
				}
				return nil
			})
		})
	}
}

func TestDetailsParseFunc(t *testing.T) {
	testCases := []struct {
		desc      string
		content   string
		parseFunc func(*Quote, *goquery.Document) func() error
		quote     Quote
		err       error
	}{
		{
			desc:      "Name",
			content:   `<span id="cp_ucStockBar_litInd_StockName" title="POWER ASSETS">POWER ASSETS</span></Label>`,
			parseFunc: name,
			quote: Quote{
				Name: "POWER ASSETS",
			},
		},
		{
			desc:      "Name/NotFound",
			content:   `<span id="xxx" title="POWER ASSETS">POWER ASSETS</span></Label>`,
			parseFunc: name,
			err:       fmt.Errorf("Name cannot be found"),
		},
		{
			desc:      "UpdateTime",
			content:   `<script>var ServerDate = new Date('2020-08-29T00:55:31');</script>`,
			parseFunc: updateTime,
			quote: Quote{
				UpdateTime: time.Date(2020, 8, 29, 0, 55, 31, 0, time.UTC),
			},
		},
		{
			desc:      "UpdateTime/NotFound",
			content:   `<script>var ServerDate2 = new Date('2020-08-29T00:55:31');</script>`,
			parseFunc: updateTime,
			err:       fmt.Errorf("Server date cannot be found"),
		},
		{
			desc:      "UpdateTime/NotParse",
			content:   `<script>var ServerDate = new Date('');</script>`,
			parseFunc: updateTime,
			err:       fmt.Errorf(`Server date failed to be parsed: parsing time "" as "2006-01-02T15:04:05": cannot parse "" as "2006"`),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			checkErrorFunc(t, tC.err, func() error {
				var err error

				doc, err := goquery.NewDocumentFromReader(strings.NewReader(tC.content))
				if err != nil {
					return err
				}

				q := Quote{}
				err = tC.parseFunc(&q, doc)()
				if err != nil {
					return err
				}

				diff := cmp.Diff(tC.quote, q, cmpopts.IgnoreUnexported(Quote{}))
				if diff != "" {
					t.Fatalf(diff)
				}
				return err
			})
		})
	}
}
