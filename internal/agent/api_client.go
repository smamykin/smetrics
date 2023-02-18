package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/smamykin/smetrics/internal/utils"
	"net/http"
	"strconv"
)

type job struct {
	metrics []IMetric
	doneCh  chan error
}

func NewClient(logger *zerolog.Logger, metricAggregatorService string, key string, rateLimit int) *Client {
	jobCh := make(chan job)
	result := &Client{
		MetricAggregatorService: metricAggregatorService,
		logger:                  logger,
		jobCh:                   jobCh,
	}

	if key != "" {
		result.hashGenerator = utils.NewHashGenerator(key)
	}

	if rateLimit <= 0 {
		return result
	}

	for i := 0; i < rateLimit; i++ {
		go func() {
			for job := range jobCh {
				err := result.makeRequest(job.metrics)
				job.doneCh <- err
			}
		}()
	}

	return result
}

type Client struct {
	MetricAggregatorService string
	logger                  *zerolog.Logger
	hashGenerator           IHashGenerator
	jobCh                   chan job
}

func (c *Client) SendMetrics(metrics []IMetric) error {
	doneCh := make(chan error)
	defer func() {
		close(doneCh)
	}()
	c.jobCh <- job{metrics, doneCh}

	return <-doneCh
}

func (c *Client) createRequestBody(metrics []IMetric) (body []byte, err error) {
	var result []Metrics
	for _, metric := range metrics {
		m := Metrics{
			MType: metric.GetType(),
			ID:    metric.GetName(),
		}
		switch m.MType {
		case MetricTypeGauge:
			value, err := strconv.ParseFloat(metric.String(), 64)
			if err != nil {
				return body, fmt.Errorf("unable to parse gauge value: %w", err)
			}
			m.Value = &value
			err = c.signMetricWithHash(&m)
			if err != nil {
				return body, err
			}
		case MetricTypeCounter:
			value, err := strconv.ParseInt(metric.String(), 10, 64)
			if err != nil {
				return body, fmt.Errorf("unable to parse counter value: %w", err)
			}
			m.Delta = &value

			err = c.signMetricWithHash(&m)
			if err != nil {
				return body, err
			}
		default:
			return body, errors.New("unknown type of the metric")
		}

		result = append(result, m)
	}
	return json.Marshal(result)
}

func (c *Client) signMetricWithHash(metrics *Metrics) (err error) {
	if c.hashGenerator == nil {
		return nil
	}
	var sign string

	switch metrics.MType {
	case MetricTypeGauge:
		stringToHash := fmt.Sprintf("%s:gauge:%f", metrics.ID, *metrics.Value)
		sign, err = c.hashGenerator.Generate(stringToHash)
	case MetricTypeCounter:
		stringToHash := fmt.Sprintf("%s:counter:%d", metrics.ID, *metrics.Delta)
		sign, err = c.hashGenerator.Generate(stringToHash)
	default:
		err = errors.New("unknown type of the metric")
	}

	if err != nil {
		return fmt.Errorf("cannot create hash for metric %v. error: %w", metrics, err)
	}

	metrics.Hash = sign

	return nil
}

func (c *Client) makeRequest(metrics []IMetric) error {
	body, err := c.createRequestBody(metrics)
	if err != nil {
		c.logger.Warn().Msgf("error while sending the metrics to server. Error: %s\n", err.Error())
		return err
	}
	url := fmt.Sprintf("%s/updates/", c.MetricAggregatorService)

	c.logger.Info().Msgf("client are making request. url: %s, body: %s \n", url, string(body))

	post, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		c.logger.Warn().Msgf("error while sending the metrics to server. Error: %s\n", err.Error())
		return err
	}
	defer post.Body.Close()

	if post.StatusCode != http.StatusOK {
		err = fmt.Errorf("error while sending the metrics to server. Status: %d", post.StatusCode)
		c.logger.Warn().Err(err).Msg("")
		return err
	}

	return nil
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type IHashGenerator interface {
	Generate(stringToHash string) (string, error)
}
