package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type UpdatesHandler struct {
	*Handler
}

func NewUpdatesHandlerWithHashGenerator(repository IRepository, parameterBag IParametersBag, hashGenerator IHashGenerator, isSkipCheckOfHashForRequest bool) *UpdatesHandler {
	return &UpdatesHandler{
		Handler: &Handler{
			Repository:                  repository,
			ParametersBag:               parameterBag,
			HashGenerator:               hashGenerator,
			IsSkipCheckOfHashForRequest: isSkipCheckOfHashForRequest,
		},
	}
}

func (u *UpdatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	u.handleHeaders(w, r)
	metrics, err := u.getMetricFromRequest(r)

	if err != nil {
		if err.Error() == "unknown metric type" {
			http.Error(w, err.Error(), http.StatusNotImplemented)
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = u.upsert(r.Context(), metrics)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (u *UpdatesHandler) upsert(ctx context.Context, metrics []Metrics) (err error) {
	var metricsToUpsert []interface{}
	for _, metric := range metrics {
		if MetricTypeGauge == metric.MType {
			metricsToUpsert = append(metricsToUpsert, GaugeMetric{Name: metric.ID, Value: *metric.Value})
		}
		if MetricTypeCounter == metric.MType {
			prevValue, _ := u.Repository.GetCounter(metric.ID)
			metricsToUpsert = append(metricsToUpsert, CounterMetric{Name: metric.ID, Value: prevValue + *metric.Delta})
		}
	}

	return u.Repository.UpsertMany(ctx, metricsToUpsert)
}

func (u *UpdatesHandler) getMetricFromRequest(r *http.Request) (metrics []Metrics, err error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return metrics, errors.New("incorrect Content-Type")
	}
	metrics, err = u.getMetricsFromJSON(r)

	if err != nil {
		return metrics, err
	}

	for _, metric := range metrics {
		_, err = u.validateMetric(&metric)

		if err != nil {
			return metrics, err
		}
	}

	return metrics, nil
}

func (u *UpdatesHandler) getMetricsFromJSON(r *http.Request) (metrics []Metrics, err error) {
	var body []byte

	body, err = io.ReadAll(r.Body)
	if err != nil {
		return metrics, err
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &metrics)
	if err != nil {
		return metrics, err
	}

	return metrics, nil
}
