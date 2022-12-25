package agent

import (
	"fmt"
	"log"
	"net/http"
	"os"
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
	url := c.getUpdateMetricUrl(metricType, metricName, metricValue)

	c.loggerInfo.Printf("client are making request. url: %s\n", url)
	post, err := http.Post(url, "text/plain", nil)

	if err != nil {
		c.loggerWarning.Printf("error while sending the metrics to server. Error: %s\n", err.Error())
		return err
	}
	defer post.Body.Close()

	if post.StatusCode != http.StatusOK {
		err = fmt.Errorf("error while sending the metrics to server. Status: %d\n", post.StatusCode)
		c.loggerWarning.Printf(err.Error())
		return err
	}

	return nil
}

func (c *Client) getUpdateMetricUrl(metricType, metricName, metricValue string) (url string) {
	return fmt.Sprintf("%s/update/%s/%s/%s", c.MetricAggregatorService, metricType, metricName, metricValue)
}
