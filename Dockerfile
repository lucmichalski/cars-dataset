FROM golang:alpine3.11 AS builder
MAINTAINER lucmichalski <michalski.luc@gmail.com>

RUN apk add --no-cache make gcc g++ ca-certificates musl-dev make git

COPY . /go/src/github.com/lucmichalski/cars-dataset
WORKDIR /go/src/github.com/lucmichalski/cars-dataset

RUN make plugins && \
    go install

FROM zenika/alpine-chrome:latest AS runtime
# FROM alpine:3.11 AS runtime
MAINTAINER lucmichalski <michalski.luc@gmail.com>

USER root

ARG TINI_VERSION=${TINI_VERSION:-"v0.18.0"}

# Install tini to /usr/local/sbin
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-muslc-amd64 /usr/local/sbin/tini

# Install runtime dependencies & create runtime user
RUN apk --no-cache --no-progress add ca-certificates \
 && chmod +x /usr/local/sbin/tini && mkdir -p /opt \
 && adduser -D lucmichalski -h /opt/peaks-tires -s /bin/sh \
 && su lucmichalski -c 'cd /opt/peaks-tires; mkdir -p bin config data ui'

# Switch to user context
# USER lucmichalski
WORKDIR /opt/lucmichalski/bin

# copy executable
COPY --from=builder /go/bin/cars-dataset /opt/lucmichalski/bin/cars-dataset
COPY --from=builder /go/src/github.com/lucmichalski/cars-dataset/release/* /opt/lucmichalski/bin/release/

ENV PATH $PATH:/opt/lucmichalski/bin

USER chrome

# Container configuration
EXPOSE 9000
VOLUME ["/opt/lucmichalski/bin/public"]
ENTRYPOINT ["tini", "-g", "--"]
CMD ["/opt/lucmichalski/bin/cars-dataset"]

