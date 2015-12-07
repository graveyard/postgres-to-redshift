SHELL := /bin/bash
PKG := github.com/Clever/postgres-to-redshift
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := postgres-to-redshift
.PHONY: test vendor build

GOVERSION := $(shell go version | grep 1.5)
ifeq "$(GOVERSION)" ""
  $(error must be running Go version 1.5)
endif
export GO15VENDOREXPERIMENT = 1

GOLINT := $(GOPATH)/bin/golint
$(GOLINT):
	go get github.com/golang/lint/golint

GODEP := $(GOPATH)/bin/godep
$(GODEP):
	go get -u github.com/tools/godep

build:
	go build -o bin/$(EXECUTABLE) $(PKG)

test: $(PKG)

$(PKG): $(GOLINT)
	@echo ""
	@echo "FORMATTING $@..."
	go get -d -t $@
	gofmt -w=true $(GOPATH)/src/$@/*.go
	@echo ""
	@echo "LINTING $@..."
	$(GOLINT) $(GOPATH)/src/$@/*.go
	@echo ""

vendor: $(GODEP)
	$(GODEP) save $(PKGS)
	find vendor/ -path '*/vendor' -type d | xargs -IX rm -r X # remove any nested vendor directories
