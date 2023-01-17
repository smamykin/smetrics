package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	valid "github.com/asaskevich/govalidator"
	"io"
	"math"
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

	metric, err = h.getMetricFromURL(r)
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

func (h *Handler) getMetricFromURL(r *http.Request) (metric Metrics, err error) {
	metric.MType = h.ParametersBag.GetURLParam(r, paramNameMetricType)
	metric.ID = h.ParametersBag.GetURLParam(r, paramNameMetricName)

	metricValue := h.ParametersBag.GetURLParam(r, paramNameMetricValue)
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

	if err != nil {
		return
	}

	_, err = valid.ValidateStruct(&metric)

	return
}

func (h *Handler) handleBody(w http.ResponseWriter, metric Metrics, acceptHeader string) (err error) {
	var actualMetric Metrics

	switch metric.MType {
	case metricTypeCounter:
		v, err := h.Repository.GetCounter(metric.ID)
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
		v, err := h.Repository.GetGauge(metric.ID)
		if err != nil {
			return err
		}
		actualMetric = Metrics{
			ID:    metric.ID,
			MType: metricTypeGauge,
			Value: &v,
		}
	default:
		return errors.New("trying to get metric with unknown type, there is an error in logic of checking request")
	}

	if acceptHeader == "application/json" {
		body, err := json.Marshal(actualMetric)
		if err != nil {
			return err
		}

		w.Write(body)

		return nil
	}

	roundedValue := math.Round(*actualMetric.Value*1000) / 1000
	w.Write([]byte(fmt.Sprintf("%v", roundedValue)))

	return nil
}
