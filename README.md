# Pagespeed Exporter

Prometheus exporter for google pagespeed metrics


## Building And Running

## Building

```sh
make
```

### Examples 

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
| -listener | sets the listener address for the exporters | :8080   | False    |


### Docker

```sh
docker run foomo/pagespeed_exporter -api-key {KEY} -targets https://google.com,https://prometheus.io -listener :80
```