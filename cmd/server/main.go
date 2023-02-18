package main

import (
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

	r := chi.NewRouter()

	var repository handlers.IRepository
	if cfg.DatabaseDsn != "" {
		db, err := sql.Open("pgx", cfg.DatabaseDsn)
		if err != nil {
			logger.Error().Msgf("Cannot connect to db. Error: %s\n", err.Error())
		}
		defer db.Close()

		repository, err = createDBStorage(db)
		if err != nil {
			logger.Error().Msgf("Cannot create db storage. Error: %s\n", err.Error())
			return
		}
	} else {
		repository, err = createMemStorage(cfg)
		if err != nil {
			logger.Error().Msgf("Cannot create memStorage. Error: %s\n", err.Error())
			return
		}
	}

	if cfg.Key == "" {
		err = http.ListenAndServe(cfg.Address, server.AddHandlers(r, repository, nil))
	} else {
		err = http.ListenAndServe(cfg.Address, server.AddHandlers(r, repository, utils.NewHashGenerator(cfg.Key)))
	}

	logger.Error().Err(err).Msg("")
}

func createMemStorage(cfg Config) (handlers.IRepository, error) {
	memStorage, err := storage.NewMemStorage(cfg.StoreFile, cfg.Restore, cfg.StoreInterval.Seconds() == 0)
	if err != nil {
		return nil, err
	}
	memStorage.AddObserver(storage.GetLoggerObserver(logger))

	if cfg.StoreInterval.Seconds() != 0 {
		go utils.InvokeFunctionWithInterval(cfg.StoreInterval, getSaveToFileFunction(memStorage))
	}

	return memStorage, nil
}

func createDBStorage(db *sql.DB) (*storage.DBStorage, error) {

	dbStorage, err := storage.NewDBStorage(db)
	if err != nil {
		return nil, err
	}
	dbStorage.AddObserver(storage.GetLoggerObserver(logger))

	return dbStorage, nil
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
