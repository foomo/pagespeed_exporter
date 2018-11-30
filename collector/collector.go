package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/pagespeedonline/v5"
	"strconv"
	"strings"
	"time"
)

type collector struct {
	targets       []string
	scrapeService scrapeService
}

func NewCollector(targets []string, googleApiKey string) (coll prometheus.Collector, err error) {
	return collector{
		targets:       targets,
		scrapeService: newPagespeedScrapeService(60*time.Second, googleApiKey),
	}, nil
}

// Describe implements Prometheus.Collector.
func (c collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

// Collect implements Prometheus.Collector.
func (c collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	result, errScrape := c.scrapeService.Scrape(c.targets)

	if errScrape != nil {
		logrus.WithError(errScrape).Warn("Could not scrape targets")
		ch <- prometheus.NewInvalidMetric(prometheus.NewDesc("pagespeed_error", "Error scraping target", nil, nil), errScrape)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("pagespeed_scrape_duration_seconds", "Total Pagespeed time scrape took for all targets.", nil, nil),
		prometheus.GaugeValue,
		float64(time.Since(start).Seconds()))

	for _, scrape := range result {
		collect(scrape, ch)
	}
}

func collect(scrape *Scrape, ch chan<- prometheus.Metric) {
	constLables := prometheus.Labels{
		"target":   scrape.Target,
		"strategy": string(scrape.Strategy),
	}

	r := scrape.Result
	collectLoadingExperience("loading_experience", r.LoadingExperience, constLables, ch)
	collectLoadingExperience("origin_loading_experience", r.OriginLoadingExperience, constLables, ch)
	collectLighthouseResults("lighthouse", r.LighthouseResult, constLables, ch)
}

func collectLoadingExperience(prefix string, lexp *pagespeedonline.PagespeedApiLoadingExperienceV5, constLables prometheus.Labels, ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(fqname(prefix, "score"), "The specified score for the loading experience (1 FAST / 0.5 AVERAGE / 0 SLOW)  ", nil, constLables),
		prometheus.GaugeValue,
		convertCategoryToScore(lexp.OverallCategory))

	for k, v := range lexp.Metrics {
		name := strings.TrimSuffix(strings.ToLower(k), "_ms")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(fqname(prefix, name, "duration_seconds"), "Percentile metrics for "+strings.Replace(name, "_", " ", -1), nil, constLables),
			prometheus.GaugeValue,
			float64(v.Percentile)*1000)
	}

}

func collectLighthouseResults(prefix string, lhr *pagespeedonline.LighthouseResultV5, constLables prometheus.Labels, ch chan<- prometheus.Metric) {

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(fqname(prefix, "total_duration_seconds"), "The total time spent in seconds loading the page and evaluating audits.", nil, constLables),
		prometheus.GaugeValue,
		lhr.Timing.Total*1000) //ms -> seconds

	for k, v := range lhr.Categories {
		score, err := strconv.ParseFloat(fmt.Sprint(v.Score), 64)
		if err != nil {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(fqname(prefix, k, "score"), "The total/overall score for category: "+strings.Replace(k, "-", " ", -1), nil, constLables),
			prometheus.GaugeValue,
			score)
	}

	for k, v := range lhr.Audits {
		score, err := strconv.ParseFloat(fmt.Sprint(v.Score), 64)
		if err != nil {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(fqname(prefix, k), "Lighthouse audit score for "+strings.Replace(k, "-", " ", -1), nil, constLables),
			prometheus.GaugeValue,
			score)
	}
}

func convertCategoryToScore(category string) float64 {
	switch category {
	case "AVERAGE":
		return 0.5
	case "FAST":
		return 1
	case "NONE":
		return -1
	case "SLOW":
		return 0
	default:
		return -1
	}
}

func fqname(values ...string) string {
	return "pagespeed_" + strings.Replace(strings.Join(values, "_"), "-", "_", -1)
}
