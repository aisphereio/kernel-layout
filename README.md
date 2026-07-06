# Aisphere Kernel Layout

Standalone full-feature service layout for `github.com/aisphereio/kernel/cmd/kernel`.

Default mode is **full**: config, logging, metrics, DB/cache/object storage wiring, audit, DTM, HTTP/gRPC transports, proto-first Todo CRUD, governance code generation examples, and deploy manifest generation.

Use MVP when you want the smallest runnable service skeleton:

```powershell
kernel new skill-service --mvp
```

Use feature disable when you want to remove optional layout parts:

```powershell
kernel new skill-service --disable iam
kernel new skill-service --disable gateway,dtmx
```

## Use

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

## Included defaults

- Features: `__KERNEL_FEATURES__`
- Disabled features: `__KERNEL_DISABLED_FEATURES__`
- Profile: `__KERNEL_PROFILE__`
- DB: `dbx` with `__KERNEL_DB_DRIVER__`
- Cache: `cachex` with `__KERNEL_CACHE_DRIVER__`
- Object storage: `objectstorex` with `__KERNEL_OBJECTSTORE_DRIVER__`
- Authn: `__KERNEL_AUTHN_PROVIDER__`
- Authz: `__KERNEL_AUTHZ_PROVIDER__`
- Audit: `auditx` memory recorder by default
- Logging: `logx` console output for local development
- Metrics: shared `metricsx.Manager`, optional admin `/metrics` server
- DTM: optional `dtmx.Manager`
- Config: `configx` file source
- Transports: Kernel HTTP and gRPC servers with access log and metrics hooks
- API example: protobuf-first Todo CRUD with HTTP binding and optional governance annotations
- Deploy routes: Gateway API `HTTPRoute` manifests generated under `deploy/generated/gateway`
- Kernel version for generated Makefile tools: `__KERNEL_VERSION__`

External dependencies are present in `configs/config.yaml`, but DB, cache, object storage, authn, authz, and DTM are disabled by default so the service starts without local Postgres, Redis, Minio, Casdoor, SpiceDB, or DTM.

## Layout

```text
api/                 Protobuf APIs and generated HTTP/gRPC bindings
cmd/server/          Application entrypoint, renamed to cmd/<service> by kernel new
configs/             Local config with Kernel module defaults
deploy/generated/    Generated Gateway API HTTPRoute manifests split by exposure
internal/conf/        Config DTOs scanned by configx
internal/server/      Kernel HTTP and gRPC server construction
internal/service/     Transport-facing Todo service
internal/biz/         Use cases, domain contracts, errorx errors
internal/data/        Repositories and Kernel resource initialization
.kernel/              Layout profile/feature overlays consumed by kernel new
```

## Deploy route generation

`make deploy` runs `buf.gen.deploy.yaml`, which calls `protoc-gen-go-deploy` and writes Kubernetes Gateway API route manifests from protobuf annotations:

```text
PUBLIC                         -> deploy/generated/gateway/public/
AUTHENTICATED / AUTHORIZED     -> deploy/generated/gateway/authenticated/
INTERNAL / SYSTEM              -> deploy/generated/gateway/internal/
```

The generator reads both `google.api.http` and `aisphere.access.v1.policy`, so the generated route contains the HTTP method/path, upstream gRPC operation, exposure level, edge authn mode, and authz action/resource headers. This keeps route publication driven by proto contract rather than hand-written YAML.

Typical workflow:

```bash
make tools
make api
make deploy
make proto-check
```

Use local Kernel generator changes with:

```powershell
make tools-local KERNEL_LOCAL=../kernel
make api
make deploy
make proto-check
make test
```

The layout must keep generated services on the Kernel path: proto contract -> generated request info/access/gateway/deploy metadata -> HTTP/gRPC middleware -> business service.

## Local Kernel generator workflow

When the Kernel generator is being changed together with this layout, install tools from the local Kernel checkout instead of a released module version:

```powershell
make tools-local KERNEL_LOCAL=../kernel
make api
make deploy
make proto-check
make test
```

## Generate

```bash
make tools
make api
make deploy
make proto-check
```

## Verify

```bash
make verify
```
