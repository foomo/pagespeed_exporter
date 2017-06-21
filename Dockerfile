FROM        quay.io/prometheus/busybox:latest

MAINTAINER  Stefan Martinov <stefan.martinov@bestbytes.com>

COPY pagespeed_exporter /bin/pagespeed_exporter

EXPOSE      9104

ENTRYPOINT  [ "/bin/pagespeed_exporter" ]