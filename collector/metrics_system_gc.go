package collector

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// check interface
var _ Scraper = ScrapeGc{}


var (
	gcRefInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ref_work", "gc"),
		"test the replication ref work status(0 for error, 1 for success).",
		[]string{"ref", "method"}, nil,
	)

)

type ScrapeGc struct{}

// Name of the Scraper. Should be unique.
func (ScrapeGc) Name() string {
	return "systemgc"
}

// Help describes the role of the Scraper.
func (ScrapeGc) Help() string {
	return "Collect the systemgc ref work"
}

// Scrape collects data from client and sends it over channel as prometheus metric.
func (ScrapeGc) Scrape(client *HarborClient, ch chan<- prometheus.Metric) error {
	var data []idJson
	url := "/system/gc"
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

	ch <- prometheus.MustNewConstMetric(gcRefInfo, prometheus.GaugeValue,
		1, "/system/gc", "GET")

	return nil
}



