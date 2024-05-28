#!/usr/bin/make -f
# SPDX-License-Identifier: Apache-2.0
# (c) 2024, Konstantin Demin

SHELL       :=/bin/sh
.SHELLFLAGS :=-ec

.NOTPARALLEL:

BIN := publish-nexus

OUTDIR ?= .
OUTSFX ?=
OUTBIN ?= $(OUTDIR)/$(BIN)$(OUTSFX)

export GO ?= go
export CGO_ENABLED ?= 0
TAGS ?=
LDFLAGS ?=
GO_BUILDFLAGS ?=
GO_LDFLAGS := -w $(LDFLAGS)

comma :=,
ifeq ($(RELMODE),1)
  ## not ready yet
  # TAGS := nodebug$(if $(strip $(TAGS)),$(comma)$(strip $(TAGS)))
  GO_LDFLAGS += -s
endif

.PHONY: all
all: build

.PHONY: clean build dev-build ci-clean

clean:
	$(if $(wildcard $(OUTBIN)),rm -fv $(OUTBIN),:)

build: $(OUTBIN)

test_git = git -c log.showsignature=false show -s --format=%H:%ct

$(OUTBIN):
	@:; \
	GO_BUILDFLAGS='$(strip $(GO_BUILDFLAGS))' ; \
	if ! $(test_git) >/dev/null 2>&1 ; then \
	    echo "!!! git information is asbent !!!" >&2 ; \
	    GO_BUILDFLAGS="-buildvcs=false $${GO_BUILDFLAGS}" ; \
	fi ; \
	for i in $$(seq 1 3) ; do \
	    if $(GO) get ; then break ; fi ; \
	done ; \
	$(GO) build -o $@ \
	  $${GO_BUILDFLAGS} \
	  $(if $(strip $(TAGS)),-tags '$(strip $(TAGS))') \
	  $(if $(strip $(GO_LDFLAGS)),-ldflags '$(strip $(GO_LDFLAGS))') \
	  $(if $(VERBOSE),-v) ; \
	$(GO) version -m $@

dev-build: GO_BUILDFLAGS := -race $(GO_BUILDFLAGS)
dev-build: CGO_ENABLED := 1
dev-build: RELMODE := 0
dev-build: build

ci-clean:
	for d in '$(shell $(GO) env GOCACHE)' '$(shell $(GO) env GOMODCACHE)' ; do \
	    [ -n "$$d" ] || continue ; \
	    [ -d "$$d" ] || continue ; \
	    rm -rf "$$d" ; \
	done
