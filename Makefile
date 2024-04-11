PROJECT := temporal-activities
MAKEDIR := hack/make
SHELL   := /bin/bash

.DEFAULT_GOAL := help
.PHONY: *

DBG_MAKEFILE ?=
ifeq ($(DBG_MAKEFILE),1)
    $(warning ***** starting Makefile for goal(s) "$(MAKECMDGOALS)")
    $(warning ***** $(shell date))
else
    # If we're not debugging the Makefile, don't echo recipes.
    MAKEFLAGS += -s
endif

include hack/make/bootstrap.mk
include hack/make/dep_golangci_lint.mk
include hack/make/dep_golines.mk
include hack/make/dep_gomajor.mk
include hack/make/dep_gosec.mk
include hack/make/dep_gotestsum.mk
include hack/make/dep_tparse.mk
include hack/make/dep_workflowcheck.mk

# Lazy-evaluated list of tools.
TOOLS = $(GOLANGCI_LINT) \
	$(GOLINES) \
	$(GOMAJOR) \
	$(GOSEC) \
	$(GOTESTSUM) \
	$(TPARSE)

define NEWLINE


endef

IGNORED_PACKAGES := \
	github.com/artefactual-sdps/temporal-activities/hack/%
PACKAGES := $(shell go list ./...)
TEST_PACKAGES := $(filter-out $(IGNORED_PACKAGES),$(PACKAGES))
TEST_IGNORED_PACKAGES := $(filter $(IGNORED_PACKAGES),$(PACKAGES))

export PATH:=$(GOBIN):$(PATH)

env: # @HELP Print Go env variables.
env:
	go env

deps: # @HELP List available module dependency updates.
deps: $(GOMAJOR)
	gomajor list

golines: # @HELP Run the golines formatter to fix long lines.
golines: $(GOLINES)
	golines \
		--chain-split-dots \
		--ignored-dirs="$(TEST_IGNORED_PACKAGES)" \
		--max-len=120 \
		--reformat-tags \
		--shorten-comments \
		--write-output \
		.

gosec: # @HELP Run gosec security scanner.
gosec: $(GOSEC)
	gosec \
		-terse \
		-exclude-dir=hack \
		./...

help: # @HELP Print this message.
help:
	echo "TARGETS:"
	grep -E '^.*: *# *@HELP' Makefile             \
	    | awk '                                   \
	        BEGIN {FS = ": *# *@HELP"};           \
	        { printf "  %-30s %s\n", $$1, $$2 };  \
	    '


lint: # @HELP Lint the project Go files with golangci-lint.
lint: OUT_FORMAT ?= colored-line-number
lint: LINT_FLAGS ?= --timeout=5m --fix
lint: $(GOLANGCI_LINT)
	golangci-lint run --out-format $(OUT_FORMAT) $(LINT_FLAGS)

list-tested-packages: # @HELP Print a list of packages being tested.
list-tested-packages:
	$(foreach PACKAGE,$(TEST_PACKAGES),@echo $(PACKAGE)$(NEWLINE))

list-ignored-packages: # @HELP Print a list of packages ignored in testing.
list-ignored-packages:
	$(foreach PACKAGE,$(TEST_IGNORED_PACKAGES),@echo $(PACKAGE)$(NEWLINE))

pre-commit: # @HELP Check that code is ready to commit.
pre-commit:
	$(MAKE) -j lint \
	golines \
	gosec \
	test-race

test: # @HELP Run all tests and output a summary using gotestsum.
test: TFORMAT ?= short
test: GOTEST_FLAGS ?=
test: COMBINED_FLAGS ?= $(GOTEST_FLAGS) $(TEST_PACKAGES)
test: $(GOTESTSUM)
	gotestsum --format=$(TFORMAT) -- $(COMBINED_FLAGS)

test-race: # @HELP Run all tests with the race detector.
test-race:
	$(MAKE) test GOTEST_FLAGS="-race"

test-ci: # @HELP Run all tests in CI with coverage and the race detector.
test-ci:
	$(MAKE) test GOTEST_FLAGS="-race -coverprofile=covreport -covermode=atomic"

tools: # @HELP Install tools.
tools: $(TOOLS)

tparse: # @HELP Run all tests and output a coverage report using tparse.
tparse: $(TPARSE)
	go test -count=1 -json -cover $(TEST_PACKAGES) | tparse -follow -all -notests
