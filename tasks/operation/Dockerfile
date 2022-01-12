FROM node:16-alpine3.14

USER root
WORKDIR /root/cdkbot

RUN apk add --no-cache make gcc libc-dev git docker && \
    git config --global user.name cdkbot && \
    git config --global user.email operation@cdkbot.localhost

COPY ./tasks/operation/operation .

ENTRYPOINT ["./operation"]