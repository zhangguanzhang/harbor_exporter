package collector

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

// check interface
var _ Scraper = ScrapeUsers{}

const (
	usersUrl = "/users"
)

var (
	usersRefInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "ref_work", "users"),
		"test the users api ref work status(0 for error, 1 for success).",
		[]string{"ref", "method"}, nil,
	)
)

type usersJson struct {
	UID int `json:"user_id"`
}

type ScrapeUsers struct{}

// Name of the Scraper. Should be unique.
func (ScrapeUsers) Name() string {
	return "users"
}

// Help describes the role of the Scraper.
func (ScrapeUsers) Help() string {
	return "Collect the users api work, user have admin role"
}

// Scrape collects data from client and sends it over channel as prometheus metric.
func (s ScrapeUsers) Scrape(client *HarborClient, ch chan<- prometheus.Metric) error {
	var err error

	err = users(client, ch)
	if err != nil {
		return err
	}

	err = userCurrent(client, ch)
	if err != nil {
		return err
	}

	return nil
}

func users(client *HarborClient, ch chan<- prometheus.Metric) error {
	var data []usersJson
	url := usersUrl + "?page_size=1"
	body, err := client.request(url)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(body, &data); err != nil {
		return err
	}

	if len(data) != 1 || data[0].UID == 0 {
		return fmt.Errorf("cannot find a user id by %s", url)
	}

	ch <- prometheus.MustNewConstMetric(usersRefInfo, prometheus.GaugeValue,
		1, usersUrl, "GET")

	var result usersJson

	url = usersUrl + "/" + strconv.Itoa(data[0].UID)
	body, err = client.request(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.UID == 0 {
		return fmt.Errorf("cannot find the user by %s", url)
	}

	ch <- prometheus.MustNewConstMetric(usersRefInfo, prometheus.GaugeValue,
		1, "/users/{user_id}", "GET")

	return nil
}

func userCurrent(client *HarborClient, ch chan<- prometheus.Metric) error {
	var result usersJson

	url := usersUrl + "/current"
	body, err := client.request(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.UID == 0 {
		return fmt.Errorf("cannot find the current info by %s", url)
	}

	ch <- prometheus.MustNewConstMetric(usersRefInfo, prometheus.GaugeValue,
		1, url, "GET")

	// TODO
	//  /users/current/permissions will be [] default

	return nil
}
