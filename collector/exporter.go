package collector

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
	"github.com/foomo/pagespeed_exporter/googleapi"
	"github.com/foomo/pagespeed_exporter/exporter"
	"reflect"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ResultListener func(*googleapi.Result) error

type PagespeedCollector struct {
	listenerAddress string
	targets         []string
	googleApiKey    string
	checkInterval   time.Duration
	resultChannel   chan *googleapi.Result
	resultListeners []ResultListener
}

func NewCollector(listenerAddress, googleApiKey string, targets []string, checkInterval time.Duration) *PagespeedCollector {
	return &PagespeedCollector{
		listenerAddress: listenerAddress,
		googleApiKey:    googleApiKey,
		targets:         targets,
		checkInterval:   checkInterval,
		resultChannel:   make(chan *googleapi.Result, 1),
		resultListeners: []ResultListener{},
	}
}

func (e *PagespeedCollector) Start() error {
	startupMessage := "starting pagespeed exporter on listener %s for %d target(s) with re-check interval of %s"
	log.Infof(startupMessage, e.listenerAddress, len(e.targets), e.checkInterval)

	s := &http.Server{
		Addr: e.listenerAddress,
	}

	// Register prometheus handler
	http.Handle("/metrics", promhttp.Handler())

	// Register prometheus metrics resultListeners
	e.registerListener(exporter.PrometheusMetricsListener)

	go e.watch()
	go e.collect()

	return s.ListenAndServe()
}

func (e *PagespeedCollector) registerListener(listener ResultListener) {
	e.resultListeners = append(e.resultListeners, listener)
}

func (e *PagespeedCollector) watch() {
	service := googleapi.NewGoogleAPIService(e.googleApiKey)
	strategies := []googleapi.Strategy{googleapi.StrategyDesktop, googleapi.StrategyMobile}
	for true {
		for _, target := range e.targets {
			for _, strategy := range strategies {
				res, err := service.GetPagespeedResults(target, strategy)
				resultLogger := log.WithFields(log.Fields{
					"target":   target,
					"strategy": strategy,
				})

				if err != nil {
					resultLogger.Error("error occurred in pagespeed service", err)
				} else {
					resultLogger.Infof("successfully retrieved results for target %s and strategy %s", target, strategy)
					e.resultChannel <- res
				}
			}
		}
		time.Sleep(e.checkInterval)
	}
}

func (e *PagespeedCollector) collect() {
	for true {
		select {
		case res := <-e.resultChannel:
			e.handleResult(res)
		}
	}
}

func (e *PagespeedCollector) handleResult(result *googleapi.Result) {
	for _, l := range e.resultListeners {
		err := l(result)
		if err != nil {
			log.Error("listener " + reflect.TypeOf(l).String() + " thew an error while processing a result for target " + result.Target)
		}
	}
}
