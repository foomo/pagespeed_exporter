# Pagespeed Exporter
[![Travis CI](https://travis-ci.org/foomo/pagespeed_exporter.svg?branch=master)](https://travis-ci.org/foomo/pagespeed_exporter)

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

### Exporter CLI Arguments

| Flag      | Description                                 | Default | Required |
|-----------|---------------------------------------------|---------|----------|
| -api-key  | sets the google API key used for pagespeed  |         | True     |
| -targets  | comma separated list of targets to measure  |         | True     |
| -interval | check interval (e.g. 3s 4h 5d ...)          | 1h      | False    |
| -listener | sets the listener address for the exporters | :9271   | False    |


### Docker

```sh
docker run -p "9271:9271" --rm foomo/pagespeed-exporter -api-key {KEY} -targets https://google.com,https://prometheus.io
```
