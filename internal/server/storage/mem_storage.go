package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/smamykin/smetrics/internal/server/handlers"
)

func NewMemStorage(storeFile string, isRestore bool, isPersistSynchronouslyToFile bool) (memStorage *MemStorage, err error) {
	persister, err := newFsPersister(storeFile)
	if err != nil {
		return memStorage, err
	}
	memStorage = &MemStorage{
		counterStore: map[string]handlers.CounterMetric{},
		gaugeStore:   map[string]handlers.GaugeMetric{},
		fsPersister:  persister,
	}

	if isRestore {
		if err := memStorage.restore(); err != nil {
			return memStorage, err
		}
	}

	if isPersistSynchronouslyToFile {
		memStorage.AddObserver(newPersistToFileObserver(memStorage))
	}
	return memStorage, nil
}

func NewMemStorageDefault() *MemStorage {
	return &MemStorage{
		counterStore: map[string]handlers.CounterMetric{},
		gaugeStore:   map[string]handlers.GaugeMetric{},
	}
}

type MemStorage struct {
	gaugeStore   map[string]handlers.GaugeMetric
	counterStore map[string]handlers.CounterMetric
	observers    []Observer
	fsPersister  *fsPersister
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

	return m.notifyObservers(AfterUpsertEvent{
		Event{metric},
	})
}

func (m *MemStorage) UpsertCounter(metric handlers.CounterMetric) error {
	m.counterStore[metric.Name] = metric

	return m.notifyObservers(AfterUpsertEvent{
		Event{metric},
	})
}

func (m *MemStorage) UpsertMany(ctx context.Context, metrics []interface{}) error {
	for _, metric := range metrics {
		_, isCounterMetric := metric.(handlers.CounterMetric)
		_, isGaugeMetric := metric.(handlers.GaugeMetric)
		if !isCounterMetric && !isGaugeMetric {
			return errors.New("unknown metric type")
		}
	}
	for _, metric := range metrics {
		switch metric.(type) {
		case handlers.GaugeMetric:
			gaugeMetric := metric.(handlers.GaugeMetric)
			if err := m.UpsertGauge(gaugeMetric); err != nil {
				return err
			}
		case handlers.CounterMetric:
			counterMetric := metric.(handlers.CounterMetric)
			if err := m.UpsertCounter(counterMetric); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *MemStorage) notifyObservers(event IEvent) error {
	for _, observer := range m.observers {
		if err := observer.HandleEvent(event); err != nil {
			return err
		}
	}
	return nil
}

func (m *MemStorage) restore() error {
	if err := m.fsPersister.restore(m); err != nil {
		return fmt.Errorf("cannot restore the storage from the dump. Error: %w", err)
	}

	return nil
}

func (m *MemStorage) PersistToFile() error {
	return m.fsPersister.flush(m)
}

func newPersistToFileObserver(memStorage *MemStorage) Observer {
	return &FuncObserver{
		FunctionToInvoke: func(e IEvent) error {
			if _, ok := e.(AfterUpsertEvent); ok {
				return memStorage.PersistToFile()
			}
			return nil
		},
	}
}
