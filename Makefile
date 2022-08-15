GO_TEST_OPT?= -race -timeout 16m -count=1
GO_TEST=go test

GO_BUILD=go build

.PHONY: build
build:
	$(GO_BUILD) -o ./bin/qms ./cmd/qms

.PHONY: test
test:
	$(GO_TEST) $(GO_TEST_OPT) -v ./...