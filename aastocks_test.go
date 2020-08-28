package aastocks

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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

func serveAll(handlers ...http.HandlerFunc) http.HandlerFunc {
	i := 0
	return func(w http.ResponseWriter, r *http.Request) {
		if i >= len(handlers) {
			panic(fmt.Errorf("handler for %d times should be defined", i))
		}
		handlers[i](w, r)
		i++
	}
}

func serveError(err error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
