package main

import (
	"github.com/smamykin/smetrics/internal/agent"
	"time"
)

const (
	gatherInterval = 2 * time.Second
	sendInterval   = 10 * time.Second
)

func main() {
	metricAgent := agent.MetricAgent{
		Client: agent.NewClient(
			"http://127.0.0.1:8080",
			"text/plain",
		),
		Provider: &agent.MetricProvider{},
	}

	go invokeFunctionWithInterval(gatherInterval, metricAgent.GatherMetrics)
	invokeFunctionWithInterval(sendInterval, metricAgent.SendMetrics)
}

func invokeFunctionWithInterval(duration time.Duration, functionToInvoke func()) {
	ticker := time.NewTicker(duration)
	for {
		<-ticker.C
		functionToInvoke()
	}
}
