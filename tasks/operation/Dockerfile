FROM alpine:3.10

USER root
WORKDIR /root/cdkbot

RUN apk add --no-cache nodejs npm make gcc libc-dev git docker && \
    git config --global user.name cdkbot && \
    git config --global user.email operation@cdkbot.localhost

COPY ./tasks/operation/operation .

ENTRYPOINT ["./operation"]