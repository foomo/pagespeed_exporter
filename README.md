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

The  dashboard can be found at [grafana](https://grafana.com/dashboards/9510)

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

### Exporter configuration

| Flag      | Variable           | Description                                   | Default | Required |
|-----------|--------------------|-----------------------------------------------|---------|----------|
| -api-key  | PAGESPEED_API_KEY  | sets the google API key used for pagespeed    |         | False    |
| -targets  | PAGESPEED_TARGETS  | comma separated list of targets to measure    |         | False    |
| -t        | NONE               | multi-value target array (check docker comp)  |         | False    |
| -listener | PAGESPEED_LISTENER | sets the listener address for the exporters   | :9271   | False    |
| -parallel | PAGESPEED_PARALLEL | sets the execution of targets to be parallel  | false   | False    |

Note: google api key is required only if scraping more than 2 targets/second
### Docker

```sh
$ docker run -p "9271:9271" --rm foomo/pagespeed_exporter -api-key {KEY} -t https://google.com,https://prometheus.io
```
or
```sh
$ docker run -p "9271:9271" --rm \
    --env PAGESPEED_API_KEY={KEY} \
    --env PAGESPEED_TARGETS=https://google.com,https://prometheus.io \
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
