package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

const (
	paramNameMetricType  = "metricType"
	paramNameMetricName  = "metricName"
	paramNameMetricValue = "metricValue"
)

type UpdateHandler struct {
	Repository    IRepository
	ParametersBag IParametersBag
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

func (u *UpdateHandler) handleHeaders(w http.ResponseWriter, r *http.Request) {
	headerAccept := r.Header.Get("Accept")

	if headerAccept != "" {
		w.Header().Set("Content-Type", headerAccept)
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}
	return
}

func (u *UpdateHandler) getMetricFromRequest(r *http.Request) (metric Metrics, err error) {
	if r.Header.Get("Content-Type") == "application/json" {
		var body []byte

		defer r.Body.Close()
		body, err = io.ReadAll(r.Body)
		if err != nil {
			return
		}

		err = json.Unmarshal(body, &metric)

		return
	}

	metric.MType = u.ParametersBag.GetUrlParam(r, paramNameMetricType)
	metric.ID = u.ParametersBag.GetUrlParam(r, paramNameMetricName)
	metricValue := u.ParametersBag.GetUrlParam(r, paramNameMetricValue)

	switch metric.MType {
	case metricTypeGauge:
		var value float64
		value, err = strconv.ParseFloat(metricValue, 64)
		metric.Value = &value
	case metricTypeCounter:
		var delta int64
		delta, err = strconv.ParseInt(metricValue, 10, 64)
		metric.Delta = &delta
	default:
		err = errors.New("unknown metric type")
	}

	return
}

func (u *UpdateHandler) upsert(metric Metrics) (err error) {
	if metricTypeGauge == metric.MType {
		return u.Repository.UpsertGauge(GaugeMetric{Name: metric.ID, Value: *metric.Value})
	}

	if metricTypeCounter == metric.MType {
		prevValue, _ := u.Repository.GetCounter(metric.ID)

		return u.Repository.UpsertCounter(CounterMetric{Name: metric.ID, Value: prevValue + *metric.Delta})
	}

	panic("trying to upsert metric with unknown type, there is an error in logic of checking request")
}

func (u *UpdateHandler) handleBody(w http.ResponseWriter, metric Metrics, acceptHeader string) (err error) {
	var actualMetric Metrics
	if acceptHeader == "application/json" {
		switch metric.MType {
		case metricTypeCounter:
			v, err := u.Repository.GetCounter(metric.ID)
			if err != nil {
				return err
			}
			value := float64(v)

			actualMetric = Metrics{
				ID:    metric.ID,
				MType: metricTypeGauge,
				Value: &value,
			}
		case metricTypeGauge:
			value, err := u.Repository.GetGauge(metric.ID)
			if err != nil {
				return err
			}
			actualMetric = Metrics{
				ID:    metric.ID,
				MType: metricTypeGauge,
				Value: &value,
			}

		default:
			panic("trying to get metric with unknown type, there is an error in logic of checking request")
		}

		body, err := json.Marshal(actualMetric)
		if err != nil {
			return err
		}

		w.Write(body)
	}

	return nil
}
