package main

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/smamykin/smetrics/internal/server/server"
	"log"
	"time"
)

type Config struct {
	Address       string        `env:"ADDRESS" envDefault:"localhost:8080"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
}

func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting the server. The configuration: %#v", cfg)

	server.ListenAndServ(cfg.Address, cfg.Restore, cfg.StoreFile, cfg.StoreInterval)
}
