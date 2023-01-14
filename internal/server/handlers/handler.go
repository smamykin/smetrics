package handlers

import (
	"encoding/json"
	"errors"
	valid "github.com/asaskevich/govalidator"
	"io"
	"net/http"
	"strconv"
)

type Handler struct {
	Repository    IRepository
	ParametersBag IParametersBag
}

func (h *Handler) handleHeaders(w http.ResponseWriter, r *http.Request) {
	headerAccept := r.Header.Get("Accept")

	if headerAccept != "" {
		w.Header().Set("Content-Type", headerAccept)
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}
	return
}

func (h *Handler) getMetricFromRequest(r *http.Request) (metric Metrics, err error) {
	if r.Header.Get("Content-Type") == "application/json" {
		metric, err = h.getMetricsFromJson(r)
		return
	}

	metric, err = h.getMetricFromUrl(r)
	if err != nil {
		return
	}

	metricValue := h.ParametersBag.GetUrlParam(r, paramNameMetricValue)
	if metricValue == "" {
		return
	}

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
		return
	}

	_, err = valid.ValidateStruct(&metric)
	return
}

func (h *Handler) getMetricsFromJson(r *http.Request) (metric Metrics, err error) {
	var body []byte

	defer r.Body.Close()
	body, err = io.ReadAll(r.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &metric)
	if err != nil {
		return
	}

	_, err = valid.ValidateStruct(&metric)

	return
}

func (h *Handler) getMetricFromUrl(r *http.Request) (metric Metrics, err error) {
	metric.MType = h.ParametersBag.GetUrlParam(r, paramNameMetricType)
	metric.ID = h.ParametersBag.GetUrlParam(r, paramNameMetricName)

	metricValue := h.ParametersBag.GetUrlParam(r, paramNameMetricValue)
	if metricValue == "" {
		return
	}

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
		return
	}

	_, err = valid.ValidateStruct(&metric)

	return
}
