package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/smamykin/smetrics/internal/server/server"
	"log"
	"os"
	"time"
)

type Config struct {
	Address       string        `env:"ADDRESS"`
	Restore       bool          `env:"RESTORE"`
	StoreFile     string        `env:"STORE_FILE"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
}

const (
	defaultAddress       = "localhost:8080"
	defaultRestore       = true
	defaultStoreFile     = "/tmp/devops-metrics-db.json"
	defaultStoreInterval = time.Second * 300
)

func main() {
	address := flag.String("a", defaultAddress, "The address of the server")
	restore := flag.Bool("r", defaultRestore, "To restore the dump from the file")
	storeFile := flag.String("f", defaultStoreFile, "the absolute path to the dump file.")
	storeInterval := flag.Duration("i", defaultStoreInterval, "How often to save the dump of the metrics")
	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if _, isPresent := os.LookupEnv("ADDRESS"); !isPresent {
		cfg.Address = *address
	}
	if _, isPresent := os.LookupEnv("RESTORE"); !isPresent {
		cfg.Restore = *restore
	}
	if _, isPresent := os.LookupEnv("STORE_FILE"); !isPresent {
		cfg.StoreFile = *storeFile
	}
	if _, isPresent := os.LookupEnv("STORE_INTERVAL"); !isPresent {
		cfg.StoreInterval = *storeInterval
	}

	fmt.Printf("Starting the server. The configuration: %#v", cfg)

	server.ListenAndServ(cfg.Address, cfg.Restore, cfg.StoreFile, cfg.StoreInterval)
}
