package collector

import (
	"testing"
	"time"
)

func Test_PagespeedScrapeService(t *testing.T) {
	service := newPagespeedScrapeService(30*time.Second, "")
	scrapes, err := service.Scrape([]string{"https://google.com/"})
	if err != nil {
		t.Fatal("scrape should not throw an error")
	}
	if len(scrapes) != 2 {
		t.Fatal("scrape should return 2 results for strategies")
	}
}
