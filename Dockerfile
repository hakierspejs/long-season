FROM golang:1.15-alpine

WORKDIR $GOPATH/bin

RUN apk add --no-cache git make

WORKDIR /opt/db
ENV LS_BOLT_DB /opt/db/bolt.db

ENV SRC /go/src/app
WORKDIR $SRC
COPY . .

RUN make build SERVER=$GOPATH/bin/long-season CLI=$GOPATH/bin/long-season-cli

WORKDIR /
RUN rm -rf $SRC
