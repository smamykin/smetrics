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
	fsPersister := storage.NewFsPersister(storeFile)

	memStorage := storage.NewMemStorageDefault()
	memStorage.AddObserver(getLoggerObserver(loggerInfo))

	if isRestore {
		restore(storeFile, fsPersister, memStorage, loggerError)
	}

	if storeInterval.Seconds() == 0 {
		memStorage.AddObserver(getPersistToFileObserver(fsPersister, memStorage, loggerError))
	} else {
		go utils.InvokeFunctionWithInterval(storeInterval, getSaveToFileFunction(fsPersister, memStorage, loggerInfo, loggerError))
	}

	r := newRouter(memStorage)
	err := http.ListenAndServe(address, r)
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

func getSaveToFileFunction(f *storage.FsPersister, memStorage *storage.MemStorage, loggerInfo *log.Logger, loggerError *log.Logger) func() {
	return func() {
		loggerInfo.Println("Flushing storage to file")

		err := f.Flush(memStorage)
		if err != nil {
			loggerError.Println(err)
		}
	}
}

func restore(fileName string, fsPersister *storage.FsPersister, memStorage *storage.MemStorage, loggerError *log.Logger) {
	isFileExists, err := utils.IsFileExist(fileName)
	if err != nil {
		loggerError.Printf("Cannot restore the storage from the dump. Error: %s\n", err.Error())
	}

	if !isFileExists {
		return
	}

	if err = fsPersister.Restore(memStorage); err != nil {
		loggerError.Printf("Cannot restore the storage from the dump. Error: %s\n", err.Error())
	}
}

func getLoggerObserver(loggerInfo *log.Logger) storage.Observer {
	return &storage.FuncObserver{
		FunctionToInvoke: func(e storage.IEvent) {
			if _, ok := e.(storage.AfterUpsertEvent); ok {
				loggerInfo.Printf("upsert %#v\n", e.Payload())
			}
		},
	}
}

func getPersistToFileObserver(
	fsPersiter *storage.FsPersister,
	memStorage *storage.MemStorage,
	loggerError *log.Logger,
) storage.Observer {
	return &storage.FuncObserver{
		FunctionToInvoke: func(e storage.IEvent) {
			if _, ok := e.(storage.AfterUpsertEvent); ok {
				err := fsPersiter.Flush(memStorage)
				if err != nil {
					loggerError.Println(err)
				}
			}
		},
	}
}
