package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/pagespeedonline/v5"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type collector struct {
	targets       []string
	scrapeService scrapeService
}

var timeAuditMetrics = map[string]bool{
	"first-contentful-paint": true,
	"first-cpu-idle":         true,
	"first-meaningful-paint": true,
	"interactive":            true,
	"speed-index":            true,
	"bootup-time":            true,
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
		ch <- prometheus.NewInvalidMetric(prometheus.NewDesc(fqname("error"), "Error scraping target", nil, nil), errScrape)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(fqname("scrape_duration_seconds"), "Total Pagespeed time scrape took for all targets.", nil, nil),
		prometheus.GaugeValue,
		float64(time.Since(start).Seconds()))

	for _, scrape := range result {
		errCollect := collect(scrape, ch)
		if errCollect != nil {
			logrus.WithError(errCollect).WithFields(logrus.Fields{
				"target":   scrape.Target,
				"strategy": scrape.Strategy,
			}).Error("could not collect scrape target due to errors")
		}
	}
}

func collect(scrape *Scrape, ch chan<- prometheus.Metric) error {
	constLabels, errLabels := getConstLabels(scrape)
	if errLabels != nil {
		return errLabels
	}

	r := scrape.Result
	collectLoadingExperience("loading_experience", r.LoadingExperience, constLabels, ch)
	collectLoadingExperience("origin_loading_experience", r.OriginLoadingExperience, constLabels, ch)
	collectLighthouseResults("lighthouse", r.LighthouseResult, constLabels, ch)
	return nil
}

func getConstLabels(scrape *Scrape) (prometheus.Labels, error) {
	target, errParse := url.Parse(scrape.Target)
	if errParse != nil {
		return nil, errParse
	}

	return prometheus.Labels{
		"host":     fmt.Sprintf("%s://%s", target.Scheme, target.Host),
		"path":     target.RequestURI(),
		"strategy": string(scrape.Strategy),
	}, nil
}

func collectLoadingExperience(prefix string, lexp *pagespeedonline.PagespeedApiLoadingExperienceV5, constLables prometheus.Labels, ch chan<- prometheus.Metric) {
	if lexp == nil {
		return
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(fqname(prefix, "score"), "The specified score for the loading experience (1 FAST / 0.5 AVERAGE / 0 SLOW)  ", nil, constLables),
		prometheus.GaugeValue,
		convertCategoryToScore(lexp.OverallCategory))

	for k, v := range lexp.Metrics {
		name := strings.TrimSuffix(strings.ToLower(k), "_ms")
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(fqname(prefix, "metrics", name, "duration_seconds"), "Percentile metrics for "+strings.Replace(name, "_", " ", -1), nil, constLables),
			prometheus.GaugeValue,
			float64(v.Percentile)/1000)
	}

}

func collectLighthouseResults(prefix string, lhr *pagespeedonline.LighthouseResultV5, constLabels prometheus.Labels, ch chan<- prometheus.Metric) {

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(fqname(prefix, "total_duration_seconds"), "The total time spent in seconds loading the page and evaluating audits.", nil, constLabels),
		prometheus.GaugeValue,
		lhr.Timing.Total/1000) //ms -> seconds

	for k, v := range lhr.Categories {
		score, err := strconv.ParseFloat(fmt.Sprint(v.Score), 64)
		if err != nil {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(fqname(prefix, "category_score"), "Lighthouse score for the specified category", []string{"category"}, constLabels),
			prometheus.GaugeValue,
			score,
			k)
	}

	for k, v := range lhr.Audits {
		if timeAuditMetrics[k] {
			v.DisplayValue = strings.Replace(v.DisplayValue, "\u00a0", "", -1)
			if duration, errDuration := time.ParseDuration(string(v.DisplayValue)); errDuration == nil {
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(fqname(prefix, k, "duration_seconds"), v.Description, nil, constLabels),
					prometheus.GaugeValue,
					duration.Seconds())
			} else {
				logrus.WithError(errDuration).Warn("could not parse time audit metric duration for metric: ", k)
			}
		}

		score, err := strconv.ParseFloat(fmt.Sprint(v.Score), 64)
		if err != nil {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(fqname(prefix, "audit_score"), "Lighthouse audit scores", []string{"audit"}, constLabels),
			prometheus.GaugeValue,
			score,
			k)
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
