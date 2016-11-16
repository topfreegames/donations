PACKAGES = $(shell glide novendor)
DIRS = $(shell find . -type f -not -path '*/\.*' | grep '.go' | egrep -v "^[.]\/vendor" | xargs -n1 dirname | sort | uniq | grep -v '^.$$')
OS = "$(shell uname | awk '{ print tolower($$0) }')"

setup:
	@go get -u github.com/onsi/ginkgo/ginkgo
	@go get -u github.com/Masterminds/glide/...
	@glide install

build:
	@go build $(PACKAGES)
	@go build

cross:
	@mkdir -p ./bin
	@echo "Building for linux-i386..."
	@env GOOS=linux GOARCH=386 go build -o ./bin/donations-linux-i386
	@echo "Building for linux-x86_64..."
	@env GOOS=linux GOARCH=amd64 go build -o ./bin/donations-linux-x86_64
	@echo "Building for darwin-i386..."
	@env GOOS=darwin GOARCH=386 go build -o ./bin/donations-darwin-i386
	@echo "Building for darwin-x86_64..."
	@env GOOS=darwin GOARCH=amd64 go build -o ./bin/donations-darwin-x86_64
	@chmod +x bin/*

services:
	@echo "Required services are up."

test: test-services
	@ginkgo --cover $(DIRS)

test-coverage: test
	@rm -rf _build
	@mkdir -p _build
	@echo "mode: count" > _build/test-coverage-all.out
	@bash -c 'for f in $$(find . -name "*.coverprofile"); do tail -n +2 $$f >> _build/test-coverage-all.out; done'

test-coverage-html: test-coverage
	@go tool cover -html=_build/test-coverage-all.out

test-services:
	@echo "Required test services are up."
