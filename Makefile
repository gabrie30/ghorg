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

.PHONY: test
test:
		go test ./... -v

.PHONY: test-git
test-git:
		go test ./git -v

.PHONY: test-coverage
test-coverage:
		go test ./... -coverprofile=coverage.out -covermode=atomic
		go tool cover -html=coverage.out -o coverage.html
		@echo "Coverage report generated: coverage.html"

.PHONY: test-coverage-func
test-coverage-func:
		@cd git && go test -coverprofile=../coverage.out -covermode=atomic
		@go tool cover -func=coverage.out
		@echo ""
		@echo "=== New Git Helper Functions Coverage ==="
		@go tool cover -func=coverage.out | grep -E "(GetRemoteURL|HasLocalChanges|HasUnpushedCommits|GetCurrentBranch|HasCommitsNotOnDefaultBranch|IsDefaultBranchBehindHead|MergeIntoDefaultBranch|UpdateRef)"

.PHONY: test-all
test-all: test
		@echo ""
		@echo "=== All Tests Complete ==="

.PHONY: test-sync
test-sync:
		go test ./git -v -run "TestSync"

.PHONY: test-helpers
test-helpers:
		go test ./git -v -run "^Test(GetRemoteURL|HasLocalChanges|HasUnpushedCommits|GetCurrentBranch|HasCommitsNotOnDefaultBranch|IsDefaultBranchBehindHead|MergeIntoDefaultBranch|UpdateRef)"

.PHONY: release
release:
		goreleaser release


.PHONY: examples
examples:
		cp -rf examples/* cmd/examples-copy/
