package storage

import (
	"github.com/smamykin/smetrics/internal/server/handlers"
)

func NewMemStorageDefault() *MemStorage {
	return &MemStorage{
		counterStore: map[string]handlers.CounterMetric{},
		gaugeStore:   map[string]handlers.GaugeMetric{},
	}
}
func NewMemStorage(counterStore map[string]handlers.CounterMetric, gaugeStore map[string]handlers.GaugeMetric) *MemStorage {
	return &MemStorage{
		counterStore: counterStore,
		gaugeStore:   gaugeStore,
	}
}

type MemStorage struct {
	gaugeStore   map[string]handlers.GaugeMetric
	counterStore map[string]handlers.CounterMetric
	observers    []Observer
}

func (m *MemStorage) AddObserver(o Observer) {
	m.observers = append(m.observers, o)
}

func (m *MemStorage) GaugeStore() map[string]handlers.GaugeMetric {
	return m.gaugeStore
}

func (m *MemStorage) CounterStore() map[string]handlers.CounterMetric {
	return m.counterStore
}

func (m *MemStorage) GetAllGauge() ([]handlers.GaugeMetric, error) {
	var result []handlers.GaugeMetric

	for _, value := range m.gaugeStore {
		result = append(result, value)
	}
	return result, nil
}

func (m *MemStorage) GetAllCounters() ([]handlers.CounterMetric, error) {
	var result []handlers.CounterMetric

	for _, value := range m.counterStore {
		result = append(result, value)
	}
	return result, nil
}

func (m *MemStorage) GetGauge(name string) (float64, error) {
	metric, ok := m.gaugeStore[name]
	if !ok {
		return .0, handlers.MetricNotFoundError{}
	}

	return metric.Value, nil
}

func (m *MemStorage) GetCounter(name string) (int64, error) {
	metric, ok := m.counterStore[name]
	if !ok {
		return 0, handlers.MetricNotFoundError{}
	}

	return metric.Value, nil
}

func (m *MemStorage) UpsertGauge(metric handlers.GaugeMetric) error {
	m.gaugeStore[metric.Name] = metric

	m.notifyObservers(AfterUpsertEvent{
		Event{metric},
	})

	return nil
}

func (m *MemStorage) UpsertCounter(metric handlers.CounterMetric) error {
	m.counterStore[metric.Name] = metric

	m.notifyObservers(AfterUpsertEvent{
		Event{metric},
	})

	return nil
}

func (m *MemStorage) notifyObservers(event IEvent) {
	for _, observer := range m.observers {
		observer.HandleEvent(event)
	}
}
