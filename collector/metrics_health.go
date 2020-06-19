package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
)

// check interface
var _ Scraper = ScrapeHealth{}

var (
	healthInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "health"),
		"components status(0 for error, 1 for success).",
		[]string{"name"}, nil,
	)
)

type ScrapeHealth struct{}

// Name of the Scraper. Should be unique.
func (ScrapeHealth) Name() string {
	return "health"
}

// Help describes the role of the Scraper.
func (ScrapeHealth) Help() string {
	return "Collect the health ref work"
}

// Scrape collects data from client and sends it over channel as prometheus metric.
func (ScrapeHealth) Scrape(client *HarborClient, ch chan<- prometheus.Metric) error {
	var data healthJson
	url := "/health"
	body, err := client.request(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	for _, v := range data.CS {
		var status float64 = 0
		if v.Status == "healthy" {
			status = 1
		}
		ch <- prometheus.MustNewConstMetric(healthInfo, prometheus.GaugeValue,
			status, v.Name)
	}

	return nil
}

type healthJson struct {
	Status string     `json:"status"`
	CS     []csStatus `json:"components"`
}

type csStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
