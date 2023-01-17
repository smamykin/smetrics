package handlers

import (
	"errors"
	"net/http"
)

type UpdateHandler struct {
	*Handler
}

func NewUpdateHandler(repository IRepository, parameterBag IParametersBag) *UpdateHandler {
	return &UpdateHandler{
		&Handler{repository, parameterBag},
	}
}

func (u *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	u.handleHeaders(w, r)
	metric, err := u.getMetricFromRequest(r)

	if err != nil {
		if "unknown metric type" == err.Error() {
			http.Error(w, err.Error(), http.StatusNotImplemented)
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = u.upsert(metric)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = u.handleBody(w, metric, r.Header.Get("Accept"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (u *UpdateHandler) upsert(metric Metrics) (err error) {
	if metricTypeGauge == metric.MType {
		return u.Repository.UpsertGauge(GaugeMetric{Name: metric.ID, Value: *metric.Value})
	}

	if metricTypeCounter == metric.MType {
		prevValue, _ := u.Repository.GetCounter(metric.ID)

		return u.Repository.UpsertCounter(CounterMetric{Name: metric.ID, Value: prevValue + *metric.Delta})
	}

	return errors.New("trying to upsert metric with unknown type, there is an error in logic of checking request")
}
