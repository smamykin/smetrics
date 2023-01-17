package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"testing"
)

func TestGetMetricFromRequest(t *testing.T) {
	type testCase struct {
		request       *http.Request
		updateHandler *UpdateHandler
		expected      Metrics
	}
	tests := map[string]testCase{}

	expectedValue := math.Round(rand.Float64()*100) / 100
	expectedDelta := rand.Int63()

	expected := Metrics{
		MType: metricTypeGauge,
		ID:    "metric_name3",
		Value: &expectedValue,
	}
	tests["json. update for gauge"] = testCase{
		request:       newJSONRequest(expected),
		updateHandler: NewUpdateHandler(RepositoryMock{}, ParametersBagMock{}),
		expected:      expected,
	}
	tests["text. update for gauge"] = testCase{
		request: &http.Request{},
		updateHandler: NewUpdateHandler(
			RepositoryMock{},
			ParametersBagMock{
				parameters: map[string]string{
					paramNameMetricType:  expected.MType,
					paramNameMetricName:  expected.ID,
					paramNameMetricValue: fmt.Sprint(expectedValue),
				},
			},
		),
		expected: expected,
	}

	expected = Metrics{
		MType: metricTypeCounter,
		ID:    "metric_name4",
		Delta: &expectedDelta,
	}
	tests["json. update for counter"] = testCase{
		request:       newJSONRequest(expected),
		updateHandler: NewUpdateHandler(RepositoryMock{}, ParametersBagMock{}),
		expected:      expected,
	}
	tests["text. update for counter"] = testCase{
		request: &http.Request{},
		updateHandler: NewUpdateHandler(
			RepositoryMock{},
			ParametersBagMock{
				parameters: map[string]string{
					paramNameMetricType:  expected.MType,
					paramNameMetricName:  expected.ID,
					paramNameMetricValue: fmt.Sprint(expectedDelta),
				},
			},
		),
		expected: expected,
	}

	expected = Metrics{
		MType: metricTypeGauge,
		ID:    "metric_name3",
	}
	tests["json. get for gauge"] = testCase{
		request:       newJSONRequest(expected),
		updateHandler: NewUpdateHandler(RepositoryMock{}, ParametersBagMock{}),
		expected:      expected,
	}
	tests["text. get for gauge"] = testCase{
		request: &http.Request{},
		updateHandler: NewUpdateHandler(
			RepositoryMock{},
			ParametersBagMock{
				parameters: map[string]string{
					paramNameMetricType: expected.MType,
					paramNameMetricName: expected.ID,
				},
			},
		),
		expected: expected,
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual, _ := tt.updateHandler.getMetricFromRequest(tt.request)
			assert.Equal(t, tt.expected, actual)
		})
	}

}

func newJSONRequest(expected Metrics) *http.Request {
	body, _ := json.Marshal(expected)
	return &http.Request{
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(string(body))),
	}
}

type ParametersBagMock struct {
	parameters map[string]string
}

func (p ParametersBagMock) GetURLParam(r *http.Request, key string) string {
	return p.parameters[key]
}

type RepositoryMock struct{}

func (r RepositoryMock) UpsertGauge(metric GaugeMetric) error {
	panic("must not be invoked")
}
func (r RepositoryMock) UpsertCounter(metric CounterMetric) error {
	panic("must not be invoked")
}
func (r RepositoryMock) GetGauge(name string) (float64, error) {
	panic("must not be invoked")
}
func (r RepositoryMock) GetCounter(name string) (int64, error) {
	panic("must not be invoked")
}
func (r RepositoryMock) GetAllGauge() ([]GaugeMetric, error) {
	panic("must not be invoked")
}
func (r RepositoryMock) GetAllCounters() ([]CounterMetric, error) {
	panic("must not be invoked")
}
