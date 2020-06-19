package collector

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

// check interface
var _ Scraper = ScrapeReplication{}


var (
	replicationRefInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ref_work", "replication"),
		"test the replication ref work status(0 for error, 1 for success).",
		[]string{"ref", "method"}, nil,
	)
)

type ScrapeReplication struct{}

// Name of the Scraper. Should be unique.
func (ScrapeReplication) Name() string {
	return "replication"
}

// Help describes the role of the Scraper.
func (ScrapeReplication) Help() string {
	return "Collect the replication ref work"
}

// Scrape collects data from client and sends it over channel as prometheus metric.
func (ScrapeReplication) Scrape(client *HarborClient, ch chan<- prometheus.Metric) error {
	var data []idJson
	url := "/replication/policies?page_size=1"
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

	ch <- prometheus.MustNewConstMetric(replicationRefInfo, prometheus.GaugeValue,
		1, "/replication/policies", "GET")

	var policy idJson
	url = "/replication/executions?page=1&page_size=1&policy_id=" + strconv.Itoa(data[0].ID)
	body, err = client.request(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &policy); err != nil {
		return err
	}

	if policy.ID == 0 {
		return errors.Wrap(resultErr, url)
	}

	ch <- prometheus.MustNewConstMetric(replicationRefInfo, prometheus.GaugeValue,
		1, "/replication/executions", "GET")

	var adadapt []string
	url = "/replication/adapters"
	body, err = client.request(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &adadapt); err != nil {
		return err
	}

	if len(adadapt) == 0 {
		return errors.Wrap(resultErr, url)
	}

	ch <- prometheus.MustNewConstMetric(replicationRefInfo, prometheus.GaugeValue,
		1, url, "GET")

	return nil
}



