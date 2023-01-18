package main

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/smamykin/smetrics/internal/server/server"
	"log"
)

type Config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
}

func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting the server. The configuration: %#v", cfg)

	server.ListenAndServ(cfg.Address)
}
