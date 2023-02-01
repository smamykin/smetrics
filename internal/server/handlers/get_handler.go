package handlers

import (
	"net/http"
)

type GetHandler struct {
	*Handler
}

func NewGetHandler(repository IRepository, parameterBag IParametersBag, hashGenerator IHashGenerator) *GetHandler {
	return &GetHandler{Handler: &Handler{
		Repository:    repository,
		ParametersBag: parameterBag,
		HashGenerator: hashGenerator,
	}}
}

func (g *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var metric Metrics

	g.handleHeaders(w, r)
	metric, err = g.getMetricFromRequest(r)

	if err != nil {
		if err.Error() == "unknown metric type" {
			http.Error(w, err.Error(), http.StatusNotImplemented)
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = g.handleBody(w, metric, r.Header.Get("Accept"))

	if err == nil {
		return
	}

	if _, ok := err.(MetricNotFoundError); ok {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
}
