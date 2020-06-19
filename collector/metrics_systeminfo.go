package collector

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

// check interface
var _ Scraper = ScrapeSystemInfo{}

const (
	systemInfoUrl = "/systeminfo"
)

var (
	HarborVersion = "" // some version is not contains version number, so could by be override
	harborInfo = prometheus.NewDesc(prometheus.BuildFQName(namespace, "version", "info"),
		"harbor system info",
		[]string{"registry_url", "project_creation_restriction", "self_registration", "version"}, nil)

)

type ScrapeSystemInfo struct{}

// Name of the Scraper. Should be unique.
func (ScrapeSystemInfo) Name() string {
	return "systeminfo"
}

// Help describes the role of the Scraper.
func (ScrapeSystemInfo) Help() string {
	return "Collect the general system info"
}


// Scrape collects data from client and sends it over channel as prometheus metric.
func (ScrapeSystemInfo) Scrape(client *HarborClient, ch chan<- prometheus.Metric) error {
	var data systemInfoJson
	body, err := client.request(systemInfoUrl)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if len(data.RegistryURL) == 0 {
		return errors.Wrap(resultErr, systemInfoUrl)
	}

	if HarborVersion != "" {
		data.HarborVersion = HarborVersion
	}

	ch <- prometheus.MustNewConstMetric(harborInfo, prometheus.GaugeValue, 1,
		data.RegistryURL, data.ProjectCreationRestriction, strconv.FormatBool(data.SelfRegistration), data.HarborVersion)

	return nil
}

type systemInfoJson struct {
	//WithNotary                  bool   `json:"with_notary"`
	//WithClair                   bool   `json:"with_clair"`
	//WithAdmiral                 bool   `json:"with_admiral"`
	//AdmiralEndpoint             string `json:"admiral_endpoint"`
	//AuthMode                    string `json:"auth_mode"`
	RegistryURL                 string `json:"registry_url"`
	//ExternalURL                 string `json:"external_url"`
	ProjectCreationRestriction  string `json:"project_creation_restriction"`
	SelfRegistration            bool   `json:"self_registration"`
	//HasCaRoot                   bool   `json:"has_ca_root"`
	HarborVersion               string `json:"harbor_version"`
	//RegistryStorageProviderName string `json:"registry_storage_provider_name"` //
	//ReadOnly                    bool   `json:"read_only"`
	//WithChartmuseum             bool   `json:"with_chartmuseum"` //
}