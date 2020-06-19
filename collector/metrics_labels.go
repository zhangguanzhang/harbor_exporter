package collector

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// check interface
var _ Scraper = ScrapeLables{}


var (
	labelsRefInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ref_work", "labels"),
		"test the labels ref work status(0 for error, 1 for success).",
		[]string{"ref", "method"}, nil,
	)

)

type ScrapeLables struct{}

// Name of the Scraper. Should be unique.
func (ScrapeLables) Name() string {
	return "labels"
}

// Help describes the role of the Scraper.
func (ScrapeLables) Help() string {
	return "Collect the labels ref work"
}

// Scrape collects data from client and sends it over channel as prometheus metric.
func (ScrapeLables) Scrape(client *HarborClient, ch chan<- prometheus.Metric) error {
	var data []idJson
	url := "/labels?scope=g&pagesize=1"
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

	ch <- prometheus.MustNewConstMetric(labelsRefInfo, prometheus.GaugeValue,
		1, "/labels", "GET")

	return nil
}



