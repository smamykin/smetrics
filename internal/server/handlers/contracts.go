package handlers

import "net/http"

const (
	metricTypeGauge   = "gauge"
	metricTypeCounter = "counter"
)

const (
	paramNameMetricType  = "metricType"
	paramNameMetricName  = "metricName"
	paramNameMetricValue = "metricValue"
)

type IRepository interface {
	UpsertGauge(GaugeMetric) error
	UpsertCounter(CounterMetric) error
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAllGauge() ([]GaugeMetric, error)
	GetAllCounters() ([]CounterMetric, error)
}
type IParametersBag interface {
	GetUrlParam(r *http.Request, key string) string
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
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
