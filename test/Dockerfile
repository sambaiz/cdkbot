FROM golang:1.13-alpine

USER root
ENV GO111MODULE=on
ENV GOPATH=/go

RUN mkdir -p /go/pkg/mod && \
    apk add --no-cache nodejs npm make gcc libc-dev git

WORKDIR /root/cdkbot
ENTRYPOINT []