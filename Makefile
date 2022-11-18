GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
BUILD_DIR = dist/${GOOS}_${GOARCH}
OUTPUT_PATH = ${BUILD_DIR}/$(notdir $(CURDIR))

.PHONY: build
build:
	rm -f ${OUTPUT_PATH}
	mkdir -p ${BUILD_DIR}
	go build -o ${OUTPUT_PATH} cmd/*.go

.PHONY: update-deps
update-deps:
	GOPRIVATE=github.com/conductorone/baton-sdk go get -d -u ./...
	go mod tidy -v
	go mod vendor

.PHONY: add-deps
add-dep:
	go mod tidy -v
	go mod vendor

.PHONY: lint
lint:
	golangci-lint run

.PHONY: protogen
protogen:
	buf generate