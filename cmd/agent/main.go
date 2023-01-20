package main

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/smamykin/smetrics/internal/agent"
	"github.com/smamykin/smetrics/internal/utils"
	"log"
	"time"
)

type Config struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}

func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting the agent. The configuration: %#v", cfg)

	metricAgent := agent.MetricAgent{
		Client:   agent.NewClient("http://" + cfg.Address),
		Provider: &agent.MetricProvider{},
	}

	go utils.InvokeFunctionWithInterval(cfg.PollInterval, metricAgent.GatherMetrics)
	utils.InvokeFunctionWithInterval(cfg.ReportInterval, metricAgent.SendMetrics)
}
