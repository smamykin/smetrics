package storage

import (
	"github.com/smamykin/smetrics/internal/server/handlers"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

type ObserverSpy struct {
	events []IEvent
}

func (o *ObserverSpy) HandleEvent(e IEvent) {
	o.events = append(o.events, e)
}

func TestMemStorage_UpsertCounter(t *testing.T) {
	store := map[string]handlers.CounterMetric{}
	m := &MemStorage{
		counterStore: store,
		gaugeStore:   map[string]handlers.GaugeMetric{},
	}

	metric := handlers.CounterMetric{rand.Int63(), "metric_name"}
	spy := &ObserverSpy{}
	m.AddObserver(spy)
	m.UpsertCounter(metric)
	require.Equal(t, map[string]handlers.CounterMetric{metric.Name: metric}, store)
	require.Equal(
		t,
		spy.events,
		[]IEvent{BeforeUpsertEvent{Event{payload: metric}}, AfterUpsertEvent{Event{payload: metric}}},
	)
}

func TestMemStorage_UpsertGauge(t *testing.T) {
	store := map[string]handlers.GaugeMetric{}
	m := &MemStorage{
		counterStore: map[string]handlers.CounterMetric{},
		gaugeStore:   store,
	}

	metric := handlers.GaugeMetric{rand.Float64(), "metric_name"}
	spy := &ObserverSpy{}
	m.AddObserver(spy)
	m.UpsertGauge(metric)

	require.Equal(t, map[string]handlers.GaugeMetric{metric.Name: metric}, store)
	require.Equal(
		t,
		spy.events,
		[]IEvent{BeforeUpsertEvent{Event{payload: metric}}, AfterUpsertEvent{Event{payload: metric}}},
	)
}
