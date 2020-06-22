package collector

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

// check interface
var _ Scraper = ScrapeRegistries{}

const (
	registryUrl = "/registries"
)

var (
	registriesRefInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "registries", "healthy"),
		" ui /harbor/registries status(0 for error, 1 for success).",
		[]string{"name"}, nil,
	)

)


type ScrapeRegistries struct{}

// Name of the Scraper. Should be unique.
func (ScrapeRegistries) Name() string {
	return "registries"
}

// Help describes the role of the Scraper.
func (ScrapeRegistries) Help() string {
	return "Collect the registries and repos api work"
}

// Scrape collects data from client and sends it over channel as prometheus metric.
func (ScrapeRegistries) Scrape(client *HarborClient, ch chan<- prometheus.Metric) error {
	var data []registryJson
	url := registryUrl
	body, err := client.request(url)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(body, &data); err != nil {
		return err
	}

	if len(data) == 0 {
		return errors.Wrap(resultErr, url)
	}

	for _, v := range data {
		var status float64 = 0
		if strings.Compare("unhealthy", v.Status) != 0 {
			status = 1
		}
		ch <- prometheus.MustNewConstMetric(registriesRefInfo, prometheus.GaugeValue,
			status, v.Name)
	}

	return nil
}


type registryJson struct {
	Name string `json:"name"`
	Status string `json:"status"`
}