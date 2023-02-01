package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/smamykin/smetrics/internal/server/handlers"
	"net/http"
)

func NewRouter(repository handlers.IRepository, hashGenerator handlers.IHashGenerator) http.Handler {
	r := chi.NewRouter()

	r.Method("POST", "/update/{metricType}/{metricName}/{metricValue}", handlers.NewUpdateHandler(
		repository,
		ParameterBag{},
		nil,
	))
	r.Method("GET", "/value/{metricType}/{metricName}", handlers.NewGetHandler(
		repository,
		ParameterBag{},
		nil,
	))
	r.Method("GET", "/", &handlers.ListHandler{
		Repository: repository,
	})

	//region JSON-API
	r.Method("POST", "/update/", handlers.NewUpdateHandler(repository, ParameterBag{}, hashGenerator))
	r.Method("POST", "/value/", handlers.NewGetHandler(repository, ParameterBag{}, hashGenerator))
	//endregion

	return gzipHandle(r)
}

type ParameterBag struct{}

func (p ParameterBag) GetURLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
