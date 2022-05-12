package collector

import (
	"os"
	"testing"
	"time"

	"google.golang.org/api/option"
)

const (
	envvarAPIKey          = "PAGESPEED_API_KEY"
	envvarCredentialsFile = "PAGESPEED_CREDENTIALS_FILE"
)

func Test_PagespeedScrapeService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	var options []option.ClientOption

	if apiKey := os.Getenv(envvarAPIKey); apiKey != "" {
		options = append(options, option.WithAPIKey(apiKey))
	}

	if cf := os.Getenv(envvarCredentialsFile); cf != "" {
		options = append(options, option.WithCredentialsFile(cf))
	}

	if len(options) == 0 {
		t.Skip("skipping testing unless API key or credentials file is set")
	}

	service, err := newPagespeedScrapeService(30*time.Second, options...)
	if err != nil {
		t.Fatalf("newPagespeedScrapeService should not throw an error: %v", err)
	}

	scrapes, err := service.Scrape(true, CalculateScrapeRequests([]string{"http://example.com/"}, nil))
	if err != nil {
		t.Fatal("scrape should not throw an error")
	}
	if len(scrapes) != 2 {
		t.Fatal("scrape should return 2 results for strategies")
	}
}
