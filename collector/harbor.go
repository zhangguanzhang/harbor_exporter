package collector

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	Scrapers = map[Scraper]bool{
		ScrapeSystemInfo{}:  true,
		ScrapeStatistics{}:  true,
		ScrapeQuotas{}:      true,
		ScrapeHealth{}:      true,
		ScrapeProjects{}:    true,
		ScrapeUsers{}:       true,
		ScrapeLogs{}:        true,
		ScrapeReplication{}: false,
		ScrapeGc{}:          true,
	}

	// TODO
	//  tags always return full tag, see https://github.com/goharbor/harbor/issues/12279

	resultErr = errors.New("cannot find data, maybe json is nil at")
)

type HarborOpts struct {
	Url      string
	Username string
	password string
	UA       string
	Timeout  time.Duration
	Insecure bool
}

type HarborClient struct {
	Client *http.Client
	Opts   *HarborOpts
}

// could use for member and repos
type subInsJson struct {
	Id        int `json:"id"`
	ProjectID int `json:"project_id"`
}

type idJson struct {
	ID int `json:"id"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// use after set Opts
func (o *HarborOpts) AddFlag() {
	flag.StringVar(&o.Url, "harbor-server", "", "HTTP API address of a harbor server or agent. (prefix with https:// to connect over HTTPS)")
	flag.StringVar(&o.Username, "harbor-user", "admin", "harbor username")
	flag.StringVar(&o.password, "harbor-pass", "password", "harbor password")
	flag.StringVar(&o.UA, "harbor-ua", "harbor_exporter", "user agent of the harbor http client")
	flag.DurationVar(&o.Timeout, "time-out", time.Millisecond*1600, "Timeout on HTTP requests to the harbor API.")
	flag.BoolVar(&o.Insecure, "insecure", false, "Disable TLS host verification.")
}

func (h *HarborClient) request(endpoint string) ([]byte, error) {
	url := h.Opts.Url + endpoint
	log.Debugf("request url %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(h.Opts.Username, h.Opts.password)
	req.Header.Set("User-Agent", h.Opts.UA)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error handling request for %s http-statuscode: %s", endpoint, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (h *HarborClient) Ping() (bool, error) {
	req, err := http.NewRequest("GET", h.Opts.Url+"/configurations", nil)
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(h.Opts.Username, h.Opts.password)

	resp, err := h.Client.Do(req)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, errors.New("username or password incorrect")
}
