package collector

import (
	"testing"
	"time"
)

func Test_PagespeedScrapeService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	service := newPagespeedScrapeService(30*time.Second, "")
	scrapes, err := service.Scrape(ScrapeConfig{
		parallel: false,
		targets:  []string{"http://example.com/"},
	})
	if err != nil {
		t.Fatal("scrape should not throw an error")
	}
	if len(scrapes) != 2 {
		t.Fatal("scrape should return 2 results for strategies")
	}
}
