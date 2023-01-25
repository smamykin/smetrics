package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/smamykin/smetrics/internal/agent"
	"github.com/smamykin/smetrics/internal/utils"
	"log"
	"os"
	"time"
)

type Config struct {
	Address        string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
}

const (
	defaultAddress        = "http://localhost:8080"
	defaultReportInterval = time.Second * 10
	defaultPollInterval   = time.Second * 2
	defaultSchema         = "http://"
)

func main() {
	address := flag.String("a", defaultAddress, "The address of the metric server")
	reportInterval := flag.Duration("r", defaultReportInterval, "How often to send metrics to server")
	pollInterval := flag.Duration("p", defaultPollInterval, "How often to refresh metrics")
	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if _, isPresent := os.LookupEnv("ADDRESS"); !isPresent {
		cfg.Address = *address
	}
	if _, isPresent := os.LookupEnv("REPORT_INTERVAL"); !isPresent {
		cfg.ReportInterval = *reportInterval
	}
	if _, isPresent := os.LookupEnv("POLL_INTERVAL"); !isPresent {
		cfg.PollInterval = *pollInterval
	}

	fmt.Printf("Starting the agent. The configuration: %#v", cfg)
	metricAgent := agent.MetricAgent{
		Client:   agent.NewClient(cfg.Address),
		Provider: &agent.MetricProvider{},
	}

	go utils.InvokeFunctionWithInterval(cfg.PollInterval, metricAgent.GatherMetrics)
	utils.InvokeFunctionWithInterval(cfg.ReportInterval, metricAgent.SendMetrics)
}
