FROM golang:1.25.6-alpine

WORKDIR $GOPATH/bin

RUN apk add --no-cache git

WORKDIR /opt/db
ENV LS_BOLT_DB /opt/db/bolt.db

ENV SRC /go/src/app
WORKDIR $SRC
COPY . .

RUN go get -d -v ./...

RUN go build -o "long-season" ./cmd/long-season/*.go \
    && mv "long-season" $GOPATH/bin/

RUN go build -o "short-season" ./cmd/short-season/*.go \
    && mv "short-season" $GOPATH/bin/

COPY Procfile .
