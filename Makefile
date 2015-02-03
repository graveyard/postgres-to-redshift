SHELL := /bin/bash
PKG := github.com/Clever/postgres-to-redshift

.PHONY: test

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
