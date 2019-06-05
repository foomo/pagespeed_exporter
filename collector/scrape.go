package collector

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/option"
	"google.golang.org/api/pagespeedonline/v5"
	"net/http"
	"sync"
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
}

var _ scrapeService = &pagespeedScrapeService{}

type scrapeService interface {
	Scrape(config ScrapeConfig) (scrapes []*Scrape, err error)
}

func newPagespeedScrapeService(clientTimeout time.Duration, googleApiKey string) scrapeService {
	client := &http.Client{
		Timeout:   clientTimeout,
		Transport: http.DefaultTransport,
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

func (s *pagespeedScrapeService) Scrape(config ScrapeConfig) (scrapes []*Scrape, err error) {
	strategies := []Strategy{StrategyDesktop, StrategyMobile}
	var scrapeRequests []scrapeRequest
	for _, t := range config.targets {
		for _, s := range strategies {
			scrapeRequests = append(scrapeRequests, scrapeRequest{
				target:   t,
				strategy: s,
			})
		}
	}

	if config.parallel {
		wg := sync.WaitGroup{}
		wg.Add(len(scrapeRequests))

		results := make(chan *Scrape, len(scrapeRequests))

		for _, sr := range scrapeRequests {
			go func(req scrapeRequest, res chan *Scrape) {
				defer wg.Done()
				scrape, err := s.scrape(req)
				if err != nil {
					logrus.WithError(err).
						WithFields(logrus.Fields{
							"target":   req.target,
							"strategy": req.strategy,
						}).Warn("target scraping returned an error")
					return
				}

				res <- scrape

			}(sr, results)
		}
		wg.Wait()
		close(results)

		for s := range results {
			scrapes = append(scrapes, s)
		}
	} else {
		for _, req := range scrapeRequests {
			scrape, err := s.scrape(req)
			if err != nil {
				logrus.WithError(err).
					WithFields(logrus.Fields{
						"target":   req.target,
						"strategy": req.strategy,
					}).Warn("target scraping returned an error")
				continue
			}
			scrapes = append(scrapes, scrape)
		}
	}
	return
}

func (s pagespeedScrapeService) scrape(request scrapeRequest) (scrape *Scrape, err error) {

	service, err := pagespeedonline.NewService(context.Background(), option.WithHTTPClient(s.scrapeClient))
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize pagespeed service")
	}
	call := service.Pagespeedapi.Runpagespeed(request.target)
	call.Category("performance", "seo", "pwa", "best-practices", "accessibility")
	call.Strategy(string(request.strategy))

	call.Context(context.WithValue(context.Background(), oauth2.HTTPClient, s.scrapeClient))

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
