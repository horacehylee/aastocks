package aastocks

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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
				symbol:  "00006",
				name:    "POWER ASSETS",
				price:   44.65,
				yield:   0.06271,
				peRatio: 13.368,
				pbRatio: 1.115,
				eps:     3.34,
				lots:    500,
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

			quote, err := Get(tC.symbol, WithClient(mock.client))
			if tC.err != nil {
				if err == nil {
					t.Fatalf("error is expected (%v), but got none", tC.err)
				}
				diff := cmp.Diff(tC.err, err, equateErrorMessage)
				if diff != "" {
					t.Fatalf(diff)
				}
				return
			}

			if err != nil {
				t.Fatalf("No error should be expected: %v", err)
			}
			diff := cmp.Diff(tC.quote, *quote, cmp.AllowUnexported(Quote{}), cmpopts.IgnoreTypes(&http.Client{}))
			if diff != "" {
				t.Fatalf(diff)
			}
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
