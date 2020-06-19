package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
)

// check interface
var _ Scraper = ScrapeSystemInfo{}

const (
	statisticsUrl = "/statistics"
)

var (
	projectCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "project_count_total"),
		"projects number relevant to the user", []string{"type"}, nil)
	repoCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "repo_count_total"),
		"repositories number relevant to the user",
		[]string{"type"}, nil,
	)
)

type ScrapeStatistics struct{}

// Name of the Scraper. Should be unique.
func (ScrapeStatistics) Name() string {
	return "statistics"
}

// Help describes the role of the Scraper.
func (ScrapeStatistics) Help() string {
	return "Collect the statistics"
}


// Scrape collects data from client and sends it over channel as prometheus metric.
func (ScrapeStatistics) Scrape(client *HarborClient, ch chan<- prometheus.Metric) error {
	var data statisticsJson
	body, err := client.request(statisticsUrl)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}


	ch <- prometheus.MustNewConstMetric(
		projectCount, prometheus.GaugeValue, data.TotalProjectCount, "total",
	)

	ch <- prometheus.MustNewConstMetric(
		projectCount, prometheus.GaugeValue, data.PublicProjectCount, "public",
	)

	ch <- prometheus.MustNewConstMetric(
		projectCount, prometheus.GaugeValue, data.PrivateProjectCount, "private",
	)

	ch <- prometheus.MustNewConstMetric(
		repoCount, prometheus.GaugeValue, data.PublicRepoCount, "public",
	)

	ch <- prometheus.MustNewConstMetric(
		repoCount, prometheus.GaugeValue, data.TotalRepoCount, "total",
	)

	ch <- prometheus.MustNewConstMetric(
		repoCount, prometheus.GaugeValue, data.PrivateRepoCount, "private",
	)


	return nil
}

type statisticsJson struct {
	PrivateProjectCount float64 `json:"private_project_count"`
	PrivateRepoCount    float64 `json:"private_repo_count"`
	PublicProjectCount  float64 `json:"public_project_count"`
	PublicRepoCount     float64 `json:"public_repo_count"`
	TotalProjectCount   float64 `json:"total_project_count"`
	TotalRepoCount      float64 `json:"total_repo_count"`
}