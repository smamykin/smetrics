package agent

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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

func (c *Client) SendMetrics(metricType string, metricName string, metricValue string) {
	url := c.MetricAggregatorService + getMetricAggregatorPath(metricType, metricName, metricValue)

	c.loggerInfo.Printf("client are making request. url: %s\n", url)
	post, err := http.Post(url, "text/plain", strings.NewReader(""))

	if err != nil {
		c.loggerWarning.Printf("error while sending the metrics to server. Error: %s\n", err.Error())
	}
	if post.StatusCode != 200 {
		c.loggerWarning.Printf("error while sending the metrics to server. Status: %d\n", post.StatusCode)
	}

	defer post.Body.Close()
	_, err = io.Copy(io.Discard, post.Body)
	if err != nil {
		panic(fmt.Errorf("cannot read the response body. url: %s", url))
	}
}

func getMetricAggregatorPath(metricType string, metricName string, metricValue string) (path string) {
	//tmpl := "/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>"
	return fmt.Sprintf("/update/%s/%s/%s", metricType, metricName, metricValue)
}
