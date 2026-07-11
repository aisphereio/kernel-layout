# Kernel Modules

This section summarizes the Kernel framework modules used in the layout. Each module follows a consistent pattern:

1. **Config definition**: in `configs/config.yaml` and `internal/conf/conf.go`
2. **Resource initialization**: in `internal/data/data.go` `NewResources()`
3. **Service assembly**: in `internal/server/` middleware injection
4. **Business usage**: in `internal/service/`, `internal/biz/`, `internal/data/`

## Module Reference

| Module | Package | Purpose | Config Section |
|--------|---------|---------|----------------|
| [Config](config.md) | `configx` | YAML file + env var overlay | `service`, `server` |
| [Log](log.md) | `logx` | Structured, leveled logging with redaction | `log` |
| [Metrics](metrics.md) | `metricsx` | Prometheus counters, histograms, gauges | `metrics` |
| [Error](error.md) | `errorx` | Structured error codes with HTTP/gRPC mapping | (code-level) |
| [AuthN](authn.md) | `authn` | Authentication principal model | `security.authn` |
| [AuthZ / Access](authz.md) | `authz` / `accessx` | Authorization provider + access guard | `security.authz`, `security.access` |
| [Audit](audit.md) | `auditx` | Audit event recording | `audit` |
| [DTM](dtm.md) | `dtmx` | Distributed transaction (Saga/TCC) | `dtm` |

## Detailed Guides

Detailed usage guides with production code examples are maintained in `docs/modules/`:

- [Config Guide](../docs/modules/config.md)
- [Log Guide](../docs/modules/log.md)
- [Metrics Guide](../docs/modules/metrics.md)
- [Error Guide](../docs/modules/error.md)
- [AuthN Guide](../docs/modules/authn.md)
- [AuthZ Guide](../docs/modules/authz.md)
- [DTM Guide](../docs/modules/dtm.md)
- [Audit Guide](../docs/modules/audit.md)

## Common Pattern

```go
// 1. Config struct in internal/conf/conf.go
type MyModuleConfig struct { ... }

// 2. Init in internal/data/data.go NewResources()
resource, err := mymodule.New(cfg, logger, metrics)

// 3. Use in biz/service/data layers
resource.DoSomething(ctx, args)
```