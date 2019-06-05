# -----------------------------------------------------------------------------
# Builder Base
# -----------------------------------------------------------------------------
FROM golang:alpine as base
MAINTAINER  Stefan Martinov <stefan.martinov@bestbytes.com>

RUN apk add --no-cache git make bash upx \
  && rm -rf /var/cache/apk/*

WORKDIR /app

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY . ./

##############################
###### STAGE: BUILD     ######
##############################
FROM base as builder
MAINTAINER  Stefan Martinov <stefan.martinov@bestbytes.com>

RUN go build -ldflags "-X main.Version=`git rev-parse --short HEAD`" -o /pagespeed_exporter pagespeed_exporter.go

# strip and compress the binary
RUN upx /pagespeed_exporter

##############################
###### STAGE: PACKAGE   ######
##############################
FROM alpine
MAINTAINER  Stefan Martinov <stefan.martinov@bestbytes.com>

COPY --from=builder /pagespeed_exporter /bin/pagespeed_exporter

RUN apk update \
    && apk --no-cache add ca-certificates

EXPOSE      9271

ENTRYPOINT  [ "/bin/pagespeed_exporter" ]

