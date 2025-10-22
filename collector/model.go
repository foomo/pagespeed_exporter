package collector

import (
	"encoding/json"
	"net/url"
	"time"

	"google.golang.org/api/pagespeedonline/v5"
)

const (
	StrategyMobile  = Strategy("mobile")
	StrategyDesktop = Strategy("desktop")

	CategoryAccessibility = "accessibility"
	CategoryBestPractices = "best-practices"
	CategorySEO           = "seo"
	CategoryPerformance   = "performance"

	Namespace = "pagespeed"
)

type Strategy string

var availableStrategies = map[Strategy]bool{
	StrategyMobile:  true,
	StrategyDesktop: true,
}

var availableCategories = map[string]bool{
	CategoryAccessibility: true,
	CategoryBestPractices: true,
	CategorySEO:           true,
	CategoryPerformance:   true,
}

type ScrapeResult struct {
	Request ScrapeRequest
	Result  *pagespeedonline.PagespeedApiPagespeedResponseV5
}

type ScrapeRequest struct {
	Url        string   `json:"url"`
	Strategy   Strategy `json:"strategy"`
	Campaign   string   `json:"campaign"`
	Source     string   `json:"source"`
	Locale     string   `json:"locale"`
	Categories []string `json:"categories"`
}

func (sr ScrapeRequest) IsValid() bool {
	if sr.Url == "" {
		return false
	}

	if !availableStrategies[sr.Strategy] {
		return false
	}

	for _, c := range sr.Categories {
		if !availableCategories[c] {
			return false
		}
	}

	if _, err := url.ParseRequestURI(sr.Url); err != nil {
		return false
	}

	return true
}

type Config struct {
	ScrapeRequests  []ScrapeRequest
	GoogleAPIKey    string
	CredentialsFile string
	Parallel        bool
	ScrapeTimeout   time.Duration
}

func CalculateScrapeRequests(targets, categories []string) []ScrapeRequest {
	if len(targets) == 0 {
		return nil
	}
	var requests []ScrapeRequest

	for _, t := range targets {
		var request ScrapeRequest
		if err := json.Unmarshal([]byte(t), &request); err == nil {
			populateCategories(&request, categories)
			if request.Strategy != "" {
				requests = append(requests, request)
			} else {
				desktop := ScrapeRequest(request)
				desktop.Strategy = StrategyDesktop
				mobile := ScrapeRequest(request)
				mobile.Strategy = StrategyMobile
				requests = append(requests, desktop, mobile)
			}
		} else {
			desktop := ScrapeRequest{Url: t, Strategy: StrategyDesktop}
			mobile := ScrapeRequest{Url: t, Strategy: StrategyMobile}
			populateCategories(&desktop, categories)
			populateCategories(&mobile, categories)
			requests = append(requests, desktop, mobile)
		}
	}

	filtered := requests[:0]

	for _, sr := range requests {
		if sr.IsValid() {
			filtered = append(filtered, sr)
		}
	}

	return filtered
}

// populateCategories sets categories in the scrape request if not already set
func populateCategories(r *ScrapeRequest, cats []string) {
	if r.Categories != nil && len(r.Categories) != 0 {
		return
	}

	if cats == nil {
		cats = make([]string, 0, len(availableCategories))
	}

	if len(cats) == 0 {
		for c := range availableCategories {
			cats = append(cats, c)
		}
	}

	r.Categories = cats
}
