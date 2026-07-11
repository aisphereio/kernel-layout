# Build, Test, and Deploy

## Makefile Targets

The layout provides a comprehensive Makefile (`Makefile`) with the following targets:

### Toolchain

| Target | Description |
|--------|-------------|
| `make tools` | Install codegen tools into `.bin/` (from published Kernel module) |
| `make tools-local KERNEL_LOCAL=../kernel` | Install codegen tools from local Kernel checkout |
| `make check-tools` | Verify required tools exist in `.bin/` |

### Code Generation

| Target | Description |
|--------|-------------|
| `make api` | Run `buf generate` with `buf.gen.yaml` — generates Go types, gRPC stubs, HTTP handlers, authz resolvers, gateway resolvers, Kernel bindings, OpenAPI specs |
| `make deploy` | Run `buf generate` with `buf.gen.deploy.yaml` — generates Gateway API HTTPRoute manifests |
| `make config` | Generate config proto code (if `buf.gen.config.yaml` exists) |
| `make wire` | Run `wire` for dependency injection |
| `make generate` | Run `go generate` |

### Quality

| Target | Description |
|--------|-------------|
| `make proto-check` | Run `buf lint` and `buf build` |
| `make test` | Run `go test ./...` |
| `make tidy` | Run `go mod tidy` |

### Build & Run

| Target | Description |
|--------|-------------|
| `make build` | Build service binary to `bin/` |
| `make run` | Run service locally with `configs/config.yaml` |
| `make verify` | Full pipeline: api → deploy → config → wire → generate → tidy → test → build |

### Cleanup

| Target | Description |
|--------|-------------|
| `make clean` | Remove `.bin/` and `bin/` directories |

## CI Pipeline

Defined in `.github/workflows/go.yml`:

- Trigger: push/PR to `main`
- Steps: checkout → setup Go → cache modules → get dependencies → build → test

## Docker

```dockerfile
# Multi-stage build
FROM golang:1.25 AS builder
COPY . /src
WORKDIR /src
RUN GOPROXY=https://goproxy.cn make build

FROM debian:stable-slim
COPY --from=builder /src/bin /app
EXPOSE 8000 9000
CMD ["./server", "-conf", "/data/conf"]
```

- Uses `goproxy.cn` for Go module proxy (China-friendly)
- Exposes port 8000 (HTTP) and 9000 (gRPC)
- Config mounted at `/data/conf`

## Deploy Manifest Generation

`make deploy` runs `protoc-gen-go-deploy` which reads proto annotations and writes Kubernetes Gateway API `HTTPRoute` manifests:

```text
deploy/generated/gateway/
├── public/           # PUBLIC exposure
├── authenticated/    # AUTHENTICATED / AUTHORIZED exposure
└── internal/         # INTERNAL / SYSTEM exposure
```

The generator reads both `google.api.http` and `aisphere.access.v1.policy` annotations to produce routes with:
- HTTP method/path
- Upstream gRPC operation
- Exposure level
- Edge authn mode
- Authz action/resource headers

## OpenWiki Update

Defined in `.github/workflows/openwiki-update.yml`:

- Schedule: daily at 08:00 UTC
- Creates a PR with updated `openwiki/` docs
- Uses `openwiki code --update --print` with OpenRouter API

## Source Map

| File | Role |
|------|------|
| `Makefile` | All build/test/deploy targets |
| `Dockerfile` | Multi-stage Docker build |
| `.github/workflows/go.yml` | CI pipeline |
| `.github/workflows/openwiki-update.yml` | Scheduled doc update |
| `buf.gen.yaml` | Main codegen config |
| `buf.gen.deploy.yaml` | Deploy manifest codegen |
| `buf.gen.grpc-gateway.yaml` | Standalone grpc-gateway codegen |
| `buf.gen.config.yaml` | Config proto codegen |
| `buf.yaml` | Buf module config |