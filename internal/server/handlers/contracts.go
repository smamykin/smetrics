package handlers

const (
	metricTypeGauge   = "gauge"
	metricTypeCounter = "counter"
)

type IRepository interface {
	UpsertGauge(GaugeMetric) error
	UpsertCounter(CounterMetric) error
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAllGauge() ([]GaugeMetric, error)
	GetAllCounters() ([]CounterMetric, error)
}

type GaugeMetric struct {
	Value float64
	Name  string
}

type CounterMetric struct {
	Value int64
	Name  string
}

type MetricNotFoundError struct {
}

func (m MetricNotFoundError) Error() string {
	return "metric not found"
}
