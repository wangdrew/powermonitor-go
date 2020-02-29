SHELL := /bin/bash  # make to use bash (instead of default sh)
GOCMD=go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOFMT = $(GOCMD) fmt
ifndef ($(GOPATH))
	GOPATH = $(HOME)/go
endif
BINARY_NAME = sensor
PWD = $(shell pwd)


.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on $(GOBUILD) -o $(BINARY_NAME)

.PHONY: buildarm6
buildarm6:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 GO111MODULE=on $(GOBUILD) -o $(BINARY_NAME)

.PHONY: buildarm7
buildarm7:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 GO111MODULE=on $(GOBUILD) -o $(BINARY_NAME)

.PHONY: buildarm8
buildarm8:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GO111MODULE=on $(GOBUILD) -o $(BINARY_NAME)

.PHONY: run
run:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on $(GOBUILD) -o $(BINARY_NAME)
	./$(BINARY_NAME)

.PHONY: runarm6
runarm6:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 GO111MODULE=on $(GOBUILD) -o $(BINARY_NAME)
	./$(BINARY_NAME)

.PHONY: runarm7
runarm7:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 GO111MODULE=on $(GOBUILD) -o $(BINARY_NAME)
	./$(BINARY_NAME)

.PHONY: runarm8
runarm8:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GO111MODULE=on $(GOBUILD) -o $(BINARY_NAME)
	./$(BINARY_NAME)

.PHONY: test
test:
	$(GOTEST)

.PHONY: format
format:
	$(GOFMT) -x ./...
