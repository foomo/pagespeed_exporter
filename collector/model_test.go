package collector

import (
	"reflect"
	"sort"
	"testing"
)

func TestCalculateScrapeRequests(t *testing.T) {
	allCategories := []string{"accessibility", "best-practices", "performance", "seo"}

	tests := []struct {
		name       string
		targets    []string
		categories []string
		want       []ScrapeRequest
	}{
		{"empty", []string{}, nil, nil},
		{"invalid URL", []string{"url"}, nil, []ScrapeRequest{}},
		{"single basic", []string{"http://test.com"}, nil, []ScrapeRequest{
			{Url: "http://test.com", Strategy: StrategyDesktop, Categories: allCategories},
			{Url: "http://test.com", Strategy: StrategyMobile, Categories: allCategories},
		}},
		{"multiple basic", []string{"http://test.com", "http://test2.com"}, nil, []ScrapeRequest{
			{Url: "http://test.com", Strategy: StrategyDesktop, Categories: allCategories},
			{Url: "http://test.com", Strategy: StrategyMobile, Categories: allCategories},
			{Url: "http://test2.com", Strategy: StrategyDesktop, Categories: allCategories},
			{Url: "http://test2.com", Strategy: StrategyMobile, Categories: allCategories},
		}},
		{"single with categories", []string{"http://test.com"}, []string{"accessibility", "seo"}, []ScrapeRequest{
			{Url: "http://test.com", Strategy: StrategyDesktop, Categories: []string{"accessibility", "seo"}},
			{Url: "http://test.com", Strategy: StrategyMobile, Categories: []string{"accessibility", "seo"}},
		}},
		{"multiple with categories", []string{"http://test.com", "http://test2.com"}, []string{"best-practices"}, []ScrapeRequest{
			{Url: "http://test.com", Strategy: StrategyDesktop, Categories: []string{"best-practices"}},
			{Url: "http://test.com", Strategy: StrategyMobile, Categories: []string{"best-practices"}},
			{Url: "http://test2.com", Strategy: StrategyDesktop, Categories: []string{"best-practices"}},
			{Url: "http://test2.com", Strategy: StrategyMobile, Categories: []string{"best-practices"}},
		}},
		{"single with wrong categories", []string{"http://test.com"}, []string{"accessibility", "pancake"}, []ScrapeRequest{}},
		{"multiple with wrong categories", []string{"http://test.com", "http://test2.com"}, []string{"accessibility", "pancake"}, []ScrapeRequest{}},
		{"json",
			[]string{`{"url":"http://test.com","strategy":"desktop","campaign":"campaign","source":"source","locale":"locale"}`}, nil, []ScrapeRequest{
				{
					Url:        "http://test.com",
					Strategy:   StrategyDesktop,
					Campaign:   "campaign",
					Source:     "source",
					Locale:     "locale",
					Categories: allCategories,
				},
			}},
		{"json with category",
			[]string{`{"url":"http://test.com","strategy":"desktop","campaign":"campaign","source":"source","locale":"locale", "categories":["seo"]}`}, nil, []ScrapeRequest{
				{
					Url:        "http://test.com",
					Strategy:   StrategyDesktop,
					Campaign:   "campaign",
					Source:     "source",
					Locale:     "locale",
					Categories: []string{"seo"},
				},
			}},
		{"json with wrong category",
			[]string{`{"url":"http://test.com","strategy":"desktop","campaign":"campaign","source":"source","locale":"locale", "categories":["waffle"]}`}, nil, []ScrapeRequest{}},
		{"json simple",
			[]string{`{"url":"http://test.com"}`}, nil, []ScrapeRequest{
				{Url: "http://test.com", Strategy: StrategyDesktop, Categories: allCategories},
				{Url: "http://test.com", Strategy: StrategyMobile, Categories: allCategories},
			}},
		{"json missing URL",
			[]string{`{"strategy":"desktop"}`}, nil,
			[]ScrapeRequest{}},
		{"json  bad strategy",
			[]string{`{"url":"http://test.com","strategy":"microwave"}`}, nil,
			[]ScrapeRequest{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateScrapeRequests(tt.targets, tt.categories)

			// To be able to reflect.DeepEqual on the categories slice
			for _, r := range got {
				sort.Strings(r.Categories)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalculateScrapeRequests() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPopulateCategories(t *testing.T) {
	allCategories := []string{"accessibility", "best-practices", "performance", "seo"}

	tests := []struct {
		msg  string
		req  *ScrapeRequest
		cats []string
		want *ScrapeRequest
	}{
		{
			"request is not changed if categories exist",
			&ScrapeRequest{
				Categories: []string{"performance"},
			},
			[]string{"best-practices"},
			&ScrapeRequest{
				Categories: []string{"performance"},
			},
		},
		{
			"available categories set if request has no categories",
			&ScrapeRequest{},
			nil,
			&ScrapeRequest{
				Categories: allCategories,
			},
		},
		{
			"input categories are set in request",
			&ScrapeRequest{},
			[]string{"best-practices", "pancake", "seo"},
			&ScrapeRequest{
				Categories: []string{"best-practices", "pancake", "seo"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			populateCategories(tt.req, tt.cats)

			// To be able to reflect.DeepEqual on the categories slice
			sort.Strings(tt.req.Categories)

			if !reflect.DeepEqual(tt.req, tt.want) {
				t.Errorf("populateCategories() = %+v, want %+v", tt.req, tt.want)
			}
		})
	}
}
