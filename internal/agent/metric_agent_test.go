package agent

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestMetricAgent_GatherMetrics(t *testing.T) {
	ma := &MetricAgent{
		container: []IMetric{},
		Client:    &apiClientMock{},
		Provider:  &providerMock{0},
	}
	ma.GatherMetrics()
	require.Equal(t, []IMetric{
		metricMock{"1", "gauge", "metricTestName1"},
		metricMock{"1", "counter", "metricTestName2"},
	}, ma.container)

	ma.GatherMetrics()
	require.Equal(t, []IMetric{
		metricMock{"1", "gauge", "metricTestName1"},
		metricMock{"1", "counter", "metricTestName2"},
		metricMock{"2", "gauge", "metricTestName1"},
		metricMock{"2", "counter", "metricTestName2"},
	}, ma.container)
}

func TestMetricAgent_SendMetrics(t *testing.T) {
	clientMock := apiClientMock{0, []IMetric{
		metricMock{"11", "gauge", "metricTestName1"},
		metricMock{"22", "counter", "metricTestName2"},
		metricMock{"22", "gauge", "metricTestName1"},
		metricMock{"33", "counter", "metricTestName2"},
	}, t}
	ma := &MetricAgent{
		container: []IMetric{
			metricMock{"11", "gauge", "metricTestName1"},
			metricMock{"22", "counter", "metricTestName2"},
			metricMock{"22", "gauge", "metricTestName1"},
			metricMock{"33", "counter", "metricTestName2"},
		},
		Client:   &clientMock,
		Provider: &providerMock{},
	}

	ma.SendMetrics()
	require.Empty(t, ma.container)
	require.Equal(t, 4, clientMock.invokedTimes)
}

type apiClientMock struct {
	invokedTimes int
	expectedArgs []IMetric
	t            *testing.T
}

func (a *apiClientMock) SendMetrics(metricType string, metricName string, metricValue string) error {
	assert.Equal(a.t, a.expectedArgs[a.invokedTimes], metricMock{
		metricValue,
		metricType,
		metricName,
	})
	a.invokedTimes++
	return nil
}

type providerMock struct {
	invokedCount int
}
type metricMock struct {
	value        string
	typeOfMetric string
	name         string
}

func (m metricMock) GetName() string {
	return m.name
}

func (m metricMock) String() string {
	return m.value
}

func (m metricMock) GetType() string {
	return m.typeOfMetric
}

func (p *providerMock) GetMetrics(pollCounter int) []IMetric {
	p.invokedCount++
	if p.invokedCount == 1 {
		return []IMetric{
			metricMock{"1", "gauge", "metricTestName1"},
			metricMock{strconv.Itoa(pollCounter), "counter", "metricTestName2"},
		}
	}

	return []IMetric{
		metricMock{"2", "gauge", "metricTestName1"},
		metricMock{strconv.Itoa(pollCounter), "counter", "metricTestName2"},
	}
}
