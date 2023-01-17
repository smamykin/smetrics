package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func NewClient(metricAggregatorService string) *Client {
	return &Client{
		MetricAggregatorService: metricAggregatorService,
		loggerWarning:           log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime),
		loggerInfo:              log.New(os.Stdout, "INFO:    ", log.Ldate|log.Ltime),
	}

}

type Client struct {
	MetricAggregatorService string
	loggerWarning           *log.Logger
	loggerInfo              *log.Logger
}

func (c *Client) SendMetrics(metricType, metricName, metricValue string) error {

	body, err := c.getUpdateMetrics(metricType, metricName, metricValue)
	if err != nil {
		c.loggerWarning.Printf("error while sending the metrics to server. Error: %s\n", err.Error())
		return err
	}

	c.loggerInfo.Printf("client are making request. url: %s, body: %s \n", "/update/", string(body))
	url := fmt.Sprintf("%s/update/", c.MetricAggregatorService)
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

func (c *Client) getUpdateMetrics(metricType, metricName, metricValue string) (body []byte, err error) {
	metrics := Metrics{
		MType: metricType,
		ID:    metricName,
	}
	if metrics.MType == MetricTypeGauge {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return body, err
		}
		metrics.Value = &value
	}
	if metrics.MType == MetricTypeCounter {
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return body, err
		}
		metrics.Delta = &value
	}

	body, err = json.Marshal(metrics)

	return body, err
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
