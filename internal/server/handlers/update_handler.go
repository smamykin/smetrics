package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type UpdateHandler struct {
	Repository IRepository
}

func (u *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-Type", "text/plain")

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	switch metricType {
	case metricTypeGauge:
		err = u.upsertGauge(metricName, metricValue)
	case metricTypeCounter:
		err = u.upsertCounter(metricName, metricValue)
	default:
		http.Error(w, "metric type is incorrect", http.StatusNotImplemented)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (u *UpdateHandler) upsertCounter(metricName string, metricValue string) error {
	value, err := strconv.ParseInt(metricValue, 10, 64)
	if err != nil {
		return err
	}

	prevValue, _ := u.Repository.GetCounter(metricName)

	return u.Repository.UpsertCounter(CounterMetric{Name: metricName, Value: prevValue + value})
}

func (u *UpdateHandler) upsertGauge(metricName string, metricValue string) error {

	value, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		return err
	}

	return u.Repository.UpsertGauge(GaugeMetric{Name: metricName, Value: value})
}
