##############################
###### STAGE: BUILD     ######
##############################
FROM golang:1.11.2 AS build-env
MAINTAINER  Stefan Martinov <stefan.martinov@bestbytes.com>

ENV DEP_VERSION 0.5.0
ENV UPX_VERSION 3.95
ENV SOURCE /go/src/github.com/foomo/pagespeed_exporter

# Install DEP (Golang Dependency Management)
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 && chmod +x /usr/local/bin/dep

RUN apt-get update && apt-get install -y xz-utils && rm -rf /var/lib/apt/lists/*

# install UPX
ADD https://github.com/upx/upx/releases/download/v${UPX_VERSION}/upx-${UPX_VERSION}-amd64_linux.tar.xz /usr/local
RUN xz -d -c /usr/local/upx-${UPX_VERSION}-amd64_linux.tar.xz | \
    tar -xOf - upx-${UPX_VERSION}-amd64_linux/upx > /bin/upx && \
    chmod a+x /bin/upx

WORKDIR ${SOURCE}

# install the dependencies without checking for go code, cache results
ADD Gopkg.lock ${SOURCE}/Gopkg.lock
ADD Gopkg.toml ${SOURCE}/Gopkg.toml
RUN dep ensure -vendor-only

ADD . ${SOURCE}

RUN GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build \
    -ldflags "-X main.Version=`git rev-parse --short HEAD`" \
    -o /pagespeed_exporter pagespeed_exporter.go

# strip and compress the binary
RUN strip --strip-unneeded /pagespeed_exporter
RUN upx /pagespeed_exporter

##############################
###### STAGE: PACKAGE   ######
##############################
FROM alpine
MAINTAINER  Stefan Martinov <stefan.martinov@bestbytes.com>

COPY --from=build-env /pagespeed_exporter /bin/pagespeed_exporter

RUN apk update \
    && apk --no-cache add ca-certificates \
    && update-ca-certificates \
    && chmod +x /bin/pagespeed_exporter

EXPOSE      9271

ENTRYPOINT  [ "/bin/pagespeed_exporter" ]

