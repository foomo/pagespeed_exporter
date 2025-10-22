package collector

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"google.golang.org/api/pagespeedonline/v5"
)

var (
	_ Factory = factory{}
)

var (
	timeValueRe = regexp.MustCompile(`(\d*[.]?\d+(ms|s))|0`)
	timeUnitRe  = regexp.MustCompile(`(ms|s)`)
)

type Factory interface {
	Create(config Config) (prometheus.Collector, error)
}

func NewFactory() Factory {
	return factory{}
}

type factory struct {
}

type collector struct {
	requests      []ScrapeRequest
	scrapeService scrapeService
	parallel      bool
}

func (factory) Create(config Config) (prometheus.Collector, error) {
	return newCollector(config)
}

var timeAuditMetrics = map[string]bool{
	"first-contentful-paint":    true,
	"first-cpu-idle":            true,
	"first-meaningful-paint":    false,
	"interactive":               true,
	"speed-index":               true,
	"bootup-time":               true,
	"largest-contentful-paint":  true,
	"mainthread-work-breakdown": true,
	"cumulative-layout-shift":   true,
	"total-blocking-time":       true,
	"server-response-time":      true,
	"max-potential-fid":         true,
	"estimated-input-latency":   true,
}

func newCollector(config Config) (coll prometheus.Collector, err error) {
	var options []option.ClientOption
	if config.GoogleAPIKey != "" {
		options = append(options, option.WithAPIKey(config.GoogleAPIKey))
	}

	if config.CredentialsFile != "" {
		options = append(options, option.WithCredentialsFile(config.CredentialsFile))
	}

	svc, err := newPagespeedScrapeService(config.ScrapeTimeout, config.CacheTTL, options...)
	if err != nil {
		return nil, err
	}

	return collector{
		requests:      config.ScrapeRequests,
		scrapeService: svc,
		parallel:      config.Parallel,
	}, nil
}

// Describe implements Prometheus.Collector.
func (c collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

// Collect implements Prometheus.Collector.
func (c collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	result, errScrape := c.scrapeService.Scrape(c.parallel, c.requests)
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
				"target":   scrape.Request.Url,
				"strategy": scrape.Request.Strategy,
			}).Error("could not collect scrape target due to errors")
		}
	}
}

func collect(scrape *ScrapeResult, ch chan<- prometheus.Metric) error {
	constLabels, errLabels := getConstLabels(scrape)
	if errLabels != nil {
		return errLabels
	}

	r := scrape.Result
	if r.LoadingExperience != nil {
		collectLoadingExperience("loading_experience", r.LoadingExperience, constLabels, ch)
	}

	if r.OriginLoadingExperience != nil {
		collectLoadingExperience("origin_loading_experience", r.OriginLoadingExperience, constLabels, ch)
	}

	if r.LighthouseResult != nil {
		collectLighthouseResults("lighthouse", scrape.Request.Categories, r.LighthouseResult, constLabels, ch)
	}
	return nil
}

func getConstLabels(scrape *ScrapeResult) (prometheus.Labels, error) {
	target, errParse := url.Parse(scrape.Request.Url)
	if errParse != nil {
		return nil, errParse
	}

	return prometheus.Labels{
		"host":     fmt.Sprintf("%s://%s", target.Scheme, target.Host),
		"path":     target.RequestURI(),
		"strategy": string(scrape.Request.Strategy),
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

func collectLighthouseResults(prefix string, cats []string, lhr *pagespeedonline.LighthouseResultV5, constLabels prometheus.Labels, ch chan<- prometheus.Metric) {

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(fqname(prefix, "total_duration_seconds"), "The total time spent in seconds loading the page and evaluating audits.", nil, constLabels),
		prometheus.GaugeValue,
		lhr.Timing.Total/1000) // ms -> seconds

	categories := map[string]*pagespeedonline.LighthouseCategoryV5{
		CategoryPerformance:   lhr.Categories.Performance,
		CategoryAccessibility: lhr.Categories.Accessibility,
		CategoryBestPractices: lhr.Categories.BestPractices,
		CategorySEO:           lhr.Categories.Seo,
	}

	for _, c := range cats {
		if categories[c] != nil {
			score, err := strconv.ParseFloat(fmt.Sprint(categories[c].Score), 64)
			if err != nil {
				logrus.WithError(err).Warn("could not parse category score")
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(fqname(prefix, "category_score"), "Lighthouse score for the specified category", []string{"category"}, constLabels),
				prometheus.GaugeValue,
				score,
				c)
		}
	}

	for k, v := range lhr.Audits {
		if timeAuditMetrics[k] {
			displayValue := strings.Replace(v.DisplayValue, "\u00a0", "", -1)
			displayValue = strings.Replace(displayValue, ",", "", -1)
			if !timeUnitRe.MatchString(displayValue) {
				displayValue = displayValue + "s"
			}

			if duration, errDuration := time.ParseDuration(timeValueRe.FindString(displayValue)); errDuration == nil {
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
	return fmt.Sprintf("%s_%s", Namespace, strings.Replace(strings.Join(values, "_"), "-", "_", -1))
}
