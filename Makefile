.PHONY: all
all: test

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: format
format: tidy
	go fmt ./...

.PHONY: vet
vet: format
	go vet ./...

.PHONY: test
test: vet
	go test ./...
