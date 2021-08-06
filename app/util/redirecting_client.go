package util

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// RedirectingTestClient makes an http server with the given handler function,
// sets cleanup for closing test server and makes an http client that redirects
// all requests to the test server.
// Used for testing, for example see notify/github_test
func RedirectingTestClient(t *testing.T, fn http.HandlerFunc) *http.Client {
	ts := httptest.NewServer(fn)
	t.Cleanup(ts.Close)

	tsURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: roundTripperFn(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = tsURL.Scheme
			req.URL.Host = tsURL.Host
			return http.DefaultTransport.RoundTrip(req)
		}),
	}
}

type roundTripperFn func(req *http.Request) (*http.Response, error)

func (f roundTripperFn) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
