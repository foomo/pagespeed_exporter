# -----------------------------------------------------------------------------
# Builder Base
# -----------------------------------------------------------------------------
FROM --platform=$BUILDPLATFORM golang:alpine as base
LABEL maintainer="Stefan Martinov <stefan.martinov@bestbytes.com>"

RUN apk add --no-cache git \
  && rm -rf /var/cache/apk/*

WORKDIR /app

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY . ./

##############################
###### STAGE: BUILD     ######
##############################
FROM base as builder
LABEL maintainer="Stefan Martinov <stefan.martinov@bestbytes.com>"

ARG TARGETOS
ARG TARGETARCH

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -ldflags "-X main.Version=`git rev-parse --short HEAD`" -o /pagespeed_exporter pagespeed_exporter.go

##############################
###### STAGE: PACKAGE   ######
##############################
FROM alpine:3.20.3
LABEL maintainer="Stefan Martinov <stefan.martinov@bestbytes.com>"

COPY --from=builder /pagespeed_exporter /bin/pagespeed_exporter

RUN apk update \
  && apk --no-cache add ca-certificates

EXPOSE      9271

ENTRYPOINT  [ "/bin/pagespeed_exporter" ]

