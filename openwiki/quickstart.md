# Aisphere Kernel Layout — OpenWiki

## Overview

This repository is the **template source** for `kernel new` — the project scaffolding command of the [Aisphere Kernel](https://github.com/aisphereio/kernel) framework. When you run `kernel new skill-service`, the contents of this layout are copied to create a new Go microservice that follows Kernel conventions.

The layout demonstrates a **proto-first, full-feature service** with:

- **Config** (`configx`): YAML file + env var overlay
- **Logging** (`logx`): structured, leveled, with redaction and access log
- **Metrics** (`metricsx`): Prometheus counters, histograms, gauges
- **Error handling** (`errorx`): structured error codes with HTTP/gRPC mapping
- **AuthN** (`authn`): gateway-trusted principal via `X-Aisphere-*` headers
- **AuthZ / Access** (`authz` / `accessx`): IAM gRPC-backed authorization with proto-declared policies
- **Audit** (`auditx`): event recording for security and compliance
- **DTM** (`dtmx`): distributed transaction (Saga/TCC) support
- **DB** (`dbx`): PostgreSQL via GORM
- **Cache** (`cachex`): Redis
- **Object storage** (`objectstorex`): MinIO/S3
- **Transports**: Kernel HTTP and gRPC servers with middleware chains
- **Deploy**: Gateway API `HTTPRoute` manifests generated from proto annotations

## Quick Start

```powershell
go install github.com/aisphereio/kernel/cmd/kernel@latest
kernel new skill-service
cd skill-service
make tools
make api
make deploy
make proto-check
make test
make run
```

## Repository Layout

| Path | Purpose |
|------|---------|
| `api/` | Protobuf API definitions and generated Go bindings |
| `cmd/server/` | Application entrypoint (renamed to `cmd/<service>` by `kernel new`) |
| `configs/` | Local config with IAM module defaults |
| `internal/conf/` | Config DTOs scanned by `configx` |
| `internal/server/` | Kernel HTTP and gRPC server construction with middleware |
| `internal/service/` | Transport-layer handlers (DTO conversion, no business logic) |
| `internal/biz/` | Business use cases, domain contracts, `errorx` errors |
| `internal/data/` | Repositories and Kernel resource initialization |
| `.kernel/` | Layout config/feature overlays consumed by `kernel new` |
| `docs/` | Architecture docs, module usage guides, design docs |

## Key Principles

1. **Proto-first**: All JSON HTTP/gRPC APIs must be declared in proto. RPCs must declare `google.api.http` and `aisphere.access.v1.policy`.
2. **Service/Biz/Data layering**: `service` does DTO conversion, `biz` owns business rules, `data` owns persistence and provider adapters.
3. **Kernel is the framework truth**: Business services use Kernel modules (`configx`, `logx`, `metricsx`, `authn`, `authz`, `accessx`, `auditx`, `dtmx`, `dbx`, `cachex`, `objectstorex`). No direct Casdoor/SpiceDB/MinIO/PostgreSQL client code in business layers.
4. **Gateway-trusted authn**: Backend services trust Envoy Gateway-injected `X-Aisphere-*` headers. No JWT parsing in service handlers.
5. **Deploy routes from proto**: Gateway API `HTTPRoute` manifests are generated from proto annotations, not hand-written YAML.

## Profiles and Features

The layout supports two profiles and feature toggles:

- **Full** (default): all modules enabled
- **MVP** (`--mvp`): minimal runnable skeleton (no IAM/authz/gateway/deploy generators)
- **Feature disable** (`--disable iam`, `--disable gateway,dtmx`): remove optional layout parts

## Documentation Map

| Section | Description |
|---------|-------------|
| [Architecture](architecture/overview.md) | System architecture, authn flow, gateway publication |
| [Modules](modules/README.md) | Per-module usage guides (config, log, metrics, error, authn, authz, audit, DTM) |
| [Domain](domain/todo.md) | Todo CRUD example — proto, service, biz, data layers |
| [Operations](operations/build-deploy.md) | Build, deploy, CI/CD, smoke test |
| [Existing Docs](../docs/modules/README.md) | Detailed module usage guides with production code examples |

## Source Map

| Concern | Key Files |
|---------|-----------|
| Entrypoint | `cmd/server/main.go` |
| Config DTOs | `internal/conf/conf.go` |
| Resource init | `internal/data/data.go` |
| Server construction | `internal/server/http.go`, `internal/server/grpc.go`, `internal/server/access.go` |
| Proto API | `api/todo/v1/todo.proto` |
| Access policy proto | `api/aisphere/access/v1/access.proto` |
| Business layer | `internal/biz/` |
| Service layer | `internal/service/` |
| Layout config | `.kernel/features/`, `.kernel/profiles/` |
| CI | `.github/workflows/go.yml` |
| OpenWiki CI | `.github/workflows/openwiki-update.yml` |
| Agent rules | `AGENTS.md`, `CLAUDE.md` |