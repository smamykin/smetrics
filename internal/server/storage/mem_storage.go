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

type store struct {
	name     string
	createAt time.Time
}

type GaugeMetric struct {
	value float64
	store
}

type CounterMetric struct {
	value int64
	store
}

type MemStorage struct {
	gaugeStore   map[string]GaugeMetric
	counterStore map[string]CounterMetric
	logger       *log.Logger
}

func (m *MemStorage) UpsertGauge(name string, value float64) error {
	metric := GaugeMetric{
		value,
		store{name, time.Now()},
	}
	m.gaugeStore[metric.name] = metric

	m.logger.Printf("upsert %#v\n", metric)

	return nil
}

func (m *MemStorage) UpsertCounter(name string, value int64) error {
	metric, ok := m.counterStore[name]
	if !ok {
		metric = CounterMetric{
			0,
			store{name, time.Now()},
		}
		m.counterStore[metric.name] = metric
	}
	metric.value = metric.value + value
	m.counterStore[name] = metric

	m.logger.Printf("upsert %#v\n", metric)

	return nil
}
