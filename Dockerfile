FROM        alpine:latest
MAINTAINER  Stefan Martinov <stefan.martinov@bestbytes.com>

COPY pagespeed_exporter /bin/pagespeed_exporter

EXPOSE      8080

ENTRYPOINT  [ "/bin/pagespeed_exporter" ]