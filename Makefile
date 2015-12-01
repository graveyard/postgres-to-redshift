SHELL := /bin/bash
PKG := github.com/Clever/postgres-to-redshift
PKGS := $(shell go list ./... | grep -v /vendor)

.PHONY: test vendor

GOVERSION := $(shell go version | grep 1.5)
ifeq "$(GOVERSION)" ""
  $(error must be running Go version 1.5)
endif

export GO15VENDOREXPERIMENT = 1

test: $(PKG)

$(GOPATH)/bin/golint:
	go get github.com/golang/lint/golint

$(PKG): $(GOPATH)/bin/golint
	@echo ""
	@echo "FORMATTING $@..."
	go get -d -t $@
	gofmt -w=true $(GOPATH)/src/$@/*.go
	@echo ""
	@echo "LINTING $@..."
	$(GOPATH)/bin/golint $(GOPATH)/src/$@/*.go
	@echo ""


GODEP := $(GOPATH)/bin/godep

$(GODEP):
	go get -u github.com/tools/godep

vendor: $(GODEP)
	$(GODEP) save $(PKGS)
	find vendor/ -path '*/vendor' -type d | xargs -IX rm -r X # remove any nested vendor directories
