# Config Module

## Overview

Kernel uses `configx` for configuration loading. Supports YAML files with environment variable overlay.

## Bootstrap Structure

The top-level config DTO is `conf.Bootstrap` in `internal/conf/conf.go`:

```go
type Bootstrap struct {
    Service  ServiceConfig
    Server   ServerConfig
    Log      logx.Config
    Data     DataConfig
    Security SecurityConfig
    Audit    AuditConfig
    Metrics  MetricsConfig
    DTM      dtmx.Config
}
```

## Loading

```go
// cmd/server/main.go
cfg := configx.New(configx.WithSource(file.NewSource(flagconf)))
cfg.Load()
var bc conf.Bootstrap
cfg.Scan(&bc)
```

## Key Config Sections

| Section | File | Purpose |
|---------|------|---------|
| `service` | `configs/config.yaml` | Name, version, env |
| `server.http` | `configs/config.yaml` | HTTP addr, timeout, CORS |
| `server.grpc` | `configs/config.yaml` | gRPC addr, timeout |
| `log` | `configs/config.yaml` | Log level, format, redaction, access log |
| `data.database` | `configs/config.yaml` | DB driver, DSN, pool settings |
| `data.cache` | `configs/config.yaml` | Cache driver, Redis addrs |
| `data.object_store` | `configs/config.yaml` | Object store driver, MinIO endpoint |
| `security.authn` | `configs/config.yaml` | Authn mode, provider, OIDC |
| `security.authz` | `configs/config.yaml` | Authz provider, IAM gRPC endpoint |
| `audit` | `configs/config.yaml` | Audit store backend |
| `metrics` | `configs/config.yaml` | Prometheus addr, path, runtime |
| `dtm` | `configs/config.yaml` | DTM server, protocol, timeout |

## External Dependencies

All external dependencies (DB, Cache, ObjectStore, Authn, Authz, DTM) are **disabled by default** so the service starts without local Postgres, Redis, MinIO, Casdoor, SpiceDB, or DTM.

## Source Map

| File | Role |
|------|------|
| `internal/conf/conf.go` | Config DTO definitions |
| `configs/config.yaml` | Default config values |
| `cmd/server/main.go` | Config loading and scanning |