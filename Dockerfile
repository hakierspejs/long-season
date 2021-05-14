FROM golang:1.16-alpine

WORKDIR $GOPATH/bin

RUN apk add --no-cache git

WORKDIR /opt/db
ENV LS_BOLT_DB /opt/db/bolt.db

ENV SRC /go/src/app
WORKDIR $SRC
COPY . .

RUN go get -d -v ./...

RUN go build -o "long-season" ./cmd/server/main.go \
    && mv "long-season" $GOPATH/bin/

RUN go build -o "long-season-cli" ./cmd/cli/main.go \
    && mv "long-season-cli" $GOPATH/bin/

WORKDIR /
