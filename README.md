# Aisphere Kernel Layout

Standalone service layout for `github.com/aisphereio/kernel/cmd/kernel`.

This repository is copied by `kernel new` and becomes a new business service. It contains the standard service skeleton, runtime wiring, config shape, proto-first Todo example, and local development Makefile.

This layout intentionally does **not** include access/authz policy contracts. Access/authz governance should be added later as an optional feature template instead of being forced into the base service layout.

## Use

```powershell
go install github.com/aisphereio/kernel/cmd/kernel@latest
kernel new skill-service --repo https://github.com/aisphereio/kernel-layout.git
cd skill-service
make tools
make api
make proto-check
make test
make run
```

## Included defaults

- Features: `__KERNEL_FEATURES__`
- DB: `dbx` with `__KERNEL_DB_DRIVER__`
- Cache: `cachex` with `__KERNEL_CACHE_DRIVER__`
- Object storage: `objectstorex` with `__KERNEL_OBJECTSTORE_DRIVER__`
- Audit: `auditx` memory recorder by default
- Logging: `logx` console output for local development
- Metrics: shared `metricsx.Manager`, optional admin `/metrics` server
- DTM: optional `dtmx.Manager`
- Config: `configx` file source
- Transports: Kernel HTTP and gRPC servers with access log and metrics hooks
- API example: protobuf-first Todo CRUD, HTTP binding, gRPC binding, streaming
- Kernel version for generated Makefile tools: `__KERNEL_VERSION__`

External dependencies are present in `configs/config.yaml`, but DB, cache, object storage, and DTM are disabled by default so the service starts without local Postgres, Redis, Minio, or DTM.

## Boot order

```text
configx load
  -> logx.New and install via kernel.LogxLogger
  -> metricsx manager creation
  -> optional dtmx.New
  -> dbx/cachex/objectstorex resource initialization with shared logger/metrics
  -> HTTP/gRPC server construction with access log + metrics
  -> kernel.New(..., kernel.Metrics(...), kernel.DTM(...))
```

## Layout

```text
api/                 Protobuf APIs and generated HTTP/gRPC bindings
cmd/server/          Application entrypoint, renamed to cmd/<service> by kernel new
configs/             Local config with Kernel module defaults
internal/conf/        Config DTOs scanned by configx
internal/server/      Kernel HTTP and gRPC server construction
internal/service/     Transport-facing Todo service
internal/biz/         Use cases, domain contracts, errorx errors
internal/data/        Repositories and Kernel resource initialization
```

## Run

```bash
go run ./cmd/server -conf ./configs/config.yaml
```

Default HTTP endpoints:

- `GET /healthz`
- `GET /readyz`
- `POST /v1/todos/create`
- `GET /v1/todos/{id}`
- `GET /v1/todos/list`
- `PUT /v1/todos/update`
- `DELETE /v1/todos/{id}`
- `GET /v1/todos/watch`
- `POST /v1/todos/sync`

Default transport ports:

- HTTP: `0.0.0.0:8000`
- gRPC: `0.0.0.0:9000`
- Metrics admin server: `127.0.0.1:9090/metrics` when `metrics.enabled=true`

## Generate

```bash
make tools
make api
make proto-check
```

## Verify

```bash
make verify
```
