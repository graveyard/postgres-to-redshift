include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

.PHONY: test vendor build
SHELL := /bin/bash
PKG := github.com/Clever/postgres-to-redshift
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := $(shell basename $(PKG))
$(eval $(call golang-version-check,1.9))

all: test build

build:
	go build -o bin/$(EXECUTABLE) $(PKG)

test: $(PKG)
$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)



install_deps: golang-dep-vendor-deps
	$(call golang-dep-vendor)