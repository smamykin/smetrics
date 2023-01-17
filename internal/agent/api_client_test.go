package agent

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestClient_SendMetrics(t *testing.T) {

	value := rand.Int()
	handler := handlerForTest{
		expectedMethod:      "POST",
		expectedPath:        "/update/",
		expectedContentType: "application/json",
		expectedBody:        fmt.Sprintf(`{"id":"metricNameTest","type":"counter","delta":%d}`, value),
		t:                   t,
	}
	server := httptest.NewServer(&handler)
	defer server.Close()

	client := Client{
		server.URL,
		log.New(writerMock{}, "test: ", log.Ldate|log.Ltime),
		log.New(writerMock{}, "test: ", log.Ldate|log.Ltime),
	}
	client.SendMetrics("counter", "metricNameTest", strconv.Itoa(value))

	require.True(t, handler.isInvoked)
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
	defer request.Body.Close()
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return
	}
	require.Equal(h.t, h.expectedBody, string(body))
}

type writerMock struct {
}

func (t writerMock) Write(p []byte) (n int, err error) {
	return 0, nil
}
