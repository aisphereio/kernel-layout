# Todo Domain — Example CRUD

The Todo service is the **reference example** in the layout, demonstrating proto-first development with full governance annotations.

## Proto API

Defined in `api/todo/v1/todo.proto`:

| RPC | HTTP | Exposure | Authz Action | Resource | Audit Risk |
|-----|------|----------|-------------|----------|------------|
| `CreateTodo` | `POST /v1/todos/create` | AUTHORIZED | `create` | `todo:collection` | low |
| `GetTodo` | `GET /v1/todos/{id}` | AUTHORIZED | `read` | `todo:{id}` | low |
| `ListTodos` | `GET /v1/todos/list` | AUTHORIZED | `read` | `todo:collection` | low |
| `UpdateTodo` | `PUT /v1/todos/update` | AUTHORIZED | `update` | `todo:{todo.id}` | medium |
| `DeleteTodo` | `DELETE /v1/todos/{id}` | AUTHORIZED | `delete` | `todo:{id}` | high |
| `WatchTodos` | `GET /v1/todos/watch` (SSE) | AUTHORIZED | `watch` | `todo:collection` | low |
| `SyncTodos` | `POST /v1/todos/sync` (bidi) | AUTHORIZED | `sync` | `todo:collection` | medium |

## Error Reasons

Defined in `api/todo/v1/error_reason.proto`:

```proto
enum ErrorReason {
  TODO_UNSPECIFIED = 0;
  TODO_NOT_FOUND = 1;
  TODO_INVALID_ARGUMENT = 2;
}
```

## Layer Walkthrough

### Service Layer (`internal/service/`)

Implements the proto-generated `TodoServiceServer` interface. Handles DTO conversion and calls the use case.

### Biz Layer (`internal/biz/`)

Defines the `Todo` domain model and `TodoRepo` interface. The `TodoUsecase` implements business logic.

### Data Layer (`internal/data/`)

Implements `TodoRepo` with an in-memory store (for the template). In production, this would use PostgreSQL via `dbx`.

## Key Patterns Demonstrated

1. **Proto-first**: All RPCs declared in proto with HTTP bindings and access policies
2. **Governance annotations**: Every RPC has `aisphere.access.v1.policy` with exposure, authz, audit, and rate limit
3. **Streaming**: Server-side streaming (WatchTodos) and bidirectional streaming (SyncTodos)
4. **Error codes**: Proto-defined error reasons mapped to `errorx` errors
5. **Generated code**: `buf.gen.yaml` generates Go types, gRPC stubs, HTTP handlers, authz resolvers, gateway resolvers, and Kernel bindings

## Source Map

| File | Role |
|------|------|
| `api/todo/v1/todo.proto` | Proto API definition |
| `api/todo/v1/error_reason.proto` | Error reason enum |
| `api/todo/v1/todo.pb.go` | Generated message types |
| `api/todo/v1/todo_grpc.pb.go` | Generated gRPC stubs |
| `api/todo/v1/todo_http.pb.go` | Generated HTTP handlers |
| `api/todo/v1/todo_authz.pb.go` | Generated authz resolver |
| `api/todo/v1/todo_gateway.pb.go` | Generated gateway resolver |
| `api/todo/v1/todo_kernel.pb.go` | Generated Kernel bindings |
| `internal/service/` | Transport handler |
| `internal/biz/` | Business use case |
| `internal/data/` | Repository implementation |