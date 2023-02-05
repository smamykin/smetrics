package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/smamykin/smetrics/internal/server/handlers"
	"github.com/smamykin/smetrics/internal/server/server"
	"github.com/smamykin/smetrics/internal/server/storage"
	"github.com/smamykin/smetrics/internal/utils"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Address       string        `env:"ADDRESS"`
	Restore       bool          `env:"RESTORE"`
	StoreFile     string        `env:"STORE_FILE"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	Key           string        `env:"KEY"`
	DatabaseDsn   string        `env:"DATABASE_DSN"`
}

const (
	defaultAddress       = "localhost:8080"
	defaultRestore       = true
	defaultStoreFile     = "/tmp/devops-metrics-db.json"
	defaultStoreInterval = time.Second * 300
	defaultKey           = ""
	defaultDatabaseDsn   = ""
)

var logger = zerolog.New(os.Stdout)

func main() {

	address := flag.String("a", defaultAddress, "The address of the server")
	restore := flag.Bool("r", defaultRestore, "To restore the dump from the file")
	storeFile := flag.String("f", defaultStoreFile, "the absolute path to the dump file.")
	storeInterval := flag.Duration("i", defaultStoreInterval, "How often to save the dump of the metrics")
	key := flag.String("k", defaultKey, "The secret key")
	databaseDsn := flag.String("d", defaultDatabaseDsn, "The database url")
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
	if _, isPresent := os.LookupEnv("KEY"); !isPresent {
		cfg.Key = *key
	}
	if _, isPresent := os.LookupEnv("DATABASE_DSN"); !isPresent {
		cfg.DatabaseDsn = *databaseDsn
	}

	fmt.Printf("Starting the server. The configuration: %#v\n", cfg)

	//var repository handlers.IRepository

	r := chi.NewRouter()

	var repository handlers.IRepository
	if cfg.DatabaseDsn != "" {
		//region database connection setup
		db, err := sql.Open("pgx", cfg.DatabaseDsn)
		if err != nil {
			logger.Error().Msgf("Cannot connect to db. Error: %s\n", err.Error())
			return
		}
		defer db.Close()

		addProbeEndpoint(r, db)
		//endregion
		dbStorage, err := storage.NewDBStorage(db)
		if err != nil {
			logger.Error().Msgf("Cannot create dbStorage. Error: %s\n", err.Error())
			return
		}
		dbStorage.AddObserver(storage.GetLoggerObserver(logger))

		repository = dbStorage
	} else {
		memStorage, err := storage.NewMemStorage(cfg.StoreFile, cfg.Restore, cfg.StoreInterval.Seconds() == 0)
		if err != nil {
			logger.Error().Msgf("Cannot create memStorage. Error: %s\n", err.Error())
		}
		memStorage.AddObserver(storage.GetLoggerObserver(logger))

		if storeInterval.Seconds() != 0 {
			go utils.InvokeFunctionWithInterval(cfg.StoreInterval, getSaveToFileFunction(memStorage))
		}
		repository = memStorage
	}

	if cfg.Key == "" {
		err = http.ListenAndServe(cfg.Address, server.AddHandlers(r, repository, nil))
	} else {
		err = http.ListenAndServe(cfg.Address, server.AddHandlers(r, repository, utils.NewHashGenerator(cfg.Key)))
	}

	logger.Error().Err(err).Msg("")
}

func addProbeEndpoint(r *chi.Mux, db *sql.DB) {
	r.Method("GET", "/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
}

func getSaveToFileFunction(memStorage *storage.MemStorage) func() {
	return func() {
		logger.Info().Msg("Flushing storage to file")
		err := memStorage.PersistToFile()
		if err != nil {
			logger.Error().Err(err).Msg("")
		}
	}
}
