# Data Layer

## Resource Initialization

The `data.NewResources()` function in `internal/data/data.go` is the central resource initialization point. It conditionally initializes each Kernel resource based on config:

```go
func NewResources(ctx context.Context, cfg conf.Bootstrap, opts ResourceOptions) (*Resources, func(), error) {
    // 1. Logger (from opts or default)
    // 2. Metrics (ensure non-nil)
    // 3. Audit: memory store (default) or noop (disabled)
    // 4. Authz: deny all (default), allow all (dev), or IAM gRPC
    // 5. DTM: from context or opts
    // 6. DB: if enabled, init via dbx
    // 7. Cache: if enabled, init via cachex
    // 8. ObjectStore: if enabled, init via objectstorex
    // 9. Authn: if enabled, init via authn provider
    // 10. Authz: if enabled and not dev_allow_all, init via authz provider
    // 11. Access: accessx.New(authn, authz, audit)
    // 12. Ping: verify DB and Cache connectivity
}
```

## Resources Struct

```go
type Resources struct {
    DB          dbx.DB
    Cache       cachex.Cache
    ObjectStore objectstorex.Client
    Audit       auditx.Recorder
    Authn       authn.Authenticator
    Authz       authz.Authorizer
    Access      accessx.Guard
    DTM         dtmx.Manager
}
```

## Provider Wiring

Provider details are isolated in the data layer:

| Resource | Default Provider | Config Section | Import |
|----------|-----------------|----------------|--------|
| DB | PostgreSQL via `dbx` | `data.database` | `github.com/aisphereio/kernel/dbx/postgres` |
| Cache | Redis via `cachex` | `data.cache` | `github.com/aisphereio/kernel/cachex/redis` |
| ObjectStore | MinIO via `objectstorex` | `data.object_store` | `github.com/aisphereio/kernel/objectstorex/minio` |
| Authn | Casdoor via `authn/casdoor` | `security.authn` | `github.com/aisphereio/kernel/authn/casdoor` |
| Authz | IAM gRPC via `aisphere-iam` | `security.authz` | `github.com/aisphereio/aisphere-iam/client/authzgrpc` |

## Key Rules

1. **Provider details only in data layer** — no Casdoor/SpiceDB/MinIO SDK in biz or service
2. **All external dependencies disabled by default** — service starts without Postgres, Redis, MinIO, Casdoor, SpiceDB, or DTM
3. **Idempotent bootstrap** — schema/bootstrap/relationship projection must be repeatable
4. **Resource cleanup** — `Resources.Close()` reverses initialization order

## Source Map

| File | Role |
|------|------|
| `internal/data/data.go` | Resource initialization and provider wiring |
| `internal/data/README.md` | Data layer usage guide |
| `internal/conf/conf.go` | Data config DTOs |
| `configs/config.yaml` | Default data config (all disabled) |