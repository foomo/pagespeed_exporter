FROM        alpine:latest
MAINTAINER  Stefan Martinov <stefan.martinov@bestbytes.com>

RUN apk --no-cache add ca-certificates && update-ca-certificates

COPY pagespeed_exporter /bin/pagespeed_exporter

RUN chmod +x /bin/pagespeed_exporter

EXPOSE      8080

ENTRYPOINT  [ "/bin/pagespeed_exporter" ]