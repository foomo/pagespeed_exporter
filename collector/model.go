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

	Namespace = "pagespeed"
)

type Strategy string

var availableStrategies = map[Strategy]bool{
	StrategyMobile:  true,
	StrategyDesktop: true,
}

type ScrapeResult struct {
	Request ScrapeRequest
	Result  *pagespeedonline.PagespeedApiPagespeedResponseV5
}

type ScrapeRequest struct {
	Url      string   `json:"url"`
	Strategy Strategy `json:"strategy"`
	Campaign string   `json:"campaign"`
	Source   string   `json:"source"`
	Locale   string   `json:"locale"`
}

func (sr ScrapeRequest) IsValid() bool {
	if sr.Url == "" {
		return false
	}

	if !availableStrategies[sr.Strategy] {
		return false
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

func CalculateScrapeRequests(targets ...string) []ScrapeRequest {
	if len(targets) == 0 {
		return nil
	}
	var requests []ScrapeRequest

	for _, t := range targets {
		var request ScrapeRequest
		if err := json.Unmarshal([]byte(t), &request); err == nil {
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
			requests = append(requests,
				ScrapeRequest{Url: t, Strategy: StrategyDesktop},
				ScrapeRequest{Url: t, Strategy: StrategyMobile},
			)
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
