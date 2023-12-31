.DEFAULT_GOAL := build

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint: fmt
	golint ./...

.PHONY: vet
vet: fmt
	go vet ./...

.PHONY: build
build: vet
	go build -o gitFetchHelper

# run with the diff arg, as it doesn't make network calls
.PHONY: run
run:
	./gitFetchHelper diff

.PHONY: buildandrun
buildandrun: build run