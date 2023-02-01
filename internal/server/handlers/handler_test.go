package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/smamykin/smetrics/internal/utils"
	"github.com/stretchr/testify/require"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"testing"
)

func TestGetMetricFromRequest(t *testing.T) {
	type testCase struct {
		request        *http.Request
		updateHandler  *UpdateHandler
		expectedMetric Metrics
		expectedError  error
	}
	tests := map[string]testCase{}

	expectedValue := math.Round(rand.Float64()*100) / 100
	expectedDelta := rand.Int63()

	expected := Metrics{
		MType: MetricTypeGauge,
		ID:    "metric_name3",
		Value: &expectedValue,
	}
	tests["json. update for gauge"] = testCase{
		request:        newJSONRequest(expected),
		updateHandler:  NewUpdateHandler(RepositoryMock{}, ParametersBagMock{}, nil),
		expectedMetric: expected,
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
			nil,
		),
		expectedMetric: expected,
	}

	expected = Metrics{
		MType: MetricTypeCounter,
		ID:    "metric_name4",
		Delta: &expectedDelta,
	}
	tests["json. update for counter"] = testCase{
		request:        newJSONRequest(expected),
		updateHandler:  NewUpdateHandler(RepositoryMock{}, ParametersBagMock{}, nil),
		expectedMetric: expected,
	}

	h := utils.NewHashGenerator("secret key")
	stringToHash := fmt.Sprintf("metric_name4:gauge:%f", expectedValue)
	sign, err := h.Generate(stringToHash)
	require.Nil(t, err)
	expected = Metrics{
		MType: MetricTypeGauge,
		ID:    "metric_name4",
		Value: &expectedValue,
		Hash:  sign,
	}
	tests["with sign. gauge counter. success"] = testCase{
		request:        newJSONRequest(expected),
		updateHandler:  NewUpdateHandler(RepositoryMock{}, ParametersBagMock{}, h),
		expectedMetric: expected,
	}

	wrongSign, err := h.Generate("some wrong string")
	require.Nil(t, err)
	expected = Metrics{
		MType: MetricTypeGauge,
		ID:    "metric_name4",
		Value: &expectedValue,
		Hash:  wrongSign,
	}
	tests["with sign. update gauge. fail because of wrong hash"] = testCase{
		request:        newJSONRequest(expected),
		updateHandler:  NewUpdateHandler(RepositoryMock{}, ParametersBagMock{}, h),
		expectedMetric: expected,
		expectedError: govalidator.Errors{govalidator.Errors{
			govalidator.Error{Name: "hash", Err: errors.New("hash is not correct"), Validator: "customHash", CustomErrorMessageExists: true},
		}},
	}
	expected = Metrics{
		MType: MetricTypeGauge,
		ID:    "metric_name4",
		Value: &expectedValue,
	}
	tests["with sign. update gauge. fail because of empty hash"] = testCase{
		request:        newJSONRequest(expected),
		updateHandler:  NewUpdateHandler(RepositoryMock{}, ParametersBagMock{}, h),
		expectedMetric: expected,
		expectedError:  errors.New("hash is not correct"),
	}

	sign, err = h.Generate(fmt.Sprintf("metric_name4:counter:%d", expectedDelta))
	require.Nil(t, err)
	expected = Metrics{
		MType: MetricTypeCounter,
		ID:    "metric_name4",
		Delta: &expectedDelta,
		Hash:  sign,
	}
	tests["with sign. update counter. success"] = testCase{
		request:        newJSONRequest(expected),
		updateHandler:  NewUpdateHandler(RepositoryMock{}, ParametersBagMock{}, h),
		expectedMetric: expected,
	}

	wrongSign, err = h.Generate("some wrong string")
	require.Nil(t, err)
	expected = Metrics{
		MType: MetricTypeCounter,
		ID:    "metric_name4",
		Delta: &expectedDelta,
		Hash:  wrongSign,
	}
	tests["with sign. update counter. fail because of wrong hash"] = testCase{
		request:        newJSONRequest(expected),
		updateHandler:  NewUpdateHandler(RepositoryMock{}, ParametersBagMock{}, h),
		expectedMetric: expected,
		expectedError: govalidator.Errors{govalidator.Errors{
			govalidator.Error{Name: "hash", Err: errors.New("hash is not correct"), Validator: "customHash", CustomErrorMessageExists: true},
		}},
	}
	expected = Metrics{
		MType: MetricTypeCounter,
		ID:    "metric_name4",
		Delta: &expectedDelta,
	}
	tests["with sign. update counter. fail because of empty hash"] = testCase{
		request:        newJSONRequest(expected),
		updateHandler:  NewUpdateHandler(RepositoryMock{}, ParametersBagMock{}, h),
		expectedMetric: expected,
		expectedError:  errors.New("hash is not correct"),
	}

	expected = Metrics{
		MType: MetricTypeCounter,
		ID:    "metric_name4",
		Delta: &expectedDelta,
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
			nil,
		),
		expectedMetric: expected,
	}

	expected = Metrics{
		MType: MetricTypeGauge,
		ID:    "metric_name3",
	}
	tests["json. get for gauge"] = testCase{
		request:        newJSONRequest(expected),
		updateHandler:  NewUpdateHandler(RepositoryMock{}, ParametersBagMock{}, nil),
		expectedMetric: expected,
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
			nil,
		),
		expectedMetric: expected,
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := tt.updateHandler.getMetricFromRequest(tt.request)
			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedMetric, actual)
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
