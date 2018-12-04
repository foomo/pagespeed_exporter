package collector

import (
	"context"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi/transport"
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

type Strategy string

type scrapeRequest struct {
	target   string
	strategy Strategy
}

var _ scrapeService = &pagespeedScrapeService{}

type scrapeService interface {
	Scrape(targets []string) (scrapes []*Scrape, err error)
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

func (s *pagespeedScrapeService) Scrape(targets []string) (scrapes []*Scrape, err error) {
	strategies := []Strategy{StrategyDesktop, StrategyMobile}
	var scrapeRequests []scrapeRequest
	for _, t := range targets {
		for _, s := range strategies {
			scrapeRequests = append(scrapeRequests, scrapeRequest{
				target:   t,
				strategy: s,
			})
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(scrapeRequests))

	results := make(chan *Scrape)
	defer close(results)

	for _, sr := range scrapeRequests {
		go func(req scrapeRequest, res chan *Scrape) {
			defer wg.Done()
			scrape, errScrape := s.scrape(req)
			if errScrape != nil {
				logrus.WithError(errScrape).
					WithFields(logrus.Fields{
						"target":   req.target,
						"strategy": req.strategy,
					}).Warn("target scraping returned an error")
			} else {
				res <- scrape
			}
		}(sr, results)
	}
	go func() {
		for results != nil {
			select {
			case elem, ok := <-results:
				if ok {
					scrapes = append(scrapes, elem)
				}
			}
		}
	}()
	wg.Wait()

	return
}

func (s pagespeedScrapeService) scrape(request scrapeRequest) (scrape *Scrape, err error) {

	service, errClient := pagespeedonline.New(s.scrapeClient)
	if err != nil {
		return nil, errClient
	}
	call := service.Pagespeedapi.Runpagespeed(request.target)
	call.Category("performance", "seo", "pwa")
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
