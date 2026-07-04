GOFILES := $(shell find . -name "*.go" -type f ! -path "./vendor/*" ! -path "*/bindata.go")
GOFMT ?= gofmt -s
GOLANGCI_LINT_VERSION ?= v2.12.2

.PHONY: install
install:
		mkdir -p ${HOME}/.config/ghorg
		cp sample-conf.yaml ${HOME}/.config/ghorg/conf.yaml

.PHONY: homebrew
homebrew:
		mkdir -p ${HOME}/.config/ghorg
		cp sample-conf.yaml ${HOME}/.config/ghorg/conf.yaml

.PHONY: fmt
fmt:
		$(GOFMT) -w $(GOFILES)

.PHONY: lint-install
lint-install:
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

.PHONY: lint
lint:
		golangci-lint run ./...

.PHONY: release
release:
		goreleaser release


.PHONY: examples
examples:
		cp -rf examples/* cmd/examples-copy/
