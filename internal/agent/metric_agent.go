package agent

import "fmt"

type IMetric interface {
	fmt.Stringer
	GetType() string
}

type IClient interface {
	SendMetrics(metricType string, metricName string, metricValue string)
}

type IMetricProvider interface {
	GetMetrics(pollCounter int) map[string]IMetric
}

type MetricAgent struct {
	container []map[string]IMetric
	Client    IClient
	Provider  IMetricProvider
}

func (mc *MetricAgent) GatherMetrics() {
	mc.container = append(mc.container, mc.Provider.GetMetrics(len(mc.container)))
}

func (mc *MetricAgent) SendMetrics() {
	defer mc.reset()

	for _, metricMap := range mc.container {
		for name, metric := range metricMap {
			mc.Client.SendMetrics(metric.GetType(), name, metric.String())
		}
	}
}

func (mc *MetricAgent) reset() *MetricAgent {
	mc.container = []map[string]IMetric{}

	return mc
}
