default: all

all: test

tidy:
	@echo "Tidying up..." && go mod tidy
	@echo "Tidying up... Done!"

fmt: tidy
	@echo "Formatting..." && \
		echo "  .. fmt ..." && go fmt ./... && \
		echo "  .. fix ..." && go fix ./... && \
		echo "  .. vet ..." && go vet ./... && \
		(command -v gofmt >/dev/null 2>&1 && echo "  .. gofmt ..." && gofmt -s -w . || true) && \
		(command -v goimports >/dev/null 2>&1 && echo "  .. goimports ..." && goimports -w . || true) && \
		(command -v golines >/dev/null 2>&1 && echo "  .. golines ..." && golines -w . || true) && \
		(command -v gofumpt >/dev/null 2>&1 && echo "  .. gofumpt ..." && gofumpt -w . || true)
	@echo "Formatting... Done!"

lint: fmt
	@echo "Linting..." && golangci-lint run ./...
	@echo "Linting... Done!"

test: lint
	@echo "Testing..." && go test -race -v ./...
	@echo "Testing... Done!"

.PHONY: default all tidy fmt lint test