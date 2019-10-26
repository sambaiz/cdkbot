FROM alpine:3.10

USER root
WORKDIR /root/cdkbot

RUN apk add --no-cache nodejs npm make gcc libc-dev git docker

COPY ./tasks/operation/operation .

ENTRYPOINT ["./operation"]