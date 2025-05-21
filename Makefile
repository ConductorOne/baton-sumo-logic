GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
BUILD_DIR = dist/${GOOS}_${GOARCH}

ifeq ($(GOOS),windows)
OUTPUT_PATH = ${BUILD_DIR}/baton-sumo-logic.exe
else
OUTPUT_PATH = ${BUILD_DIR}/baton-sumo-logic
endif

.PHONY: build
build:
	go build -o ${OUTPUT_PATH} ./cmd/baton-sumo-logic

.PHONY: update-deps
update-deps:
	go get -d -u ./...
	go mod tidy -v
	go mod vendor

.PHONY: add-dep
add-dep:
	go mod tidy -v
	go mod vendor

.PHONY: lint
lint:
	golangci-lint run
