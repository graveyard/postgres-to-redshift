include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

.PHONY: test vendor build
SHELL := /bin/bash
PKG := github.com/Clever/postgres-to-redshift
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := $(shell basename $(PKG))
$(eval $(call golang-version-check,1.8))

all: test build

build:
	go build -o bin/$(EXECUTABLE) $(PKG)

test: $(PKG)
$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

vendor: golang-godep-vendor-deps
	$(call golang-godep-vendor,$(PKGS))
