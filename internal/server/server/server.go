package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/smamykin/smetrics/internal/server/handlers"
	"github.com/smamykin/smetrics/internal/server/storage"
	"log"
	"net/http"
)

func ListenAndServ() {
	memStorage := storage.NewMemStorage()
	r := NewRouter(memStorage)
	log.Println("Starting the server.")
	err := http.ListenAndServe("127.0.0.1:8080", r)
	panic(err)
}

func NewRouter(repository handlers.IRepository) chi.Router {
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

	return r
}

type ParameterBag struct{}

func (p ParameterBag) GetUrlParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
