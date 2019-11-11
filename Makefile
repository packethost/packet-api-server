SHELL=/bin/sh
PACKAGE_NAME?=github.com/packethost/packet-api-server
GIT_VERSION?=$(shell git log -1 --format="%h")
VERSION?=$(GIT_VERSION)
RELEASE_TAG ?= $(shell git tag --points-at HEAD)
ifneq (,$(RELEASE_TAG))
VERSION=$(RELEASE_TAG)-$(VERSION)
endif
GO_FILES := $(shell find . -type f -not -path './vendor/*' -name '*.go')


export GO111MODULE=on
BUILD_CMD = CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH)
ifdef DOCKERBUILD
BUILD_CMD = docker run --rm \
                -e GOARCH=$(ARCH) \
                -e GOOS=linux \
                -e CGO_ENABLED=0 \
                -v $(CURDIR):/go/src/$(PACKAGE_NAME) \
                -w /go/src/$(PACKAGE_NAME) \
		$(BUILDER_IMAGE)
endif

GOBIN ?= $(shell go env GOPATH)/bin
LINTER ?= $(GOBIN)/golangci-lint

pkgs:
ifndef PKG_LIST
	$(eval PKG_LIST := $(shell $(BUILD_CMD) go list ./... | grep -v vendor))
endif

.PHONY: fmt fmt-check lint test vet golint tag version

## report the git tag that would be used for the images
tag:
	@echo $(GIT_VERSION)

## report the version that would be put in the binary
version:
	@echo $(VERSION)


## Check the file format
fmt-check: 
	@if [ -n "$(shell $(BUILD_CMD) gofmt -l ${GO_FILES})" ]; then \
	  $(BUILD_CMD) gofmt -s -e -d ${GO_FILES}; \
	  exit 1; \
	fi

## fmt files
fmt:
	$(BUILD_CMD) gofmt -w -s $(GO_FILES)

golangci-lint: $(LINTER)
$(LINTER):
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.17.1

golint:
ifeq (, $(shell which golint))
	go get -u golang.org/x/lint/golint
endif

## Lint the files
lint: pkgs golint golangci-lint
	@$(BUILD_CMD) $(LINTER) run --disable-all --enable=golint ./ ./pkg

## Run unittests
test: pkgs
	@$(BUILD_CMD) go test -short ${PKG_LIST}

## Vet the files
vet: pkgs
	@$(BUILD_CMD) go vet ${PKG_LIST}

## Read about data race https://golang.org/doc/articles/race_detector.html
## to not test file for race use `// +build !race` at top
## Run data race detector
race: pkgs
	@$(BUILD_CMD) go test -race -short ${PKG_LIST}

## Display this help screen
help: 
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###############################################################################
# CI/CD
###############################################################################
.PHONY: ci cd deploy push release confirm pull-images
## Run what CI runs
# race has an issue with alpine, see https://github.com/golang/go/issues/14481
ci: fmt-check lint test vet # race

