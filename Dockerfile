FROM golang:1.15-alpine

WORKDIR $GOPATH/bin

RUN apk add --no-cache git \
    && export GO111MODULE=off \
    && go get 4d63.com/embedfiles \
    && export GO111MODULE=on


WORKDIR /opt/db
ENV LS_BOLT_DB /opt/db/bolt.db

ENV SRC /go/src/app
WORKDIR $SRC
COPY . .

RUN go get -d -v ./...
RUN embedfiles -out=pkg/static/files.gen.go -pkg=static web


RUN go build -o "long-season" ./cmd/server/main.go \
    && mv "long-season" $GOPATH/bin/

RUN go build -o "long-season-cli" ./cmd/cli/main.go \
    && mv "long-season-cli" $GOPATH/bin/

WORKDIR /
RUN rm -rf $SRC $GOPATH/bin/embedfiles
