package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/smamykin/smetrics/internal/server/handlers"
	"net/http"
)

func AddHandlers(r *chi.Mux, repository handlers.IRepository, hashGenerator handlers.IHashGenerator) http.Handler {

	r.Method("POST", "/update/{metricType}/{metricName}/{metricValue}", handlers.NewUpdateHandlerDefault(
		repository,
		ParameterBag{},
	))
	r.Method("GET", "/value/{metricType}/{metricName}", handlers.NewGetHandlerDefault(
		repository,
		ParameterBag{},
	))
	r.Method("GET", "/", &handlers.ListHandler{
		Repository: repository,
	})

	//region JSON-API
	r.Method("POST", "/update/", handlers.NewUpdateHandlerWithHashGenerator(repository, ParameterBag{}, hashGenerator, hashGenerator == nil))
	r.Method("POST", "/value/", handlers.NewGetHandlerWIthHashGenerator(repository, ParameterBag{}, hashGenerator, true))
	//endregion

	return gzipHandle(r)
}

type ParameterBag struct{}

func (p ParameterBag) GetURLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
