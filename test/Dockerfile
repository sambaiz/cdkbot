

FROM lambci/lambda:build-go1.x AS build

ARG GO_VERSION=1.12.7

USER root
RUN curl https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xzf -
ENV PATH $PATH:/usr/local/go/bin
ENV GO111MODULE=on

WORKDIR /root/cdkbot
ENTRYPOINT []