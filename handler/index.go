package handler

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

var (
	indexPage = []byte(`
		<html>
			<head><title>Pagespeed Exporter</title></head>
		<body>
			<h1>Pagespeed Exporter</h1>
			<p><a href="/probe?target=https://www.google.com">Probe https://www.google.com</a></p>
			<p><a href="/metrics">Metrics</a></p>
		</body>
		</html>

	`)
)

func NewIndexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(indexPage); err != nil {
			log.WithError(err).Warn("could not write to stream")
		}
	})
}
