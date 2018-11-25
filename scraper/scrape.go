package scraper

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/api/pagespeedonline/v4"
	"net"
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
	Result   *pagespeedonline.PagespeedApiPagespeedResponseV4
}

type Strategy string

type scrapeRequest struct {
	target   string
	strategy Strategy
}

var _ Service = &pagespeedService{}

type Service interface {
	Scrape(targets []string) (scrapes []*Scrape, err error)
}

func New() Service {
	return &pagespeedService{
		scrapeClient: getScrapeClient(),
	}
}

type pagespeedService struct {
	scrapeClient *http.Client
}

func (s *pagespeedService) Scrape(targets []string) (scrapes []*Scrape, err error) {
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
	for _, sr := range scrapeRequests {
		go func(req scrapeRequest) {
			scrape, errScrape := s.scrape(req)
			if errScrape != nil {
				logrus.WithError(errScrape).
					WithFields(logrus.Fields{
						"target":   req.target,
						"strategy": req.strategy,
					}).Warn("target scraping returned an error")
			} else {
				scrapes = append(scrapes, scrape)
			}
			wg.Done()
		}(sr)
	}
	wg.Wait()
	return
}

func (s pagespeedService) scrape(request scrapeRequest) (scrape *Scrape, err error) {
	service, errClient := pagespeedonline.New(s.scrapeClient)
	if err != nil {
		return nil, errClient
	}
	call := service.Pagespeedapi.Runpagespeed(request.target)
	call.Strategy(string(request.strategy))
	call.Snapshots(false)
	call.Screenshot(false)
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

func getScrapeClient() *http.Client {
	var netTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	return &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

}
