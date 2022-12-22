package server

import (
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp.StatusCode, string(respBody)
}

func TestRouter(t *testing.T) {
	r := NewRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, _ := testRequest(t, ts, "POST", "/update/counter/metric_name/43")
	require.Equal(t, http.StatusOK, statusCode)

	statusCode, body := testRequest(t, ts, "GET", "/value/counter/metric_name")
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, "43", body)

	statusCode, _ = testRequest(t, ts, "GET", "/value/counter/unknown_metric")
	require.Equal(t, http.StatusNotFound, statusCode)

	statusCode, body = testRequest(t, ts, "GET", "/")
	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, "<html><ol></ol><ol><li>metric_name:43</li></ol></html>", body)
}
