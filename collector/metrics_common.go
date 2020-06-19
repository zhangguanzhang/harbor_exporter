package collector

//import (
//
//	"github.com/prometheus/client_golang/prometheus"
//
//)
//
//// check interface
//var _ Scraper = ScrapeCommon{}
//
//
//var (
//	commonRefInfo = prometheus.NewDesc(
//		prometheus.BuildFQName(namespace, "ref_work", "common"),
//		"test the common api ref work status(0 for error, 1 for success).",
//		[]string{"ref", "method"}, nil,
//	)
//
//)
//
//type ScrapeCommon struct{}
//
//// Name of the Scraper. Should be unique.
//func (ScrapeCommon) Name() string {
//	return "common"
//}
//
//// Help describes the role of the Scraper.
//func (ScrapeCommon) Help() string {
//	return "Collect the common info in most version"
//}
//
//// Scrape collects data from client and sends it over channel as prometheus metric.
//func (ScrapeCommon) Scrape(client *HarborClient, ch chan<- prometheus.Metric) error {
//
//	return nil
//}
