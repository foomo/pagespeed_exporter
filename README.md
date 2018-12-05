# Prometheus Exporter for Google Pagespeed Online Metrics
[![Travis CI](https://travis-ci.org/foomo/pagespeed_exporter.svg?branch=master)](https://travis-ci.org/foomo/pagespeed_exporter)


## Understanding Metrics

* https://github.com/GoogleChrome/lighthouse/blob/master/docs/understanding-results.md
* https://developers.google.com/speed/docs/insights/v5/reference/pagespeedapi/runpagespeed


Prometheus exporter for google pagespeed metrics


## Building And Running

### Building

```sh
make
```

### Examples 

To run pagespeed exporter we need to obtain the google api key for the pagespeed. 
Instructions how to create a key for pagespeed can be found [here](https://developers.google.com/speed/docs/insights/v2/first-app)

pagespeed_exporter <arguments>

```sh
pagespeed_exporter -api-key {KEY} -targets https://google.com,https://prometheus.io -listener :80
```

### Exporter configuration

| Flag      | Variable           | Description                                 | Default | Required |
|-----------|--------------------|---------------------------------------------|---------|----------|
| -api-key  | PAGESPEED_API_KEY  | sets the google API key used for pagespeed  |         | True     |
| -targets  | PAGESPEED_TARGETS  | comma separated list of targets to measure  |         | False    |
| -t        | NONE               | multi-value target array (check docker comp)|         | False    |
| -listener | PAGESPEED_LISTENER | sets the listener address for the exporters | :9271   | False    |

Note: google api key is required only if scraping more than 2 targets/second
### Docker

```sh
docker run -p "9271:9271" --rm foomo/pagespeed_exporter -api-key {KEY} -t https://google.com,https://prometheus.io
```
or
```sh
docker run -p "9271:9271" --rm \
    --env PAGESPEED_API_KEY={KEY} \
    --env PAGESPEED_TARGETS=https://google.com,https://prometheus.io \
    foomo/pagespeed_exporter
```


### Prometheus & Docker Compose

Check out the docker-compose folder