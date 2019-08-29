FROM cdkbot-layer

ARG GO_VERSION=1.12.7

USER root
RUN curl https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xzf - && \
    mkdir -p /go/pkg/mod
ENV PATH $PATH:/usr/local/go/bin
ENV GO111MODULE=on
ENV GOPATH=/go

WORKDIR /root/cdkbot
ENTRYPOINT []