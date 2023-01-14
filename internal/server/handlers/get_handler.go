package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type GetHandler struct {
	*Handler
}

func NewGetHandler(repository IRepository, parameterBag IParametersBag) *GetHandler {
	return &GetHandler{&Handler{
		repository,
		parameterBag,
	}}
}

func (g *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-Type", "text/plain")

	metricType := chi.URLParam(r, paramNameMetricType)
	metricName := chi.URLParam(r, paramNameMetricName)

	switch metricType {
	case metricTypeGauge:
		var value float64
		value, err = g.Repository.GetGauge(metricName)
		if err == nil {
			w.Write([]byte(fmt.Sprintf("%.3f", value)))
			return
		}
	case metricTypeCounter:
		var value int64
		value, err = g.Repository.GetCounter(metricName)
		if err == nil {
			w.Write([]byte(fmt.Sprintf("%d", value)))
			return
		}
	default:
		http.Error(w, "metric type is incorrect", http.StatusNotImplemented)
		return
	}

	if _, ok := err.(MetricNotFoundError); ok {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Error(w, err.Error(), http.StatusBadRequest)
}
