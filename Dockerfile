FROM golang:1.15-alpine

WORKDIR $GOPATH/bin

WORKDIR /opt/db
ENV LS_BOLT_DB /opt/db/bolt.db

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...

RUN go build -o "long-season" ./cmd/server/main.go
RUN mv "long-season" $GOPATH/bin/

RUN go build -o "long-season-cli" ./cmd/cli/main.go
RUN mv "long-season-cli" $GOPATH/bin/
