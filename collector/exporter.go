package collector

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	//"github.com/prometheus/client_golang/prometheus"
)

const (
	name = "harbor_exporter"
	namespace = "harbor"
	//Subsystem(s).
	exporter = "exporter"
)

func Name() string {
	return name
}



// Verify if Exporter implements prometheus.Collector
var _ prometheus.Collector = (*Exporter)(nil)

// Metric descriptors.
var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, exporter, "collector_duration_seconds"),
		"Collector time duration.",
		[]string{"collector"}, nil,
	)
)

type Exporter struct {
	//ctx      context.Context  //http timeout will work, don't need this
	client   *HarborClient
	scrapers []Scraper
	metrics  Metrics
}


func New(opts *HarborOpts, metrics Metrics, scrapers []Scraper) (*Exporter, error) {
	uri := opts.Url
	if !strings.Contains(uri, "://") {
		uri = "http://" + uri
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid harbor URL: %s", err)
	}
	if u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("invalid harbor URL: %s", uri)
	}

	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	tlsClientConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    rootCAs,
	}

	if opts.Insecure {
		tlsClientConfig.InsecureSkipVerify = true
	}

	user := os.Getenv("HARBOR_USERNAME")
	if user != "" {
		opts.password = user
	}

	pass := os.Getenv("HARBOR_PASSWORD")
	if pass != "" {
		opts.password = pass
	}

	transport := &http.Transport{
		TLSClientConfig: tlsClientConfig,
	}

	hc := &HarborClient{
		Opts: opts,
		Client: &http.Client{
			Timeout: opts.Timeout,
			Transport: transport,
		},
	}

	return &Exporter{
		client: hc,
		metrics: metrics,
		scrapers: scrapers,
	}, nil
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.metrics.TotalScrapes.Desc()
	ch <- e.metrics.Error.Desc()
	e.metrics.ScrapeErrors.Describe(ch)
	ch <- e.metrics.HarborUp.Desc()
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.scrape(ch)

	ch <- e.metrics.TotalScrapes
	ch <- e.metrics.Error
	e.metrics.ScrapeErrors.Collect(ch)
	ch <- e.metrics.HarborUp
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) {
	e.metrics.TotalScrapes.Inc()

	scrapeTime := time.Now()

	if pong, err := e.client.Ping(); pong != true || err != nil {
		log.WithFields(log.Fields{
			"url": e.client.Opts.Url,
			"username": e.client.Opts.Username,
		}).Error(err)
		e.metrics.HarborUp.Set(0)
		e.metrics.Error.Set(1)
	}
	e.metrics.HarborUp.Set(1)
	e.metrics.Error.Set(0)

	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds(), "reach")

	var wg sync.WaitGroup
	defer wg.Wait()
	for _, scraper := range e.scrapers {

		wg.Add(1)
		go func(scraper Scraper) {
			defer wg.Done()
			label := scraper.Name()
			scrapeTime := time.Now()
			if err := scraper.Scrape(e.client, ch); err != nil {
				log.WithField("scraper", scraper.Name()).Error(err)
				e.metrics.ScrapeErrors.WithLabelValues(label).Inc()
				e.metrics.Error.Set(1)
			}
			ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds(), label)
		}(scraper)
	}
}


// Metrics represents exporter metrics which values can be carried between http requests.
type Metrics struct {
	TotalScrapes prometheus.Counter
	ScrapeErrors *prometheus.CounterVec
	Error        prometheus.Gauge
	HarborUp      prometheus.Gauge
}

// NewMetrics creates new Metrics instance.
func NewMetrics() Metrics {
	subsystem := exporter
	return Metrics{
		TotalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "scrapes_total",
			Help:      "Total number of times harbor was scraped for metrics.",
		}),
		ScrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occurred scraping a harbor.",
		}, []string{"collector"}),
		Error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from harbor resulted in an error (1 for error, 0 for success).",
		}),
		HarborUp: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Whether the harbor is up.",
		}),
	}
}