# Prometheus Exporter for Google Pagespeed Online Metrics
[![Travis CI](https://travis-ci.org/foomo/pagespeed_exporter.svg?branch=master)](https://travis-ci.org/foomo/pagespeed_exporter)


## Examples

To start the example with the default dashboard (docker-compose required)

```
$ git clone git@github.com:foomo/pagespeed_exporter.git
$ cd pagespeed_exporter/example
$ docker-compose up -d
```

After that, the application should be running on ``localhost:3000`` with username admin and password s3cr3t.

The provided dashboard (Pagespeed) will be loaded with data after the first scrape.

![Dashboard](https://github.com/foomo/pagespeed_exporter/raw/assets/dashboard.png?raw=true)

The  dashboard can be found at [grafana](https://grafana.com/grafana/dashboards/9510-pagespeed/)

Note: The example dashboard assumes you're fetching all pagespeed categories.

## Understanding Metrics

* https://github.com/GoogleChrome/lighthouse/blob/master/docs/understanding-results.md
* https://developers.google.com/speed/docs/insights/v5/reference/pagespeedapi/runpagespeed

Prometheus exporter for google pagespeed metrics


## Building And Running

### Building

```sh
$ make
```

### Examples

To run pagespeed exporter we need to obtain the google api key for the pagespeed. 
Instructions how to create a key for pagespeed can be found [here](https://developers.google.com/speed/docs/insights/v2/first-app)

`pagespeed_exporter <arguments>`

```sh
$ pagespeed_exporter -api-key {KEY} -targets https://google.com,https://prometheus.io -listener :80
```

### Exporter Target Specification

Targets can be configured in either plaintext 

```
https://github.com/foomo/pagespeed_exporter
https://mysite.com/test?test=true
```

Or via JSON which adds additional parameters

```
// URL can't be invalid
// Strategy can only be mobile/desktop
// If strategy is not specified, both desktop & mobile will be used
// Categories can be any of accessibility/best-practices/performance/seo
// If categories are not specified, all categories will be used
// Parameters are passed down to google pagespeed api

{"url":"https://github.com/foomo/pagespeed_exporter","campaign":"test","locale":"en","source":"source"}

{"url":"https://mysite.com/test?test=true","strategy":"mobile"}

{"url":"https://mysite.com/test?test=true","categories": ["best-practices"]}

```

Configuration specification in JSON and plain is supported both in command line & prometheus configuration 

### Exporter configuration 

Configuration of targets can be done via docker and via prometheus

| Flag             | Variable             | Description                                                       | Default                                          | Required |
|------------------|----------------------|-----------------------------------------------|----------------------------------------------------------------------|----------|
| -api-key         | PAGESPEED_API_KEY    | sets the google API key used for pagespeed                        |                                                  | False    |
| -targets         | PAGESPEED_TARGETS    | comma separated list of targets to measure                        |                                                  | False    |
| -categories      | PAGESPEED_CATEGORIES | comma separated list of categories to check                       | accessibility,best-practices,performance,seo | False    |
| -t               | NONE                 | multi-value target array (check docker comp)                      |                                                  | False    |
| -listener        | PAGESPEED_LISTENER   | sets the listener address for the exporters                       | :9271                                            | False    |
| -parallel        | PAGESPEED_PARALLEL   | sets the execution of targets to be parallel                      | false                                            | False    |
| -pushGatewayUrl  | PUSHGATEWAY_URL      | sets the pushgateway url to send the metrics                      |                                                  | False    |
| -pushGatewayJob  | PUSHGATEWAY_JOB      | sets the pushgateway job name                                     | pagespeed_exporter                               | False    |
| -cache-ttl       | CACHE_TTL            | cache TTL for API results (e.g. 60s, 5m); disables cache if unset |                                                  | False    |

Note: google api key is required only if scraping more than 2 targets/second

Note: exporter can be run without targets, and later targets provided via prometheus


### Pushing metrics via push gateway

If you don't want to change the prometheus `scrape_configs`, you can send the metrics using push gateway using a batch job.
Just configure the pushgateway url and use the `/probe` endpoint with query parameter `target` and the metrics will be send to prometheus.

`curl http://localhost:9271/probe?target=https://www.example.com`


### Exporter Target Configuration (VIA PROMETHEUS)

Example configuration with simple and complex values

(Examples can ve found in the example folder)

```yaml

  - job_name: pagespeed_exporter_probe
      metrics_path: /probe
      # Re-Label configurations so that we can use them
      # to configure the pagespeed exporter
      relabel_configs:
        - source_labels: [__address__]
          target_label: __param_target
        - source_labels: [__param_target]
          target_label: instance
        - target_label: __address__
          replacement: "pagespeed_exporter:9271"
      static_configs:
        - targets:
            - 'https://example.com/' # Example PLAIN
            - '{"url":"https://example.com/","campaign":"test","locale":"en","source":"source"}'  
            - '{"url":"https://example.com/mobileonly","strategy":"mobile"}'                    

```


### Docker

```sh
$ docker run -p "9271:9271" --rm foomo/pagespeed_exporter -api-key {KEY} -t https://google.com,https://prometheus.io
```
or
```sh
$ docker run -p "9271:9271" --rm \
    --env PAGESPEED_API_KEY={KEY} \
    --env PAGESPEED_TARGETS=https://google.com,https://prometheus.io \
    --env PAGESPEED_CATEGORIES=accessibility,seo \
    foomo/pagespeed_exporter
```


### Prometheus & Docker Compose

Check out the docker-compose folder

### Kubernetes/Helm

You can install the included [Helm](https://docs.helm.sh/install/) chart to your k8s cluster with:

```
$ helm install helm/pagespeed-exporter
```

And then, to quickly test it:
```
$ kubectl get pods
pagespeed-exporter-riotous-dragonfly-6b99955999-hj2kw   1/1     Running   0          1m

$ kubectl exec -ti pagespeed-exporter-riotous-dragonfly-6b99955999-hj2kw -- sh
# apk add curl
# curl localhost:9271/metrics
pagespeed_lighthouse_audit_score{audit="first-contentful-paint",host="https://www.google.com",path="/",strategy="mobile"} 1
pagespeed_lighthouse_audit_score{audit="first-contentful-paint",host="https://www.google.com",path="/webhp",strategy="desktop"} 1
pagespeed_lighthouse_audit_score{audit="first-contentful-paint",host="https://www.google.com",path="/webhp",strategy="mobile"} 1
...
```
