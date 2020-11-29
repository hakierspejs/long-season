SERVER := $(CURDIR)/long-season
CLI := $(CURDIR)/long-season-cli
EMBEDFILES := $(CURDIR)/embedfiles
STATIC := pkg/static/files.gen.go

default: build

$(SERVER): $(STATIC) $(wildcard cmd/server/** pkg/**)
	go build -o $@ cmd/server/main.go

$(CLI): $(SERVER) $(wildcard cmd/cli/** pkg/**)
	go build -o $@ cmd/cli/main.go

$(EMBEDFILES):
	GO111MODULE=off go get 4d63.com/embedfiles
	GO111MODULE=off go build -o $@ 4d63.com/embedfiles

$(STATIC): $(EMBEDFILES) $(wildcard web/**)
	$(EMBEDFILES) -out=$@ -pkg=static web

.PHONY: default build run clean lint test

build: $(SERVER) $(CLI)

run: $(SERVER)
	$(SERVER)

clean:
	rm -f $(SERVER) $(CLI) $(EMBEDFILES) $(STATIC)

lint:
	golint ./...

test: $(STATIC)
	go test ./...
