package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	valid "github.com/asaskevich/govalidator"
	"io"
	"net/http"
	"strconv"
)

type Handler struct {
	Repository                  IRepository
	ParametersBag               IParametersBag
	HashGenerator               IHashGenerator
	IsSkipCheckOfHashForRequest bool
}

func (h *Handler) handleHeaders(w http.ResponseWriter, r *http.Request) {
	headerAccept := r.Header.Get("Accept")

	if headerAccept != "" {
		w.Header().Set("Content-Type", headerAccept)
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}
}

func (h *Handler) getMetricFromRequest(r *http.Request) (metric Metrics, err error) {
	if r.Header.Get("Content-Type") == "application/json" {
		metric, err = h.getMetricsFromJSON(r)
	} else {
		metric, err = h.getMetricFromURL(r)
	}

	if err != nil {
		return metric, err
	}

	_, err = h.validateMetric(&metric)

	if err != nil {
		return metric, err
	}

	return metric, nil
}

func (h *Handler) getMetricsFromJSON(r *http.Request) (metric Metrics, err error) {
	var body []byte

	body, err = io.ReadAll(r.Body)
	if err != nil {
		return metric, err
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &metric)
	if err != nil {
		return metric, err
	}

	return metric, nil
}

func (h *Handler) getMetricFromURL(r *http.Request) (metric Metrics, err error) {
	metric.MType = h.ParametersBag.GetURLParam(r, paramNameMetricType)
	metric.ID = h.ParametersBag.GetURLParam(r, paramNameMetricName)

	metricValue := h.ParametersBag.GetURLParam(r, paramNameMetricValue)
	if metricValue == "" {
		return metric, nil
	}

	switch metric.MType {
	case MetricTypeGauge:
		var value float64
		value, err = strconv.ParseFloat(metricValue, 64)
		metric.Value = &value
	case MetricTypeCounter:
		var delta int64
		delta, err = strconv.ParseInt(metricValue, 10, 64)
		metric.Delta = &delta
	default:
		err = errors.New("unknown metric type")
	}

	return metric, err
}

func (h *Handler) handleBody(w http.ResponseWriter, metric Metrics, acceptHeader string) (err error) {
	actualMetric, err := h.getActualMetric(metric)
	if err != nil {
		return err
	}

	if h.HashGenerator != nil {
		sign, err := h.getSign(actualMetric)
		if err != nil {
			return err
		}
		actualMetric.Hash = sign
	}

	if acceptHeader == "application/json" {
		body, err := json.Marshal(actualMetric)
		if err != nil {
			return err
		}

		w.Write(body)

		return nil
	}

	if actualMetric.Value != nil {
		w.Write([]byte(fmt.Sprintf("%.3f", *actualMetric.Value)))
	} else {
		w.Write([]byte(fmt.Sprintf("%d", *actualMetric.Delta)))
	}

	return nil
}

func (h *Handler) getActualMetric(metric Metrics) (Metrics, error) {
	var actualMetric Metrics

	switch metric.MType {
	case MetricTypeCounter:
		v, err := h.Repository.GetCounter(metric.ID)
		if err != nil {
			return Metrics{}, err
		}
		actualMetric = Metrics{
			ID:    metric.ID,
			MType: MetricTypeCounter,
			Delta: &v,
		}
	case MetricTypeGauge:
		v, err := h.Repository.GetGauge(metric.ID)
		if err != nil {
			return Metrics{}, err
		}
		actualMetric = Metrics{
			ID:    metric.ID,
			MType: MetricTypeGauge,
			Value: &v,
		}
	default:
		return Metrics{}, errors.New("trying to get metric with unknown type, there is an error in logic of checking request")
	}

	return actualMetric, nil
}

func (h *Handler) validateMetric(metric *Metrics) (bool, error) {
	if h.IsSkipCheckOfHashForRequest {
		return valid.ValidateStruct(metric)
	}

	if metric.Hash == "" {
		return false, fmt.Errorf("hash is not correct")
	}

	sign, err := h.getSign(*metric)
	if err != nil {
		return false, err
	}

	if !h.HashGenerator.Equal(metric.Hash, sign) {
		return false, fmt.Errorf("hash is not correct")
	}

	return valid.ValidateStruct(metric)
}

func (h *Handler) getSign(metric Metrics) (sign string, err error) {
	switch metric.MType {
	case MetricTypeCounter:
		stringToHash := fmt.Sprintf("%s:%s:%d", metric.ID, metric.MType, *metric.Delta)
		return h.HashGenerator.Generate(stringToHash)

	default:
		stringToHash := fmt.Sprintf("%s:%s:%f", metric.ID, metric.MType, *metric.Value)
		return h.HashGenerator.Generate(stringToHash)
	}
}
