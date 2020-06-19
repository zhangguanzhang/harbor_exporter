package collector

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// check interface
var _ Scraper = ScrapeLogs{}


var (
	logRefInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ref_work", "logs"),
		"test the logs ref work status(0 for error, 1 for success).",
		[]string{"ref", "method"}, nil,
	)

)

type ScrapeLogs struct{}

// Name of the Scraper. Should be unique.
func (ScrapeLogs) Name() string {
	return "logs"
}

// Help describes the role of the Scraper.
func (ScrapeLogs) Help() string {
	return "Collect the logs ref work"
}

// Scrape collects data from client and sends it over channel as prometheus metric.
func (ScrapeLogs) Scrape(client *HarborClient, ch chan<- prometheus.Metric) error {
	var data []logJson
	url := "/logs?page_size=1"
	body, err := client.request(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if len(data) != 1 || data[0].ID == 0 {
		return errors.Wrap(resultErr, url)
	}

	ch <- prometheus.MustNewConstMetric(logRefInfo, prometheus.GaugeValue,
		1, "/logs", "GET")

	return nil
}


type logJson struct {
	ID int `json:"log_id"`
}

