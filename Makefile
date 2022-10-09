GOARCH ?= $(shell go env GOARCH)
GOOS ?= $(shell go env GOOS)

GO_TEST_OPT?= -timeout 16m -coverprofile coverage.out -covermode atomic -count=1
GO_TEST=go test

GO_BUILD=go build

.PHONY: build
build:
	GOARCH=$(GOARCH) GOOS=$(GOOS) $(GO_BUILD) -o ./bin/qms ./cmd/qms

build-sut:
	GOARCH=$(GOARCH) GOOS=$(GOOS) $(GO_BUILD) -o ./bin/sut ./cmd/sut

.PHONY: test
test:
	$(GO_TEST) $(GO_TEST_OPT) -v ./...