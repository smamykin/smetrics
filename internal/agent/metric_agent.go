package agent

import "fmt"

type IMetric interface {
	fmt.Stringer
	GetType() string
	GetName() string
}

type IClient interface {
	SendMetrics(metricType string, metricName string, metricValue string)
}

type IMetricProvider interface {
	GetMetrics(pollCounter int) []IMetric
}

type MetricAgent struct {
	container []IMetric
	Client    IClient
	Provider  IMetricProvider
	counter   int
}

func (mc *MetricAgent) GatherMetrics() {
	mc.counter++
	mc.container = append(mc.container, mc.Provider.GetMetrics(mc.counter)...)
}

func (mc *MetricAgent) SendMetrics() {
	defer mc.reset()

	for _, metric := range mc.container {
		mc.Client.SendMetrics(metric.GetType(), metric.GetName(), metric.String())
	}
}

func (mc *MetricAgent) reset() *MetricAgent {
	mc.container = []IMetric{}
	mc.counter = 0

	return mc
}
