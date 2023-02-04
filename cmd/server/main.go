package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	defaultDatabaseDsn   = "postgres://postgres:postgres@localhost:54323/postgres"
)

var loggerInfo = log.New(os.Stdout, "INFO:    ", log.Ldate|log.Ltime)
var loggerError = log.New(os.Stdout, "ERROR:   ", log.Ldate|log.Ltime)

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

	//region database connection setup
	db, err := sql.Open("pgx", cfg.DatabaseDsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	//endregion

	memStorage, err := storage.NewMemStorage(cfg.StoreFile, cfg.Restore, cfg.StoreInterval.Seconds() == 0)
	if err != nil {
		loggerError.Printf("Cannot create memStorage. Error: %s\n", err.Error())
	}
	memStorage.AddObserver(storage.GetLoggerObserver(loggerInfo))

	if storeInterval.Seconds() != 0 {
		go utils.InvokeFunctionWithInterval(cfg.StoreInterval, getSaveToFileFunction(memStorage))
	}

	if cfg.Key == "" {
		err = http.ListenAndServe(cfg.Address, server.NewRouter(memStorage, nil, db))
	} else {
		err = http.ListenAndServe(cfg.Address, server.NewRouter(memStorage, utils.NewHashGenerator(cfg.Key), db))
	}

	loggerError.Println(err)
}

func getSaveToFileFunction(memStorage *storage.MemStorage) func() {
	return func() {
		loggerInfo.Println("Flushing storage to file")
		err := memStorage.PersistToFile()
		if err != nil {
			loggerError.Println(err)
		}
	}
}
