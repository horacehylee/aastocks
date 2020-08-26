package aastocks

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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
				Symbol:  "00006",
				Name:    "POWER ASSETS",
				Price:   44.65,
				Yield:   0.06271,
				PeRatio: 13.368,
				PbRatio: 1.115,
				Eps:     3.34,
				Lots:    500,
			},
		},
		{
			symbol: "09923",
			requests: map[string]http.HandlerFunc{
				"GET-http://www.aastocks.com/en/stocks/quote/detail-quote.aspx?symbol=09923": serveFile("testdata/detail_quote_new.html"),
			},
			quote: Quote{
				Symbol:  "09923",
				Name:    "YEAHKA",
				Price:   59.6,
				Yield:   0,
				PeRatio: 0,
				PbRatio: 0,
				Eps:     0,
				Lots:    400,
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

func TestQuoteDividend(t *testing.T) {
	mock := mockClient()
	testCases := []struct {
		symbol    string
		requests  map[string]http.HandlerFunc
		dividends []*Dividend
		err       error
	}{
		{
			symbol: "00006",
			requests: map[string]http.HandlerFunc{
				"GET-http://www.aastocks.com/en/stocks/quote/detail-quote.aspx?symbol=00006": serveFile("testdata/detail_quote.html"),
				"GET-http://www.aastocks.com/en/stocks/analysis/dividend.aspx?symbol=00006":  serveFile("testdata/dividend.html"),
			},
			dividends: []*Dividend{
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
			dividends: []*Dividend{},
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

type mockHTTPClient struct {
	client   *http.Client
	requests map[string]http.HandlerFunc
}

func mockClient() *mockHTTPClient {
	mock := &mockHTTPClient{}
	client := &http.Client{
		Transport: mock,
	}
	mock.client = client
	mock.requests = make(map[string]http.HandlerFunc)
	return mock
}

func (m *mockHTTPClient) clear() {
	m.requests = make(map[string]http.HandlerFunc)
}

func (m *mockHTTPClient) add(method string, url string, handlerFunc http.HandlerFunc) {
	key := fmt.Sprintf("%s-%s", method, url)
	m.requests[key] = handlerFunc
}

func (m *mockHTTPClient) set(mapping map[string]http.HandlerFunc) {
	m.requests = mapping
}

func (m *mockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	key := fmt.Sprintf("%s-%s", req.Method, req.URL.String())
	handler, ok := m.requests[key]
	if !ok {
		return nil, fmt.Errorf("Handler not found for %s", key)
	}
	recorder := httptest.NewRecorder()
	handler(recorder, req)
	return recorder.Result(), nil
}

func serveFile(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(name)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		b, err := ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}
		_, err = w.Write(b)
		if err != nil {
			panic(err)
		}
	}
}

var equateErrorMessage = cmp.Comparer(func(x, y error) bool {
	if x == nil || y == nil {
		return x == nil && y == nil
	}
	return x.Error() == y.Error()
})

type errorFunc func() error

func checkErrorFunc(t *testing.T, expectedError error, f errorFunc) {
	err := f()

	if err != nil {
		if expectedError != nil {
			diff := cmp.Diff(expectedError, err, equateErrorMessage)
			if diff != "" {
				t.Fatalf(diff)
			}
		} else {
			t.Fatalf("No error should be expected: %v", err)
		}
	} else {
		if expectedError != nil {
			t.Fatalf("error is expected (%v), but got none", expectedError)
		}
	}
}
