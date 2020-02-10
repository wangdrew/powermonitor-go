SHELL := /bin/bash  # make to use bash (instead of default sh)
GOCMD=go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOFMT = $(GOCMD) fmt
ifndef ($(GOPATH))
	GOPATH = $(HOME)/go
endif
BINARY_NAME = powermonitor
PWD = $(shell pwd)


.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on $(GOBUILD) -o $(BINARY_NAME)

.PHONY: arm6build
arm6build:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 GO111MODULE=on $(GOBUILD) -o $(BINARY_NAME)

.PHONY: run
run:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on $(GOBUILD) -o $(BINARY_NAME)
	./$(BINARY_NAME)

.PHONY: arm6run
arm6run:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 GO111MODULE=on $(GOBUILD) -o $(BINARY_NAME)
	./$(BINARY_NAME)

.PHONY: test
test:
	$(GOTEST)

.PHONY: format
format:
	$(GOFMT) -x ./...