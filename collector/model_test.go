package collector

import (
	"reflect"
	"testing"
)

func TestCalculateScrapeRequests(t *testing.T) {
	tests := []struct {
		name    string
		targets []string
		want    []ScrapeRequest
	}{
		{"empty", []string{}, nil},
		{"invalid URL", []string{"url"}, []ScrapeRequest{}},
		{"single basic", []string{"http://test.com"}, []ScrapeRequest{
			{Url: "http://test.com", Strategy: StrategyDesktop},
			{Url: "http://test.com", Strategy: StrategyMobile},
		}},
		{"multiple basic", []string{"http://test.com", "http://test2.com"}, []ScrapeRequest{
			{Url: "http://test.com", Strategy: StrategyDesktop},
			{Url: "http://test.com", Strategy: StrategyMobile},
			{Url: "http://test2.com", Strategy: StrategyDesktop},
			{Url: "http://test2.com", Strategy: StrategyMobile},
		}},
		{"json",
			[]string{`{"url":"http://test.com","strategy":"desktop","campaign":"campaign","source":"source","locale":"locale"}`}, []ScrapeRequest{
				{
					Url:      "http://test.com",
					Strategy: StrategyDesktop,
					Campaign: "campaign",
					Source:   "source",
					Locale:   "locale",
				},
			}},
		{"json simple",
			[]string{`{"url":"http://test.com"}`}, []ScrapeRequest{
				{Url: "http://test.com", Strategy: StrategyDesktop},
				{Url: "http://test.com", Strategy: StrategyMobile},
			}},
		{"json missing URL",
			[]string{`{"strategy":"desktop"}`},
			[]ScrapeRequest{}},
		{"json  bad strategy",
			[]string{`{"url":"http://test.com","strategy":"microwave"}`},
			[]ScrapeRequest{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateScrapeRequests(tt.targets...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalculateScrapeRequests() = %v, want %v", got, tt.want)
			}
		})
	}
}
