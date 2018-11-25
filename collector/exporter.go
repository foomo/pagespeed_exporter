package collector

import (
	"github.com/foomo/pagespeed_exporter/googleapi"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricsLabels = []string{"target", "strategy"}

	pageSpeedSpeedScore = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_rulegroup_speed_score",
		Help: "Pagespeed spe2ed rulegroup rating for the target",
	}, metricsLabels)

	pageSpeedUsabilityScore = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_rulegroup_usability_score",
		Help: "Pagespeed usability rulegroup rating for the target",
	}, metricsLabels)

	pageSpeedCssResponseBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_pagestats_css_response_bytes",
		Help: "Number of uncompressed response bytes for CSS resources on the page",
	}, metricsLabels)

	pageSpeedFlashResponseBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_pagestats_flash_response_bytes",
		Help: "Number of response bytes for flash resources on the page",
	}, metricsLabels)

	pageSpeedHtmlResponseBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_pagestats_html_response_bytes",
		Help: "Number of uncompressed response bytes for the main HTML document and all iframes on the page",
	}, metricsLabels)

	pageSpeedImageResponseBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_pagestats_image_response_bytes",
		Help: "Number of response bytes for image resources on on the page",
	}, metricsLabels)

	pageSpeedJavascriptResponseBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_pagestats_javascript_response_bytes",
		Help: "Number of uncompressed response bytes for JS resources on on the page",
	}, metricsLabels)

	pageSpeedOtherResponseBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_pagestats_other_response_bytes",
		Help: "Number of response bytes for other resources on the page",
	}, metricsLabels)

	pageSpeedTextResponseBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_pagestats_text_response_bytes",
		Help: "Number of uncompressed response bytes for text resources not covered by other statistics (i.e non-HTML, non-script, non-CSS resources) on the page",
	}, metricsLabels)

	pageSpeedTotalResponseBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_pagestats_total_response_bytes",
		Help: "Total size of all request bytes sent by the page",
	}, metricsLabels)

	pageSpeedCssResourceCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_pagestats_css_resources",
		Help: "Number of CSS resources referenced by the page",
	}, metricsLabels)

	pageSpeedHostCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_pagestats_hosts",
		Help: "Number of unique hosts referenced by the page",
	}, metricsLabels)

	pageSpeedJsResourceCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_pagestats_js_resources",
		Help: "Number of JavaScript resources referenced by the page",
	}, metricsLabels)

	pageSpeedHttpResourceCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_pagestats_http_resources",
		Help: "Number of HTTP resources loaded by the page",
	}, metricsLabels)

	pageSpeedStaticResourceCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pagespeed_pagestats_static_resources",
		Help: " Number of static (i.e. cacheable) resources on the page",
	}, metricsLabels)
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(pageSpeedSpeedScore)
	prometheus.MustRegister(pageSpeedUsabilityScore)
	prometheus.MustRegister(pageSpeedCssResponseBytes)
	prometheus.MustRegister(pageSpeedFlashResponseBytes)
	prometheus.MustRegister(pageSpeedHtmlResponseBytes)
	prometheus.MustRegister(pageSpeedImageResponseBytes)
	prometheus.MustRegister(pageSpeedJavascriptResponseBytes)
	prometheus.MustRegister(pageSpeedOtherResponseBytes)
	prometheus.MustRegister(pageSpeedTextResponseBytes)
	prometheus.MustRegister(pageSpeedTotalResponseBytes)
	prometheus.MustRegister(pageSpeedCssResourceCount)
	prometheus.MustRegister(pageSpeedHostCount)
	prometheus.MustRegister(pageSpeedJsResourceCount)
	prometheus.MustRegister(pageSpeedHttpResourceCount)
	prometheus.MustRegister(pageSpeedStaticResourceCount)
}

func PrometheusMetricsListener(result *googleapi.Result) error {

	if _, ok := result.RuleGroups["SPEED"]; ok {
		pageSpeedSpeedScore.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.RuleGroups["SPEED"].Score))
	}
	if _, ok := result.RuleGroups["USABILITY"]; ok {
		pageSpeedUsabilityScore.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.RuleGroups["USABILITY"].Score))
	}
	pageSpeedCssResponseBytes.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.PageStats.CssResponseBytes))
	pageSpeedFlashResponseBytes.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.PageStats.FlashResponseBytes))
	pageSpeedHtmlResponseBytes.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.PageStats.HtmlResponseBytes))
	pageSpeedImageResponseBytes.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.PageStats.ImageResponseBytes))
	pageSpeedJavascriptResponseBytes.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.PageStats.JavascriptResponseBytes))
	pageSpeedOtherResponseBytes.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.PageStats.OtherResponseBytes))
	pageSpeedTextResponseBytes.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.PageStats.TextResponseBytes))
	pageSpeedTotalResponseBytes.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.PageStats.TotalRequestBytes))
	pageSpeedCssResourceCount.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.PageStats.NumberCssResources))
	pageSpeedHostCount.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.PageStats.NumberHosts))
	pageSpeedJsResourceCount.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.PageStats.NumberJsResources))
	pageSpeedHttpResourceCount.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.PageStats.NumberResources))
	pageSpeedStaticResourceCount.WithLabelValues(result.Target, string(result.Strategy)).Set(float64(result.PageStats.NumberStaticResources))

	return nil
}
