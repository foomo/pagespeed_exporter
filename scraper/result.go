package scraper

import "google.golang.org/api/pagespeedonline/v4"

type Category float64

const (
	CategoryUnknown = Category(-1)
	CategoryFast    = Category(0)
	CategoryAverage = Category(1)
	CategorySlow    = Category(2)
)

type Scrape struct {
	Target   string
	Strategy Strategy
	Result   *Result
}

type Strategy string

type scrapeRequest struct {
	target   string
	strategy Strategy
}

type Result struct {
	Score         float64
	ResponseBytes ResponseData
	Speed         Speed
}
type ResponseData struct {
	FlashBytes      float64
	CssBytes        float64
	HtmlBytes       float64
	ImageBytes      float64
	JavaScriptBytes float64
	OtherBytes      float64
	TextBytes       float64
	TotalBytes      float64
	TotalRoundTrips float64
}

type Speed struct {
	Overall           Category
	DomLoadedMetrics  SpeedMetrics
	FirstPaintMetrics SpeedMetrics
}

type SpeedMetrics struct {
	Category Category
	Median   float64 //Seconds
}

func newFromPagespeedResults(resp *pagespeedonline.PagespeedApiPagespeedResponseV4) (result *Result) {
	r := &Result{}

	r.ResponseBytes.FlashBytes = float64(resp.PageStats.FlashResponseBytes)
	r.ResponseBytes.CssBytes = float64(resp.PageStats.CssResponseBytes)
	r.ResponseBytes.HtmlBytes = float64(resp.PageStats.HtmlResponseBytes)
	r.ResponseBytes.ImageBytes = float64(resp.PageStats.ImageResponseBytes)
	r.ResponseBytes.JavaScriptBytes = float64(resp.PageStats.JavascriptResponseBytes)
	r.ResponseBytes.OtherBytes = float64(resp.PageStats.OtherResponseBytes)
	r.ResponseBytes.TextBytes = float64(resp.PageStats.TextResponseBytes)
	r.ResponseBytes.TotalBytes = float64(resp.PageStats.TotalRequestBytes)
	r.ResponseBytes.TotalRoundTrips = float64(resp.PageStats.NumTotalRoundTrips)

	if resp.LoadingExperience != nil {
		r.Speed.Overall = getCategoryFromString(resp.LoadingExperience.OverallCategory)
		for k, v := range resp.LoadingExperience.Metrics {
			if k == "DOM_CONTENT_LOADED_EVENT_FIRED_MS" {
				r.Speed.DomLoadedMetrics.Category = getCategoryFromString(v.Category)
				r.Speed.DomLoadedMetrics.Median = float64(v.Median) / 1000 // To Seconds
			}
			if k == "FIRST_CONTENTFUL_PAINT_MS" {
				r.Speed.FirstPaintMetrics.Category = getCategoryFromString(v.Category)
				r.Speed.FirstPaintMetrics.Median = float64(v.Median) / 1000 // To Seconds
			}
		}

	}

	return r
}

func getCategoryFromString(category string) Category {
	switch category {
	case "FAST":
		return CategoryFast
	case "AVERAGE":
		return CategoryAverage
	case "SLOW":
		return CategorySlow
	default:
		return CategoryUnknown
	}
}
