GOFILES := $(shell find . -name "*.go" -type f ! -path "./vendor/*" ! -path "*/bindata.go")
GOFMT ?= gofmt -s

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

.PHONY: release
release:
		goreleaser release
