package server

import (
	"github.com/smamykin/smetrics/internal/server/handlers"
	"github.com/smamykin/smetrics/internal/server/storage"
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

const expectedListBody = `
<html>
    <ol>
        <li>metric_name2:50.111</li>
    </ol>
    <ol>
        <li>metric_name1:43</li>
    </ol>
</html>`

func TestRouter(t *testing.T) {
	type request struct {
		method string
		url    string
	}
	type expected struct {
		statusCode   int
		body         string
		counterStore map[string]handlers.CounterMetric
		gaugeStore   map[string]handlers.GaugeMetric
	}
	tests := []struct {
		name     string
		requests []request
		expected
	}{
		{
			name:     "the first update of the counter",
			requests: []request{{method: http.MethodPost, url: "/update/counter/metric_name/43"}},
			expected: expected{
				statusCode:   http.StatusOK,
				body:         "",
				counterStore: map[string]handlers.CounterMetric{"metric_name": {Value: 43, Name: "metric_name"}},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		{
			name: "update counter",
			requests: []request{
				{method: http.MethodPost, url: "/update/counter/metric_name/43"},
				{method: http.MethodPost, url: "/update/counter/metric_name/7"},
			},
			expected: expected{
				statusCode:   http.StatusOK,
				body:         "",
				counterStore: map[string]handlers.CounterMetric{"metric_name": {Value: 50, Name: "metric_name"}},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		{
			name: "the first update of gauge",
			requests: []request{
				{method: http.MethodPost, url: "/update/gauge/metric_name/43.332"},
			},
			expected: expected{
				statusCode:   http.StatusOK,
				body:         "",
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{"metric_name": {Value: 43.332, Name: "metric_name"}},
			},
		},
		{
			name: "update of gauge",
			requests: []request{
				{method: http.MethodPost, url: "/update/gauge/metric_name/43.332"},
				{method: http.MethodPost, url: "/update/gauge/metric_name/50.111"},
			},
			expected: expected{
				statusCode:   http.StatusOK,
				body:         "",
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{"metric_name": {Value: 50.111, Name: "metric_name"}},
			},
		},
		{
			name: "method not post for update endpoint",
			requests: []request{
				{method: http.MethodGet, url: "/update/gauge/metric_name/43.332"},
			},
			expected: expected{
				statusCode:   http.StatusMethodNotAllowed,
				body:         "",
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		{
			name: "unknown type for update",
			requests: []request{
				{method: http.MethodPost, url: "/update/unknown_type/metric_name/43.332"},
			},
			expected: expected{
				statusCode:   http.StatusNotImplemented,
				body:         "metric type is incorrect\n",
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		{
			name: "get gauge",
			requests: []request{
				{method: http.MethodPost, url: "/update/gauge/metric_name/50.111123"},
				{method: http.MethodGet, url: "/value/gauge/metric_name"},
			},
			expected: expected{
				statusCode:   http.StatusOK,
				body:         "50.111",
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{"metric_name": {Value: 50.111123, Name: "metric_name"}},
			},
		},
		{
			name: "get counter",
			requests: []request{
				{method: http.MethodPost, url: "/update/counter/metric_name/43"},
				{method: http.MethodGet, url: "/value/counter/metric_name"},
			},
			expected: expected{
				statusCode:   http.StatusOK,
				body:         "43",
				counterStore: map[string]handlers.CounterMetric{"metric_name": {Value: 43, Name: "metric_name"}},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		{
			name: "unknown type",
			requests: []request{
				{method: http.MethodGet, url: "/value/counter/unknown_metric"},
			},
			expected: expected{
				statusCode:   http.StatusNotFound,
				body:         "metric not found\n",
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		{
			name: "list",
			requests: []request{
				{method: http.MethodPost, url: "/update/counter/metric_name1/43"},
				{method: http.MethodPost, url: "/update/gauge/metric_name2/50.111"},
				{method: http.MethodGet, url: "/"},
			},
			expected: expected{
				statusCode:   http.StatusOK,
				body:         expectedListBody,
				counterStore: map[string]handlers.CounterMetric{"metric_name1": {Value: 43, Name: "metric_name1"}},
				gaugeStore:   map[string]handlers.GaugeMetric{"metric_name2": {Value: 50.111, Name: "metric_name2"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := storage.NewMemStorage()
			r := NewRouter(repository)
			ts := httptest.NewServer(r)
			defer ts.Close()

			var statusCode int
			var body string
			for _, request := range tt.requests {
				statusCode, body = testRequest(t, ts, request.method, request.url)
			}
			require.Equal(t, tt.expected.statusCode, statusCode)
			require.Equal(t, tt.expected.body, body)
			require.Equal(t, tt.expected.gaugeStore, repository.GaugeStore())
			require.Equal(t, tt.expected.counterStore, repository.CounterStore())
		})
	}
}
