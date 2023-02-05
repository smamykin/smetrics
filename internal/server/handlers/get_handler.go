package handlers

import (
	"errors"
	"net/http"
)

type GetHandler struct {
	*Handler
}

func NewGetHandlerDefault(repository IRepository, parameterBag IParametersBag) *GetHandler {
	return &GetHandler{Handler: &Handler{
		Repository:                  repository,
		ParametersBag:               parameterBag,
		HashGenerator:               nil,
		IsSkipCheckOfHashForRequest: true,
	}}
}

func NewGetHandlerWithHashGenerator(repository IRepository, parameterBag IParametersBag, hashGenerator IHashGenerator, isSkipCheckOfHashForRequest bool) *GetHandler {
	return &GetHandler{Handler: &Handler{
		Repository:                  repository,
		ParametersBag:               parameterBag,
		HashGenerator:               hashGenerator,
		IsSkipCheckOfHashForRequest: isSkipCheckOfHashForRequest,
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

	if errors.Is(err, ErrMetricNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
}
