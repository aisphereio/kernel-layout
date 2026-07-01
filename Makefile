GO ?= go
BUF ?= .bin/buf
KERNEL_MODULE ?= github.com/aisphereio/kernel
KERNEL_VERSION ?= __KERNEL_VERSION__
CONF ?= ./configs/config.yaml
LOCAL_BIN := $(CURDIR)/.bin

ifeq ($(OS),Windows_NT)
LOCAL_BIN := $(CURDIR)\.bin
export PATH := $(LOCAL_BIN);$(PATH)
endif

.PHONY: help tools check-tools api proto-check run build test tidy verify clean

help:
	@echo "Kernel service layout targets:"
	@echo "  make tools        install local codegen tools into .bin"
	@echo "  make api          generate protobuf/http/grpc/gateway/kernel code"
	@echo "  make proto-check  run buf lint and Kernel proto contract check"
	@echo "  make run          run the generated service"
	@echo "  make build        build service binary"
	@echo "  make test         run tests"
	@echo "  make verify       run api, proto-check, tidy, test, build"

tools:
ifeq ($(OS),Windows_NT)
	@cmd /c "if not exist .bin mkdir .bin"
	@cmd /c "set GOBIN=$(LOCAL_BIN)&& $(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11"
	@cmd /c "set GOBIN=$(LOCAL_BIN)&& $(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1"
	@cmd /c "set GOBIN=$(LOCAL_BIN)&& $(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.29.0"
	@cmd /c "set GOBIN=$(LOCAL_BIN)&& $(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.29.0"
	@cmd /c "set GOBIN=$(LOCAL_BIN)&& $(GO) install github.com/bufbuild/buf/cmd/buf@v1.50.0"
	@cmd /c "set GOBIN=$(LOCAL_BIN)&& $(GO) install $(KERNEL_MODULE)/cmd/protoc-gen-go-http@$(KERNEL_VERSION)"
	@cmd /c "set GOBIN=$(LOCAL_BIN)&& $(GO) install $(KERNEL_MODULE)/cmd/protoc-gen-go-errors@$(KERNEL_VERSION)"
	@cmd /c "set GOBIN=$(LOCAL_BIN)&& $(GO) install $(KERNEL_MODULE)/cmd/protoc-gen-go-authz@$(KERNEL_VERSION)"
	@cmd /c "set GOBIN=$(LOCAL_BIN)&& $(GO) install $(KERNEL_MODULE)/cmd/protoc-gen-go-gateway@$(KERNEL_VERSION)"
	@cmd /c "set GOBIN=$(LOCAL_BIN)&& $(GO) install $(KERNEL_MODULE)/cmd/protoc-gen-go-kernel@$(KERNEL_VERSION)"
	@cmd /c "set GOBIN=$(LOCAL_BIN)&& $(GO) install $(KERNEL_MODULE)/cmd/buf-check-aisphere@$(KERNEL_VERSION)"
else
	@mkdir -p $(LOCAL_BIN)
	@GOBIN=$(LOCAL_BIN) $(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	@GOBIN=$(LOCAL_BIN) $(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	@GOBIN=$(LOCAL_BIN) $(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.29.0
	@GOBIN=$(LOCAL_BIN) $(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.29.0
	@GOBIN=$(LOCAL_BIN) $(GO) install github.com/bufbuild/buf/cmd/buf@v1.50.0
	@GOBIN=$(LOCAL_BIN) $(GO) install $(KERNEL_MODULE)/cmd/protoc-gen-go-http@$(KERNEL_VERSION)
	@GOBIN=$(LOCAL_BIN) $(GO) install $(KERNEL_MODULE)/cmd/protoc-gen-go-errors@$(KERNEL_VERSION)
	@GOBIN=$(LOCAL_BIN) $(GO) install $(KERNEL_MODULE)/cmd/protoc-gen-go-authz@$(KERNEL_VERSION)
	@GOBIN=$(LOCAL_BIN) $(GO) install $(KERNEL_MODULE)/cmd/protoc-gen-go-gateway@$(KERNEL_VERSION)
	@GOBIN=$(LOCAL_BIN) $(GO) install $(KERNEL_MODULE)/cmd/protoc-gen-go-kernel@$(KERNEL_VERSION)
	@GOBIN=$(LOCAL_BIN) $(GO) install $(KERNEL_MODULE)/cmd/buf-check-aisphere@$(KERNEL_VERSION)
endif

check-tools:
ifeq ($(OS),Windows_NT)
	@cmd /c "if not exist .bin\buf.exe echo missing .bin\buf.exe && exit /b 1"
else
	@test -x "$(LOCAL_BIN)/buf" || (echo "missing $(LOCAL_BIN)/buf"; exit 1)
endif

api: check-tools
ifeq ($(OS),Windows_NT)
	@cmd /c "set PATH=$(LOCAL_BIN);%PATH%&& .bin\buf.exe generate --template buf.gen.yaml"
else
	@PATH="$(LOCAL_BIN):$$PATH" $(LOCAL_BIN)/buf generate --template buf.gen.yaml
endif

proto-check: check-tools
ifeq ($(OS),Windows_NT)
	@cmd /c "set PATH=$(LOCAL_BIN);%PATH%&& .bin\buf.exe lint"
	@cmd /c "set PATH=$(LOCAL_BIN);%PATH%&& .bin\buf.exe build -o - | .bin\buf-check-aisphere.exe"
else
	@PATH="$(LOCAL_BIN):$$PATH" $(LOCAL_BIN)/buf lint
	@PATH="$(LOCAL_BIN):$$PATH" $(LOCAL_BIN)/buf build -o - | $(LOCAL_BIN)/buf-check-aisphere
endif

run:
	$(GO) run ./cmd/... -conf $(CONF)

build:
	$(GO) build -o ./bin/ ./cmd/...

test:
	$(GO) test ./...

tidy:
	$(GO) mod tidy

verify: api proto-check tidy test build

clean:
ifeq ($(OS),Windows_NT)
	@cmd /c "if exist .bin rmdir /s /q .bin"
	@cmd /c "if exist bin rmdir /s /q bin"
else
	rm -rf .bin bin
endif
