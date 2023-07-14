ROOT=github.com/imxw/gitlab-scaffold
SELF_DIR=$(dir $(lastword $(MAKEFILE_LIST)))
CMD=glfast
GOOS=$(shell go env GOOS)
GOPATH=$(shell go env GOPATH)
GOARCH=$(shell go env GOARCH)


# COLORS
RED    = $(shell printf "\33[31m")
GREEN  = $(shell printf "\33[32m")
WHITE  = $(shell printf "\33[37m")
YELLOW = $(shell printf "\33[33m")
RESET  = $(shell printf "\33[0m")

ifeq ($(origin ROOT_DIR),undefined)
ROOT_DIR := $(abspath $(shell cd $(SELF_DIR) && pwd -P))
endif

FIND := find . -path './cmd/*.go' -o -path './internal/**/*.go' -o -path './main.go'

.PHONY: all
all: build

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9.%-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


.PHONY: build
build: mod fmt vet ## Build glfast binary.
	go build -trimpath -o ${CMD} .
	@echo "${GREEN}✔[${CMD}] has been generated in the current directory($(PWD))!${RESET}"

.PHONY: clean
clean: ## Remove glfast.
	-rm -f ${CMD}

.PHONY: fmt
fmt: verify.goimports ## Run 'go fmt' & goimports against code.
	@echo "$(YELLOW)Formating codes$(RESET)"
	@$(FIND) -type f | xargs gofmt -s -w
	@$(FIND) -type f | xargs goimports -w -local $(ROOT)
	@go mod edit -fmt

.PHONY: lint
lint: verify.golangcilint ## Run 'golangci-lint' against code.
	@echo "$(YELLOW)Run golangci to lint source codes$(RESET)"
	@golangci-lint -c $(ROOT_DIR)/.golangci.yml run $(ROOT_DIR)/...

.PHONY: vet
vet: ## Run "go vet ./...".
	go vet ./...

.PHONY: mod
mod: ## Run "go mod tidy".
	go mod tidy

.PHONY: verify.%
verify.%:
	@if ! command -v $* >/dev/null 2>&1; then $(MAKE) install.$*; fi

.PHONY: install.goimports
install.goimports:
	@go install golang.org/x/tools/cmd/goimports@latest

.PHONY: install.golangcilint
install.golangcilint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
