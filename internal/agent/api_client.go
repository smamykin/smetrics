package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/smamykin/smetrics/internal/utils"
	"log"
	"net/http"
	"os"
	"strconv"
)

func NewClient(metricAggregatorService string, key string) *Client {
	result := &Client{
		MetricAggregatorService: metricAggregatorService,
		loggerWarning:           log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime),
		loggerInfo:              log.New(os.Stdout, "INFO:    ", log.Ldate|log.Ltime),
	}

	if key != "" {
		result.hashGenerator = utils.NewHashGenerator(key)
	}

	return result
}

type Client struct {
	MetricAggregatorService string
	loggerWarning           *log.Logger
	loggerInfo              *log.Logger
	hashGenerator           IHashGenerator
}

func (c *Client) SendMetrics(metricType, metricName, metricValue string) error {

	body, err := c.createRequestBody(metricType, metricName, metricValue)
	if err != nil {
		c.loggerWarning.Printf("error while sending the metrics to server. Error: %s\n", err.Error())
		return err
	}
	url := fmt.Sprintf("%s/update/", c.MetricAggregatorService)

	c.loggerInfo.Printf("client are making request. url: %s, body: %s \n", url, string(body))

	post, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		c.loggerWarning.Printf("error while sending the metrics to server. Error: %s\n", err.Error())
		return err
	}
	defer post.Body.Close()

	if post.StatusCode != http.StatusOK {
		err = fmt.Errorf("error while sending the metrics to server. Status: %d", post.StatusCode)
		c.loggerWarning.Println(err.Error())
		return err
	}

	return nil
}

func (c *Client) createRequestBody(metricType, metricName, metricValue string) (body []byte, err error) {
	metrics := Metrics{
		MType: metricType,
		ID:    metricName,
	}
	switch metrics.MType {
	case MetricTypeGauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return body, fmt.Errorf("unable to parse gauge value: %w", err)
		}
		metrics.Value = &value
		err = c.signMetricWithHash(&metrics)
		if err != nil {
			return body, err
		}
	case MetricTypeCounter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return body, fmt.Errorf("unable to parse counter value: %w", err)
		}
		metrics.Delta = &value

		err = c.signMetricWithHash(&metrics)
		if err != nil {
			return body, err
		}
	default:
		return body, errors.New("unknown type of the metric")
	}

	return json.Marshal(metrics)
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
