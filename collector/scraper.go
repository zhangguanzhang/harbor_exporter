package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

//https://github.com/prometheus/mysqld_exporter/blob/master/collector/scraper.go

// Scraper is minimal interface that let's you add new prometheus metrics to mysqld_exporter.
type Scraper interface {
	// Name of the Scraper. Should be unique.
	Name() string

	// Help describes the role of the Scraper.
	// Example: "Collect from SHOW ENGINE INNODB STATUS"
	Help() string

	// Scrape collects data from client and sends it over channel as prometheus metric.
	Scrape(client *HarborClient, ch chan<- prometheus.Metric) error
}
