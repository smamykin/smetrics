package agent

import (
	"fmt"
	"math/rand"
	"runtime"
)

const (
	MetricTypeGauge   = "gauge"
	MetricTypeCounter = "counter"
)

type MetricProvider struct {
	counter int
}

func (mp *MetricProvider) GetMetrics() map[string]IMetric {
	var memStats = runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	mp.counter++

	return map[string]IMetric{
		"Alloc":         MetricGauge(memStats.Alloc),
		"BuckHashSys":   MetricGauge(memStats.BuckHashSys),
		"Frees":         MetricGauge(memStats.Frees),
		"GCCPUFraction": MetricGauge(memStats.GCCPUFraction),
		"GCSys":         MetricGauge(memStats.GCSys),
		"HeapAlloc":     MetricGauge(memStats.HeapAlloc),
		"HeapIdle":      MetricGauge(memStats.HeapIdle),
		"HeapInuse":     MetricGauge(memStats.HeapInuse),
		"HeapObjects":   MetricGauge(memStats.HeapObjects),
		"HeapReleased":  MetricGauge(memStats.HeapReleased),
		"HeapSys":       MetricGauge(memStats.Sys),
		"LastGC":        MetricGauge(memStats.LastGC),
		"Lookups":       MetricGauge(memStats.Lookups),
		"MCacheInuse":   MetricGauge(memStats.MCacheInuse),
		"MCacheSys":     MetricGauge(memStats.MCacheSys),
		"MSpanInuse":    MetricGauge(memStats.MSpanInuse),
		"MSpanSys":      MetricGauge(memStats.MSpanSys),
		"Mallocs":       MetricGauge(memStats.Mallocs),
		"NextGC":        MetricGauge(memStats.NextGC),
		"NumForcedGC":   MetricGauge(memStats.NumForcedGC),
		"NumGC":         MetricGauge(memStats.NumGC),
		"OtherSys":      MetricGauge(memStats.OtherSys),
		"PauseTotalNs":  MetricGauge(memStats.PauseTotalNs),
		"StackInuse":    MetricGauge(memStats.StackInuse),
		"StackSys":      MetricGauge(memStats.StackSys),
		"Sys":           MetricGauge(memStats.Sys),
		"TotalAlloc":    MetricGauge(memStats.TotalAlloc),

		//custom
		"PollCount":   MetricCounter(mp.counter),
		"RandomValue": MetricGauge(rand.Float64()),
	}

}

type MetricGauge float64
type MetricCounter int64

func (m MetricGauge) String() string {
	return fmt.Sprintf("%f", m)
}
func (m MetricGauge) GetType() string {
	return MetricTypeGauge
}
func (m MetricCounter) String() string {
	return fmt.Sprintf("%d", m)
}
func (m MetricCounter) GetType() string {
	return MetricTypeCounter
}
