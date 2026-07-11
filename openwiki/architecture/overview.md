# Architecture Overview

## System Architecture

The Kernel Layout produces services that follow a **gateway-proxy → backend** deployment model with Envoy Gateway as the authentication boundary.

```text
Client (Browser / Mobile / CLI)
  │
  ▼
Envoy Gateway
  ├── OIDC login (browser) or JWT Bearer (API)
  ├── Verifies Casdoor JWT via JWKS
  ├── Strips spoofable headers
  ├── Injects trusted X-Aisphere-* Principal headers
  ├── Injects X-Aisphere-Internal-Token
  └── Forwards to backend service
        │
        ▼
  Backend Service (Kernel HTTP/gRPC)
    ├── Middleware chain: requestinfo → authn → access
    ├── authn: validates internal token, restores Principal from headers
    ├── access: authn + authz + audit guard
    └── Business logic (service → biz → data)
```

## Internal Architecture

The service follows a strict **Service / Biz / Data** layering:

```text
cmd/server/main.go
  │
  ├── configx.Load() → conf.Bootstrap
  ├── logx.New() → Logger
  ├── metricsx.NewPrometheusManager() → Metrics
  ├── dtmx.New() → DTM Manager
  ├── data.NewResources() → Resources (DB, Cache, ObjectStore, Authn, Authz, Audit, Access)
  ├── data.NewData() → Data
  ├── data.NewTodoRepo() → biz.TodoRepo
  ├── biz.NewTodoUsecase() → UseCase
  ├── service.NewTodoService() → Transport Handler
  ├── server.NewHTTPServer() → Kernel HTTP Server
  ├── server.NewGRPCServer() → Kernel gRPC Server
  └── kernel.New() → App (lifecycle management)
```

## Layer Responsibilities

### Service Layer (`internal/service/`)
- Implements proto-generated service interfaces
- Extracts `authn.Principal` from context
- Calls biz use cases
- Handles request/response DTO conversion
- **No business rules, no direct data access**

### Biz Layer (`internal/biz/`)
- Defines domain models (Todo, etc.)
- Defines Repository interfaces
- Implements use cases with business rules
- Defines business errors via `errorx`
- Uses `logx` for business logging, `dtmx` for distributed transactions
- **No transport concerns, no persistence details**

### Data Layer (`internal/data/`)
- Initializes all Kernel resources (DB, Cache, ObjectStore, Authn, Authz, Audit, DTM)
- Implements Repository interfaces
- Manages resource lifecycle (create, close)
- Injects Logger and Metrics into sub-modules
- **Provider details only here** — no Casdoor/SpiceDB/MinIO SDK in biz/service

## Middleware Chain

The HTTP and gRPC servers share a middleware chain built by `serverx.ServerMiddlewareFromProviders`:

1. **RequestInfo**: extracts operation name, request ID, route pattern
2. **AuthN**: validates internal token, restores `authn.Principal` from trusted headers
3. **Access**: runs `accessx.Guard` — checks exposure level, runs authz check, records audit
4. **Rate Limit**: per-operation rate limiting (proto-declared)
5. **Circuit Breaker**: per-operation circuit breaker (proto-declared)

## AuthN Architecture

See [AuthN Architecture](authn.md) for details.

## Gateway Route Publication

See [Gateway Route Publication](gateway.md) for the policy on how proto annotations drive deploy manifest generation.

## Source Map

| File | Role |
|------|------|
| `cmd/server/main.go` | Bootstrap, wiring, lifecycle |
| `internal/server/http.go` | HTTP server construction with middleware |
| `internal/server/grpc.go` | gRPC server construction with middleware |
| `internal/server/access.go` | Middleware assembly, security runtime, access resolver |
| `internal/data/data.go` | Resource initialization (DB, Cache, ObjectStore, Authn, Authz, Audit, Access) |
| `internal/conf/conf.go` | Config DTOs |
| `cmd/fullflow-smoke/main.go` | Integration smoke test |