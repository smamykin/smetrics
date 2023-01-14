package handlers

import (
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
		updateHandler UpdateHandler
		expected      Metrics
	}
	tests := map[string]testCase{}

	expectedValue := math.Round(rand.Float64()*100) / 100
	expected := Metrics{
		MType: metricTypeGauge,
		ID:    "metric_name3",
		Value: &expectedValue,
	}
	tests["json for gauge"] = testCase{
		request: newJsonRequest(expected),
		updateHandler: UpdateHandler{
			ParametersBag: ParametersBagMock{},
		},
		expected: expected,
	}
	tests["text for gauge"] = testCase{
		request: &http.Request{},
		updateHandler: UpdateHandler{
			ParametersBag: ParametersBagMock{
				parameters: map[string]string{
					paramNameMetricType:  expected.MType,
					paramNameMetricName:  expected.ID,
					paramNameMetricValue: fmt.Sprint(expectedValue),
				},
			},
		},
		expected: expected,
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual, _ := tt.updateHandler.getMetricFromRequest(tt.request)
			assert.Equal(t, expected, actual)
		})
	}

}

func newJsonRequest(expected Metrics) *http.Request {
	body := fmt.Sprintf(`{"id":"%s", "type":"%s", "value":%f}`, expected.ID, expected.MType, *expected.Value)
	return &http.Request{
		Header: map[string][]string{
			"Content-Type": {"application/json"},
			//"Accept":       {"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}
}

type ParametersBagMock struct {
	parameters map[string]string
}

func (p ParametersBagMock) GetUrlParam(r *http.Request, key string) string {
	return p.parameters[key]
}
