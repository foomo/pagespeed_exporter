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

func NewCollector(listenerAddress, googleApiKey string, targets []string) *PagespeedCollector {
	return &PagespeedCollector{
		listenerAddress: listenerAddress,
		googleApiKey:    googleApiKey,
		targets:         targets,
		checkInterval:   30 * time.Minute,
		resultChannel:   make(chan *googleapi.Result, 1),
		resultListeners: []ResultListener{},
	}
}

func (e *PagespeedCollector) Start() error {
	log.Info("starting pagespeed exporter on listener ", e.listenerAddress, " for ", len(e.targets), " target(s)")

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
	for true {
		for _, target := range e.targets {
			res, err := service.GetPagespeedResults(target)
			if err != nil {
				log.WithField("target", target).Error("error occurred in pagespeed service", err)
			} else {
				e.resultChannel <- res
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
