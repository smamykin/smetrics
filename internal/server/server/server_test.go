package server

import (
	"bytes"
	"compress/gzip"
	"github.com/smamykin/smetrics/internal/server/handlers"
	"github.com/smamykin/smetrics/internal/server/storage"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type requestDefinition struct {
	method          string
	url             string
	body            string
	contentType     string
	contentEncoding string
}

func testRequest(t *testing.T, ts *httptest.Server, request requestDefinition) (status int, responseContentType string, responseBody string) {
	req, err := http.NewRequest(request.method, ts.URL+request.url, strings.NewReader(request.body))
	req.Header.Set("Accept", request.contentType)
	req.Header.Set("Content-Type", request.contentType)
	req.Header.Set("Content-Encoding", request.contentEncoding)

	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp.StatusCode, resp.Header.Get("Content-Type"), string(respBody)
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

	type expected struct {
		statusCode   int
		body         string
		contentType  string
		counterStore map[string]handlers.CounterMetric
		gaugeStore   map[string]handlers.GaugeMetric
	}
	type testCase struct {
		name     string
		requests []requestDefinition
		expected
	}

	tests := map[string]testCase{
		"the first update of the counter": {
			requests: []requestDefinition{{method: http.MethodPost, url: "/update/counter/metric_name/43"}},
			expected: expected{
				contentType:  "text/plain",
				statusCode:   http.StatusOK,
				body:         "43",
				counterStore: map[string]handlers.CounterMetric{"metric_name": {Value: 43, Name: "metric_name"}},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		"update counter": {
			requests: []requestDefinition{
				{method: http.MethodPost, url: "/update/counter/metric_name/43"},
				{method: http.MethodPost, url: "/update/counter/metric_name/7"},
			},
			expected: expected{
				contentType:  "text/plain",
				statusCode:   http.StatusOK,
				body:         "50",
				counterStore: map[string]handlers.CounterMetric{"metric_name": {Value: 50, Name: "metric_name"}},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		"invalid counter value": {
			requests: []requestDefinition{
				{method: http.MethodPost, url: "/update/counter/metric_name/none"},
			},
			expected: expected{
				contentType:  "text/plain; charset=utf-8",
				statusCode:   http.StatusBadRequest,
				body:         "strconv.ParseInt: parsing \"none\": invalid syntax\n",
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		"the first update of gauge": {
			requests: []requestDefinition{
				{method: http.MethodPost, url: "/update/gauge/metric_name/43.332"},
			},
			expected: expected{
				contentType:  "text/plain",
				statusCode:   http.StatusOK,
				body:         "43.332",
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{"metric_name": {Value: 43.332, Name: "metric_name"}},
			},
		},
		"update of gauge": {
			requests: []requestDefinition{
				{method: http.MethodPost, url: "/update/gauge/metric_name/43.332"},
				{method: http.MethodPost, url: "/update/gauge/metric_name/50.111"},
			},
			expected: expected{
				contentType:  "text/plain",
				statusCode:   http.StatusOK,
				body:         "50.111",
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{"metric_name": {Value: 50.111, Name: "metric_name"}},
			},
		},
		"method not post for update endpoint": {
			requests: []requestDefinition{
				{method: http.MethodGet, url: "/update/gauge/metric_name/43.332"},
			},
			expected: expected{
				contentType:  "",
				statusCode:   http.StatusMethodNotAllowed,
				body:         "",
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		"unknown type for update": {
			requests: []requestDefinition{
				{method: http.MethodPost, url: "/update/unknown_type/metric_name/43.332"},
			},
			expected: expected{
				contentType:  "text/plain; charset=utf-8",
				statusCode:   http.StatusNotImplemented,
				body:         "unknown metric type\n",
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		"get gauge": {
			requests: []requestDefinition{
				{method: http.MethodPost, url: "/update/gauge/metric_name/50.111123"},
				{method: http.MethodGet, url: "/value/gauge/metric_name"},
			},
			expected: expected{
				contentType:  "text/plain",
				statusCode:   http.StatusOK,
				body:         "50.111",
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{"metric_name": {Value: 50.111123, Name: "metric_name"}},
			},
		},
		"get counter": {
			requests: []requestDefinition{
				{method: http.MethodPost, url: "/update/counter/metric_name/43"},
				{method: http.MethodGet, url: "/value/counter/metric_name"},
			},
			expected: expected{
				contentType:  "text/plain",
				statusCode:   http.StatusOK,
				body:         "43",
				counterStore: map[string]handlers.CounterMetric{"metric_name": {Value: 43, Name: "metric_name"}},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		"unknown type": {
			requests: []requestDefinition{
				{method: http.MethodGet, url: "/value/counter/unknown_metric"},
			},
			expected: expected{
				contentType:  "text/plain; charset=utf-8",
				statusCode:   http.StatusNotFound,
				body:         "metric not found\n",
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		"list": {
			requests: []requestDefinition{
				{method: http.MethodPost, url: "/update/counter/metric_name1/43"},
				{method: http.MethodPost, url: "/update/gauge/metric_name2/50.111"},
				{method: http.MethodGet, url: "/"},
			},
			expected: expected{
				contentType:  "text/html",
				statusCode:   http.StatusOK,
				body:         expectedListBody,
				counterStore: map[string]handlers.CounterMetric{"metric_name1": {Value: 43, Name: "metric_name1"}},
				gaugeStore:   map[string]handlers.GaugeMetric{"metric_name2": {Value: 50.111, Name: "metric_name2"}},
			},
		},
		"JSON-API update gauge": {
			requests: []requestDefinition{
				{method: http.MethodPost, url: "/update/", body: `{"id":"metric_name3", "type":"gauge", "value":11.12}`, contentType: "application/json"},
				{method: http.MethodPost, url: "/update/", body: `{"id":"metric_name3", "type":"gauge", "value":13.14}`, contentType: "application/json"},
			},
			expected: expected{
				contentType:  "application/json",
				statusCode:   http.StatusOK,
				body:         `{"id":"metric_name3","type":"gauge","value":13.14}`,
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{"metric_name3": {Value: 13.14, Name: "metric_name3"}},
			},
		},
		"JSON-API update counter": {
			requests: []requestDefinition{
				{method: http.MethodPost, url: "/update/", body: `{"id":"metric_name3", "type":"counter", "delta":11}`, contentType: "application/json"},
				{method: http.MethodPost, url: "/update/", body: `{"id":"metric_name3", "type":"counter", "delta":13}`, contentType: "application/json"},
			},
			expected: expected{
				contentType:  "application/json",
				statusCode:   http.StatusOK,
				body:         `{"id":"metric_name3","type":"counter","delta":24}`,
				counterStore: map[string]handlers.CounterMetric{"metric_name3": {Value: 24, Name: "metric_name3"}},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
		"JSON-API get gauge": {
			requests: []requestDefinition{
				{method: http.MethodPost, url: "/update/", body: `{"id":"metric_name3", "type":"gauge", "value":11.12}`, contentType: "application/json"},
				{method: http.MethodPost, url: "/value/", body: `{"id":"metric_name3", "type":"gauge"}`, contentType: "application/json"},
			},
			expected: expected{
				contentType:  "application/json",
				statusCode:   http.StatusOK,
				body:         `{"id":"metric_name3","type":"gauge","value":11.12}`,
				counterStore: map[string]handlers.CounterMetric{},
				gaugeStore:   map[string]handlers.GaugeMetric{"metric_name3": {Value: 11.12, Name: "metric_name3"}},
			},
		},
		"JSON-API get counter": {
			requests: []requestDefinition{
				{method: http.MethodPost, url: "/update/", body: `{"id":"metric_name3", "type":"counter", "delta":11}`, contentType: "application/json"},
				{method: http.MethodPost, url: "/value/", body: `{"id":"metric_name3", "type":"counter"}`, contentType: "application/json"},
			},
			expected: expected{
				contentType:  "application/json",
				statusCode:   http.StatusOK,
				body:         `{"id":"metric_name3","type":"counter","delta":11}`,
				counterStore: map[string]handlers.CounterMetric{"metric_name3": {Value: 11, Name: "metric_name3"}},
				gaugeStore:   map[string]handlers.GaugeMetric{},
			},
		},
	}

	tests["gzip"] = testCase{
		requests: []requestDefinition{
			{method: http.MethodPost, url: "/update/", body: compress(t, `{"id":"metric_name3", "type":"counter", "delta":11}`), contentType: "application/json", contentEncoding: "gzip"},
			{method: http.MethodPost, url: "/value/", body: compress(t, `{"id":"metric_name3", "type":"counter"}`), contentType: "application/json", contentEncoding: "gzip"},
		},
		expected: expected{
			contentType:  "application/json",
			statusCode:   http.StatusOK,
			body:         `{"id":"metric_name3","type":"counter","delta":11}`,
			counterStore: map[string]handlers.CounterMetric{"metric_name3": {Value: 11, Name: "metric_name3"}},
			gaugeStore:   map[string]handlers.GaugeMetric{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repository := storage.NewMemStorageDefault()
			r := NewRouter(repository)
			ts := httptest.NewServer(r)
			defer ts.Close()

			var statusCode int
			var body string
			var contentType string
			for _, request := range tt.requests {
				statusCode, contentType, body = testRequest(t, ts, request)
			}
			require.Equal(t, tt.expected.statusCode, statusCode)
			require.Equal(t, tt.expected.body, body)
			require.Equal(t, tt.expected.gaugeStore, repository.GaugeStore())
			require.Equal(t, tt.expected.counterStore, repository.CounterStore())
			require.Equal(t, tt.expected.contentType, contentType)
		})
	}
}

func compress(t *testing.T, s string) string {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(s)); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	gzipBodyUpdate := b.String()
	return gzipBodyUpdate
}
