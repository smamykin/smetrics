package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/smamykin/smetrics/internal/server/handlers"
	"github.com/smamykin/smetrics/internal/server/storage"
	"github.com/smamykin/smetrics/internal/utils"
	"log"
	"net/http"
	"os"
	"time"
)

func ListenAndServ(address string, isRestore bool, storeFile string, storeInterval time.Duration) {
	loggerInfo := log.New(os.Stdout, "INFO:    ", log.Ldate|log.Ltime)
	loggerError := log.New(os.Stdout, "ERROR:   ", log.Ldate|log.Ltime)

	memStorage, err := storage.NewMemStorage(storeFile, isRestore, storeInterval.Seconds() == 0)
	if err != nil {
		loggerError.Printf("Cannot create memStorage. Error: %s\n", err.Error())
	}
	memStorage.AddObserver(getLoggerObserver(loggerInfo))

	if storeInterval.Seconds() != 0 {
		go utils.InvokeFunctionWithInterval(storeInterval, getSaveToFileFunction(memStorage, loggerError))
	}

	r := newRouter(memStorage)
	err = http.ListenAndServe(address, r)
	loggerError.Println(err)
}

func newRouter(repository handlers.IRepository) http.Handler {
	r := chi.NewRouter()

	r.Method("POST", "/update/{metricType}/{metricName}/{metricValue}", handlers.NewUpdateHandler(
		repository,
		ParameterBag{},
	))
	r.Method("GET", "/value/{metricType}/{metricName}", handlers.NewGetHandler(
		repository,
		ParameterBag{},
	))
	r.Method("GET", "/", &handlers.ListHandler{
		Repository: repository,
	})

	//region JSON-API
	r.Method("POST", "/update/", handlers.NewUpdateHandler(repository, ParameterBag{}))
	r.Method("POST", "/value/", handlers.NewGetHandler(repository, ParameterBag{}))
	//endregion

	return gzipHandle(r)
}

type ParameterBag struct{}

func (p ParameterBag) GetURLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func getSaveToFileFunction(memStorage *storage.MemStorage, loggerError *log.Logger) func() {
	return func() {
		err := memStorage.PersistToFile()
		if err != nil {
			loggerError.Println(err)
		}
	}
}

func getLoggerObserver(loggerInfo *log.Logger) storage.Observer {
	return &storage.FuncObserver{
		FunctionToInvoke: func(e storage.IEvent) error {
			if _, ok := e.(storage.AfterUpsertEvent); ok {
				loggerInfo.Printf("upsert %#v\n", e.Payload())
			}
			return nil
		},
	}
}
