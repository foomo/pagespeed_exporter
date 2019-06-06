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

const (
	StrategyMobile  = Strategy("mobile")
	StrategyDesktop = Strategy("desktop")
)

type Scrape struct {
	Target   string
	Strategy Strategy
	Result   *pagespeedonline.PagespeedApiPagespeedResponseV5
}

type ScrapeConfig struct {
	targets  []string
	parallel bool
}

type Strategy string

type scrapeRequest struct {
	target   string
	strategy Strategy
	campaign string
	source   string
	locale   string
}

var _ scrapeService = &pagespeedScrapeService{}

type scrapeService interface {
	Scrape(config ScrapeConfig) (scrapes []*Scrape, err error)
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

func (pss *pagespeedScrapeService) Scrape(config ScrapeConfig) (scrapes []*Scrape, err error) {
	strategies := []Strategy{StrategyDesktop, StrategyMobile}

	maxWorkers := 1
	if config.parallel {
		maxWorkers = runtime.NumCPU()
	}

	wp := workerpool.New(maxWorkers)

	results := make(chan *Scrape, 2*len(config.targets))

	for _, t := range config.targets {
		for _, s := range strategies {
			target := t
			strategy := s

			wp.Submit(func() {
				req := scrapeRequest{
					target:   target,
					strategy: strategy,
				}
				scrape, err := pss.scrape(req)
				if err != nil {
					logrus.WithError(err).
						WithFields(logrus.Fields{
							"target":   req.target,
							"strategy": req.strategy,
						}).Warn("target scraping returned an error")
					return
				}
				results <- scrape
			})

		}
	}

	wp.StopWait()
	close(results)

	// Drain the channel after receiving all the results
	for scrape := range results {
		scrapes = append(scrapes, scrape)
	}

	return
}

func (pss pagespeedScrapeService) scrape(request scrapeRequest) (scrape *Scrape, err error) {

	service, err := pagespeedonline.NewService(context.Background(), option.WithHTTPClient(pss.scrapeClient))
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize pagespeed service")
	}

	call := service.Pagespeedapi.Runpagespeed(request.target)
	call.Category("performance", "seo", "pwa", "best-practices", "accessibility")
	call.Strategy(string(request.strategy))

	call.Context(context.WithValue(context.Background(), oauth2.HTTPClient, pss.scrapeClient))

	result, errResult := call.Do()

	if errResult != nil {
		return nil, errResult
	}

	return &Scrape{
		Target:   request.target,
		Strategy: request.strategy,
		Result:   result,
	}, nil
}
