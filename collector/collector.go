package collector

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/foomo/pagespeed_exporter/scraper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"time"
)

type collector struct {
	targets       []string
	scrapeService scraper.Service
}

func NewCollector(targets []string) (coll prometheus.Collector, err error) {
	return collector{
		targets:       targets,
		scrapeService: scraper.New(),
	}, nil
}

// Describe implements Prometheus.Collector.
func (c collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

// Collect implements Prometheus.Collector.
func (c collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	scrapes, errScrape := c.scrapeService.Scrape(c.targets)

	if errScrape != nil {
		logrus.WithError(errScrape).Warn("Could not scrape targets")
		ch <- prometheus.NewInvalidMetric(prometheus.NewDesc("pagespeed_error", "Error scraping target", nil, nil), errScrape)
		return
	}

	for _, s := range scrapes {
		spew.Dump(s)
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("pagespeed_scrape_duration_seconds", "Total Pagespeed time scrape took for all targets.", nil, nil),
		prometheus.GaugeValue,
		float64(time.Since(start).Seconds()))
}
