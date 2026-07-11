# AuthZ / Access Module

## Overview

Kernel provides two layers of authorization:

- **`authz`**: Authorization provider contract (`Authorizer` interface) — answers "can this principal access this resource?"
- **`accessx`**: Runtime access control guard — orchestrates authn + authz + audit into a unified middleware

## Architecture

```text
Request → Middleware Chain
  ├── authn: extract Principal
  ├── accessx.Guard:
  │   ├── Check exposure level (PUBLIC / AUTHENTICATED / AUTHORIZED / INTERNAL / SYSTEM)
  │   ├── Run authz check (from proto-declared policy)
  │   └── Record audit event
  └── Business handler
```

## Proto-Declared Policies

Access policies are declared in proto RPC annotations:

```proto
rpc CreateTodo(CreateTodoRequest) returns (Todo) {
  option (aisphere.access.v1.policy) = {
    exposure: AUTHORIZED
    authz: { action: "create" resource: "todo:collection" audience: "todo-service" mode: CHECK_ONLY }
    audit: { enabled: true event: "todo.create" risk: "low" }
    rate_limit: { enabled: true key: "principal" qps: 100 burst: 200 backend: MEMORY }
  };
}
```

### Exposure Levels

| Level | Description |
|-------|-------------|
| `PUBLIC` | No authn or authz required |
| `AUTHENTICATED` | Valid Principal required, no resource-level authz |
| `AUTHORIZED` | Valid Principal + resource-level authz check |
| `INTERNAL` | Internal service only |
| `SYSTEM` | System-level operations |

## Access Resolver

The access resolver bridges proto-declared policies with runtime middleware:

```go
// internal/server/access.go
func todoAccessResolver(ctx context.Context, operation string, req any) (accessx.Check, bool, error) {
    check, ok, err := v1.TodoServiceAccessResolver(ctx, operation, req)
    if err != nil || ok {
        return check, ok, err
    }
    return accessx.Check{}, false, nil
}
```

## Authz Providers

| Provider | Description | Config |
|----------|-------------|--------|
| `iam_grpc` (default) | IAM service gRPC client via `aisphere-iam` | `security.authz.iam_grpc` |
| `dev_allow_all` | Skip actual authz (development only) | `security.authz.dev_allow_all: true` |
| `deny_all` | Default when authz disabled | (no config) |

## Relationship Projection

Business services are responsible for projecting business events into SpiceDB relationships:

- **Create resource**: write owner tuple, e.g. `skill:{name}#owner@user:{uid}`
- **Share/authorize**: write allowed relations (`viewer`, `editor`), never transfer `owner`
- **History repair**: backfill from durable source (e.g., PostgreSQL `owner_id`)
- **Write API**: use Kernel `authz.Service.WriteRelationships` with TOUCH semantics

## Source Map

| File | Role |
|------|------|
| `internal/server/access.go` | Middleware assembly, access resolver |
| `internal/data/data.go` | Authz provider initialization |
| `internal/conf/conf.go` | Authz config DTOs |
| `api/aisphere/access/v1/access.proto` | Access policy proto |
| `api/aisphere/options/v1/authz.pb.go` | Authz mode enum |
| `docs/modules/authz.md` | Detailed usage guide |