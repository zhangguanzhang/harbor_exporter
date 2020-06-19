package collector

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// check interface
var _ Scraper = ScrapeQuotas{}

const (
	volumesUrl = "/systeminfo/volumes"
)

var (
	systemVolumes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "system_volumes_bytes"),
		"Get system volume info (total/free size).", []string{"type"}, nil)

)

type ScrapeQuotas struct{}

// Name of the Scraper. Should be unique.
func (ScrapeQuotas) Name() string {
	return "systeminfoVolumes"
}

// Help describes the role of the Scraper.
func (ScrapeQuotas) Help() string {
	return "Collect the systeminfoVolumes, user must have admin"
}


// Scrape collects data from client and sends it over channel as prometheus metric.
func (ScrapeQuotas) Scrape(client *HarborClient, ch chan<- prometheus.Metric) error {
	var data quotasJson
	body, err := client.request(volumesUrl)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if data.Storage.Total == 0 {
		return errors.Wrap(resultErr, volumesUrl)
	}

	ch <- prometheus.MustNewConstMetric(
		systemVolumes, prometheus.GaugeValue, data.Storage.Total, "total",
	)

	ch <- prometheus.MustNewConstMetric(
		systemVolumes, prometheus.GaugeValue, data.Storage.Free, "free",
	)

	ch <- prometheus.MustNewConstMetric(
		systemVolumes, prometheus.GaugeValue, data.Storage.Total - data.Storage.Free, "used",
	)

	return nil
}

type quotasJson struct {
	Storage struct {
		Total float64 `json:"total"`
		Free  float64 `json:"free"`
	} `json:"storage"`
}