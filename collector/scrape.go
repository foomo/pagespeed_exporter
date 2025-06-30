package collector

import (
	"context"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/foomo/pagespeed_exporter/cache"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/pagespeedonline/v5"
	googlehttp "google.golang.org/api/transport/http"
)

var _ scrapeService = &pagespeedScrapeService{}

type scrapeService interface {
	Scrape(parallel bool, config []ScrapeRequest) (scrapes []*ScrapeResult, err error)
}

// newPagespeedScrapeService creates a new HTTP client service for pagespeed.
// If the client timeout is set to 0 there will be no timeout
func newPagespeedScrapeService(clientTimeout time.Duration, cache cache.ResultCache, options ...option.ClientOption) (scrapeService, error) {
	transport, err := googlehttp.NewTransport(context.Background(), http.DefaultTransport, options...)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: transport,
	}

	if clientTimeout != 0 {
		client.Timeout = clientTimeout
	}

	return &pagespeedScrapeService{
		scrapeClient: client,
		options:      options,
		cache:        cache,
	}, nil
}

type pagespeedScrapeService struct {
	scrapeClient *http.Client
	options      []option.ClientOption
	cache        cache.ResultCache
}

func (pss *pagespeedScrapeService) Scrape(parallel bool, requests []ScrapeRequest) (scrapes []*ScrapeResult, err error) {

	maxWorkers := 1
	if parallel {
		maxWorkers = runtime.NumCPU()
	}

	results := make(chan *ScrapeResult, 2*len(requests))

	// Fill queue with scrape requests
	requestChan := make(chan ScrapeRequest)
	go func() {
		for _, r := range requests {
			requestChan <- r
		}
		close(requestChan)
	}()

	wg := sync.WaitGroup{}
	wg.Add(maxWorkers)

	for i := 0; i < maxWorkers; i++ {
		go func() {
			defer wg.Done()
			for request := range requestChan {
				scrape, err := pss.scrape(request)
				if err != nil {
					log.WithError(err).
						WithFields(log.Fields{
							"target":   request.Url,
							"strategy": request.Strategy,
						}).Warn("target scraping returned an error")
					continue
				}
				results <- scrape
			}
		}()
	}

	wg.Wait()
	close(results)

	// Drain the channel after receiving all the results
	for scrape := range results {
		scrapes = append(scrapes, scrape)
	}

	return
}

func (pss pagespeedScrapeService) scrape(request ScrapeRequest) (scrape *ScrapeResult, err error) {
	// Generate cache key
	cacheKey := cache.NewCacheKey(
		request.Url,
		string(request.Strategy),
		request.Categories,
		request.Locale,
		request.Campaign,
		request.Source,
	)

	// Check cache first
	if pss.cache != nil {
		if cachedResult, found := pss.cache.Get(cacheKey); found {
			log.WithFields(log.Fields{
				"url":      request.Url,
				"strategy": request.Strategy,
			}).Debug("cache hit")
			return &ScrapeResult{
				Request: request,
				Result:  cachedResult,
			}, nil
		}
	}

	// Cache miss, proceed with API call
	opts := []option.ClientOption{
		option.WithHTTPClient(pss.scrapeClient),
	}
	opts = append(opts, pss.options...)
	service, err := pagespeedonline.NewService(
		context.Background(),
		opts...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize pagespeed service")
	}

	call := service.Pagespeedapi.Runpagespeed(request.Url)
	call.Category(request.Categories...)
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

	// Store result in cache
	if pss.cache != nil && result != nil {
		pss.cache.Set(cacheKey, result)
		log.WithFields(log.Fields{
			"url":      request.Url,
			"strategy": request.Strategy,
		}).Debug("cached result")
	}

	return &ScrapeResult{
		Request: request,
		Result:  result,
	}, nil
}
