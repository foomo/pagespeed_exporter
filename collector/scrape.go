package collector

import (
	"context"
	"github.com/gammazero/workerpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/option"
	"google.golang.org/api/pagespeedonline/v5"
	"net/http"
	"runtime"
	"time"
)

var _ scrapeService = &pagespeedScrapeService{}

type scrapeService interface {
	Scrape(parallel bool, config []ScrapeRequest) (scrapes []*ScrapeResult, err error)
}

// newPagespeedScrapeService creates a new HTTP client service for pagespeed.
// If the client timeout is set to 0 there will be no timeout
func newPagespeedScrapeService(clientTimeout time.Duration, googleApiKey string) scrapeService {

	client := &http.Client{
		Transport: http.DefaultTransport,
	}

	if clientTimeout != 0 {
		client.Timeout = clientTimeout
	}

	if googleApiKey != "" {
		client.Transport = &transport.APIKey{Key: googleApiKey}
	}

	return &pagespeedScrapeService{
		scrapeClient: client,
	}
}

type pagespeedScrapeService struct {
	scrapeClient *http.Client
}

func (pss *pagespeedScrapeService) Scrape(parallel bool, requests []ScrapeRequest) (scrapes []*ScrapeResult, err error) {

	maxWorkers := 1
	if parallel {
		maxWorkers = runtime.NumCPU()
	}

	wp := workerpool.New(maxWorkers)

	results := make(chan *ScrapeResult, 2*len(requests))

	for _, req := range requests {
		request := req
		wp.Submit(func() {

			scrape, err := pss.scrape(req)
			if err != nil {
				logrus.WithError(err).
					WithFields(logrus.Fields{
						"target":   request.Url,
						"strategy": request.Strategy,
					}).Warn("target scraping returned an error")
				return
			}
			results <- scrape
		})

	}

	wp.StopWait()
	close(results)

	// Drain the channel after receiving all the results
	for scrape := range results {
		scrapes = append(scrapes, scrape)
	}

	return
}

func (pss pagespeedScrapeService) scrape(request ScrapeRequest) (scrape *ScrapeResult, err error) {

	service, err := pagespeedonline.NewService(context.Background(), option.WithHTTPClient(pss.scrapeClient))
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize pagespeed service")
	}

	call := service.Pagespeedapi.Runpagespeed(request.Url)
	call.Category("performance", "seo", "pwa", "best-practices", "accessibility")
	call.Strategy(string(request.Strategy))

	if request.Campaign != "" {
		call.UtmCampaign(request.Campaign)
	}

	if request.Locale != "" {
		call.Locale(request.Locale)
	}

	if request.Source != "" {
		call.UtmSource(request.Source)
	}

	call.Context(context.WithValue(context.Background(), oauth2.HTTPClient, pss.scrapeClient))

	result, errResult := call.Do()
	if errResult != nil {
		return nil, errResult
	}

	return &ScrapeResult{
		Request: request,
		Result:  result,
	}, nil
}
