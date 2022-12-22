package storage

import (
	"log"
	"os"
	"time"
)

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counterStore: map[string]CounterMetric{},
		gaugeStore:   map[string]GaugeMetric{},
		logger:       log.New(os.Stdout, "INFO:    ", log.Ldate|log.Ltime),
	}
}

type metric struct {
	Name     string
	createAt time.Time
}

type GaugeMetric struct {
	Value float64
	metric
}

type CounterMetric struct {
	Value int64
	metric
}

type MemStorage struct {
	gaugeStore   map[string]GaugeMetric
	counterStore map[string]CounterMetric
	logger       *log.Logger
}

func (m *MemStorage) GetAllGauge() []GaugeMetric {
	var result []GaugeMetric

	for _, value := range m.gaugeStore {
		result = append(result, value)
	}
	return result
}

func (m *MemStorage) GetAllCounters() []CounterMetric {
	var result []CounterMetric

	for _, value := range m.counterStore {
		result = append(result, value)
	}
	return result
}

func (m *MemStorage) GetGauge(name string) (float64, error) {
	metric, ok := m.gaugeStore[name]
	if !ok {
		return .0, NotFoundError{"metric not found"}
	}

	return metric.Value, nil
}

func (m *MemStorage) GetCounter(name string) (int64, error) {
	metric, ok := m.counterStore[name]
	if !ok {
		return .0, NotFoundError{"metric not found"}
	}

	return metric.Value, nil
}

func (m *MemStorage) UpsertGauge(name string, value float64) error {
	metric := GaugeMetric{
		value,
		metric{name, time.Now()},
	}
	m.gaugeStore[metric.Name] = metric

	m.logger.Printf("upsert %#v\n", metric)

	return nil
}

func (m *MemStorage) UpsertCounter(name string, value int64) error {
	metricInstance, ok := m.counterStore[name]
	if !ok {
		metricInstance = CounterMetric{
			0,
			metric{name, time.Now()},
		}
		m.counterStore[metricInstance.Name] = metricInstance
	}
	metricInstance.Value = metricInstance.Value + value
	m.counterStore[name] = metricInstance

	m.logger.Printf("upsert %#v\n", metricInstance)

	return nil
}
