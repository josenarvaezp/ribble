FROM alpine as coordinator

ARG CGO_ENABLED=0

RUN apk add --update \
    make \
    && rm -rf /var/cache/apk/*

# create work directory
WORKDIR /build

ADD ./Makefile ./Makefile