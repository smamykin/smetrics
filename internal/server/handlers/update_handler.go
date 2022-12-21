package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type IRepository interface {
	UpsertGauge(name string, value float64) error
	UpsertCounter(name string, value int64) error
}

type UpdateHandler struct {
	Repository IRepository
}

func (u *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-Type", "text/plain")

	pathElements := strings.Split(r.URL.Path, "/")
	if len(pathElements) != 5 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	metricType := pathElements[2]
	metricName := pathElements[3]
	metricValue := pathElements[4]

	switch metricType {
	case "gauge":
		err = u.upsertGauge(metricName, metricValue)
	case "counter":
		err = u.upsertCounter(metricName, metricValue)
	default:
		err = errors.New("metric type is incorrect")
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

	return u.Repository.UpsertCounter(metricName, value)
}

func (u *UpdateHandler) upsertGauge(metricName string, metricValue string) error {

	value, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		return err
	}

	return u.Repository.UpsertGauge(metricName, value)
}
