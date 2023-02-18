package agent

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/smamykin/smetrics/internal/utils"
	"github.com/stretchr/testify/require"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_SendMetrics(t *testing.T) {
	type testCase struct {
		key          string
		expectedBody string
	}
	value := rand.Int()

	h := utils.NewHashGenerator("secret")
	sign, err := h.Generate(fmt.Sprintf("metricNameTest:counter:%d", value))
	require.Nil(t, err)

	tests := map[string]testCase{
		"default": {
			"",
			fmt.Sprintf(`[{"id":"metricNameTest","type":"counter","delta":%d}]`, value),
		},
		"with key": {
			"secret",
			fmt.Sprintf(`[{"id":"metricNameTest","type":"counter","delta":%d,"hash":"%s"}]`, value, sign),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := handlerForTest{
				expectedMethod:      "POST",
				expectedPath:        "/updates/",
				expectedContentType: "application/json",
				expectedBody:        tt.expectedBody,
				t:                   t,
			}
			server := httptest.NewServer(&handler)
			defer server.Close()

			logger := zerolog.Nop()
			client := NewClient(
				&logger,
				server.URL,
				tt.key,
				1,
			)
			client.SendMetrics([]IMetric{
				MetricCounter{value, "metricNameTest"},
			})

			require.True(t, handler.isInvoked)
		})
	}
}

type handlerForTest struct {
	t                   *testing.T
	expectedMethod      string
	expectedPath        string
	expectedContentType string
	expectedBody        string
	isInvoked           bool
}

func (h *handlerForTest) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.isInvoked = true
	require.Equal(h.t, h.expectedMethod, request.Method)
	require.Equal(h.t, h.expectedContentType, request.Header.Get("Content-Type"))
	require.Equal(h.t, h.expectedPath, request.URL.Path)
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return
	}
	defer request.Body.Close()
	require.Equal(h.t, h.expectedBody, string(body))
}

type writerMock struct {
}

func (t writerMock) Write(p []byte) (n int, err error) {
	return 0, nil
}
