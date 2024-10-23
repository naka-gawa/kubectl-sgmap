SHELL := /bin/bash
INSTALL_DIR ?= /usr/bin
PLUGIN_BIN ?= kubectl-sgmap
PLUGIN_DEPENDENCIES := $(shell find . -name "*.go")
# if you want to execute gotest with verbosity, set this flag to `true`.
TEST_VERBOSE ?= true

build: format test $(PLUGIN_BIN)

format:
	go fmt ./...

test:
ifeq ($(TEST_VERBOSE), true)
	go test -v ./... -count=1
else
	go test ./... -count=1
endif

$(PLUGIN_BIN): $(PLUGIN_DEPENDENCIES)
	go build -o $(PLUGIN_BIN) ./cmd/$(PLUGIN_BIN)/main.go

generate:
	kubectl-plugin-builder generate

install: $(PLUGIN_BIN)
	mv $(PLUGIN_BIN) $(INSTALL_DIR)
