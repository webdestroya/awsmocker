.PHONY: generate
generate:
	go generate ./...


.PHONY: tidy
tidy:
	go mod verify
	go mod tidy
	@if ! git diff --quiet go.mod go.sum; then \
		echo "please run go mod tidy and check in changes, you might have to use the same version of Go as the CI"; \
		exit 1; \
	fi

.PHONY: lint-install
lint-install:
	@echo "Installing golangci-lint"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2

.PHONY: lint
lint:
	@which golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found, please run: make lint-install"; \
		exit 1; \
	}
	@golangci-lint run && echo "All Good!"

.PHONY: test-release
test-release:
	goreleaser release --skip-publish --rm-dist --snapshot

.PHONY: test
test:
	go test -timeout 180s -v ./...

.PHONY: test-debug
test-debug:
	AWSMOCKER_DEBUG=true go test -timeout 180s -v ./...

.PHONY: coverage
coverage:
	@mkdir -p coverage
	@go test . -cover -coverprofile=coverage/c.out -covermode=count
	@go tool cover -html=coverage/c.out -o coverage/index.html