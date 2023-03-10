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

type MetricProvider struct{}

func (mp *MetricProvider) GetMetrics(pollCounter int) []IMetric {
	var memStats = runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	return []IMetric{
		MetricGauge{float64(memStats.Alloc), "Alloc"},
		MetricGauge{float64(memStats.BuckHashSys), "BuckHashSys"},
		MetricGauge{float64(memStats.Frees), "Frees"},
		MetricGauge{memStats.GCCPUFraction, "GCCPUFraction"},
		MetricGauge{float64(memStats.GCSys), "GCSys"},
		MetricGauge{float64(memStats.HeapAlloc), "HeapAlloc"},
		MetricGauge{float64(memStats.HeapIdle), "HeapIdle"},
		MetricGauge{float64(memStats.HeapInuse), "HeapInuse"},
		MetricGauge{float64(memStats.HeapObjects), "HeapObjects"},
		MetricGauge{float64(memStats.HeapReleased), "HeapReleased"},
		MetricGauge{float64(memStats.Sys), "HeapSys"},
		MetricGauge{float64(memStats.LastGC), "LastGC"},
		MetricGauge{float64(memStats.Lookups), "Lookups"},
		MetricGauge{float64(memStats.MCacheInuse), "MCacheInuse"},
		MetricGauge{float64(memStats.MCacheSys), "MCacheSys"},
		MetricGauge{float64(memStats.MSpanInuse), "MSpanInuse"},
		MetricGauge{float64(memStats.MSpanSys), "MSpanSys"},
		MetricGauge{float64(memStats.Mallocs), "Mallocs"},
		MetricGauge{float64(memStats.NextGC), "NextGC"},
		MetricGauge{float64(memStats.NumForcedGC), "NumForcedGC"},
		MetricGauge{float64(memStats.NumGC), "NumGC"},
		MetricGauge{float64(memStats.OtherSys), "OtherSys"},
		MetricGauge{float64(memStats.PauseTotalNs), "PauseTotalNs"},
		MetricGauge{float64(memStats.StackInuse), "StackInuse"},
		MetricGauge{float64(memStats.StackSys), "StackSys"},
		MetricGauge{float64(memStats.Sys), "Sys"},
		MetricGauge{float64(memStats.TotalAlloc), "TotalAlloc"},

		//custom
		MetricCounter{pollCounter, "PollCount"},
		MetricGauge{rand.Float64(), "RandomValue"},
	}

}

type MetricGauge struct {
	value float64
	name  string
}

func (m MetricGauge) GetName() string {
	return m.name
}
func (m MetricGauge) GetType() string {
	return MetricTypeGauge
}
func (m MetricGauge) String() string {
	return fmt.Sprintf("%f", m.value)
}

type MetricCounter struct {
	delta int
	name  string
}

func (m MetricCounter) GetName() string {
	return m.name
}
func (m MetricCounter) GetType() string {
	return MetricTypeCounter
}
func (m MetricCounter) String() string {
	return fmt.Sprintf("%d", m.delta)
}
