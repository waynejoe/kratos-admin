APP := kratos-admin
CMD := ./cmd/kratos-admin
CONF ?= ./configs/config.yaml
BIN_DIR := ./bin
BIN := $(BIN_DIR)/$(APP)
GOCACHE ?= /tmp/go-build
API_PROTO_FILES := $(shell find api -name '*.proto')

.PHONY: all init tidy fmt api openapi wire test build run clean

all: tidy fmt api wire test build

init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest
	GOCACHE=$(GOCACHE) go mod tidy

tidy:
	GOCACHE=$(GOCACHE) go mod tidy

fmt:
	gofmt -w cmd internal pkg

api: openapi

openapi:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:./pb \
	       --go-http_out=paths=source_relative:./pb \
	       --go-grpc_out=paths=source_relative:./pb \
	       --go-errors_out=paths=source_relative:./pb \
	       --validate_out=paths=source_relative,lang=go:./pb \
	       --openapi_out=fq_schema_naming=true,default_response=false:. \
	       $(API_PROTO_FILES)

wire:
	GOCACHE=$(GOCACHE) wire ./cmd/kratos-admin

test:
	GOCACHE=$(GOCACHE) go test ./...

build:
	mkdir -p $(BIN_DIR)
	GOCACHE=$(GOCACHE) go build -o $(BIN) $(CMD)

run:
	GOCACHE=$(GOCACHE) go run $(CMD) -conf $(CONF)

clean:
	rm -rf $(BIN_DIR)
