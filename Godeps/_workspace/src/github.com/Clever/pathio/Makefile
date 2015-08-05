SHELL := /bin/bash
PKG = github.com/Clever/pathio
PKGS := $(PKG)
READMES = $(addsuffix /README.md, $(PKGS))

.PHONY: test golint README

golint:
	@go get github.com/golang/lint/golint

test: $(PKGS)
docs: $(READMES)
%/README.md:
	@go get github.com/robertkrimen/godocdown/godocdown
	@$(GOPATH)/bin/godocdown $(shell dirname $@) > $(GOPATH)/src/$@

$(PKGS): golint docs
	@go get -d -t $@
	@gofmt -w=true $(GOPATH)/src/$@*/**.go
	@echo "LINTING..."
	@PATH=$(PATH):$(GOPATH)/bin golint $(GOPATH)/src/$@*/**.go
	@echo ""
ifeq ($(COVERAGE),1)
	@go test -cover -coverprofile=$(GOPATH)/src/$@/c.out $@ -test.v
	@go tool cover -html=$(GOPATH)/src/$@/c.out
else
	@echo "TESTING..."
	@go test $@ -test.v
endif
