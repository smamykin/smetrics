package handlers

import (
	"github.com/smamykin/smetrics/internal/server/storage"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdateHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		Repository *repositoryMock
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		expectedStatusCode int
	}{
		{
			name:               "update gauge",
			expectedStatusCode: http.StatusOK,
			fields: fields{Repository: &repositoryMock{
				t:                     t,
				expectedInvokedMethod: "UpsertGauge",
				expectedArgs:          expectedArgs{"metric_name", float64(43)},
			}},
			args: args{
				httptest.NewRecorder(),
				httptest.NewRequest("POST", "/update/gauge/metric_name/43", strings.NewReader("")),
			},
		},
		{
			name:               "update counter",
			expectedStatusCode: http.StatusOK,
			fields: fields{Repository: &repositoryMock{
				t:                     t,
				expectedInvokedMethod: "UpsertCounter",
				expectedArgs:          expectedArgs{"metric_name", int64(43)},
			}},
			args: args{
				httptest.NewRecorder(),
				httptest.NewRequest("POST", "/update/counter/metric_name/43", strings.NewReader("")),
			},
		},
		{
			name:               "method not post",
			expectedStatusCode: http.StatusNotFound,
			fields: fields{Repository: &repositoryMock{
				t:                     t,
				expectedInvokedMethod: "",
				expectedArgs:          expectedArgs{},
			}},
			args: args{
				httptest.NewRecorder(),
				httptest.NewRequest("GET", "/update/counter/metric_name/43", strings.NewReader("")),
			},
		},
		{
			name:               "wrong url",
			expectedStatusCode: http.StatusNotFound,
			fields: fields{Repository: &repositoryMock{
				t:                     t,
				expectedInvokedMethod: "",
				expectedArgs:          expectedArgs{},
			}},
			args: args{
				httptest.NewRecorder(),
				httptest.NewRequest("POST", "/update/counter/metric_name/43/some_tail", strings.NewReader("")),
			},
		},
		{
			name:               "unknown type",
			expectedStatusCode: http.StatusNotImplemented,
			fields: fields{Repository: &repositoryMock{
				t:                     t,
				expectedInvokedMethod: "",
				expectedArgs:          expectedArgs{},
			}},
			args: args{
				httptest.NewRecorder(),
				httptest.NewRequest("POST", "/update/unknown_type/metric_name/43", strings.NewReader("")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UpdateHandler{
				Repository: tt.fields.Repository,
			}
			u.ServeHTTP(tt.args.w, tt.args.r)

			require.Equal(t, len(tt.fields.Repository.expectedInvokedMethod) != 0, tt.fields.Repository.isInvoked)
			res := tt.args.w.Result()
			defer res.Body.Close()
			_, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, tt.expectedStatusCode, res.StatusCode)
		})
	}
}

type repositoryMock struct {
	t                     *testing.T
	expectedArgs          expectedArgs
	expectedInvokedMethod string
	isInvoked             bool
}

func (r *repositoryMock) GetGauge(name string) (float64, error) {
	panic("must not be used in the test")
}

func (r *repositoryMock) GetCounter(name string) (int64, error) {
	panic("must not be used in the test")
}

func (r *repositoryMock) GetAllGauge() []storage.GaugeMetric {
	panic("must not be used in the test")
}

func (r *repositoryMock) GetAllCounters() []storage.CounterMetric {
	panic("must not be used in the test")
}

type expectedArgs struct {
	name  string
	value interface{}
}

func (r *repositoryMock) UpsertGauge(name string, value float64) error {
	require.True(r.t, r.expectedInvokedMethod == "UpsertGauge")
	require.Equal(r.t, r.expectedArgs, expectedArgs{name, value})
	r.isInvoked = true
	return nil
}

func (r *repositoryMock) UpsertCounter(name string, value int64) error {
	require.True(r.t, r.expectedInvokedMethod == "UpsertCounter")
	require.Equal(r.t, r.expectedArgs, expectedArgs{name, value})
	r.isInvoked = true
	return nil
}
