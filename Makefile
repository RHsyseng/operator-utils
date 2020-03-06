.PHONY: all
all: test

.PHONY: mod
mod:
	go mod tidy && go mod vendor

.PHONY: format
format: mod
	go fmt ./...

.PHONY: vet
vet: format
	go vet -mod=vendor ./...

.PHONY: test
test: vet
	go test -mod=vendor ./...
