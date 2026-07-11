# Testing

## Unit Tests

Standard Go tests with `go test ./...`. The CI pipeline runs tests on every push/PR to `main`.

## Full-Flow Smoke Test

The layout includes a standalone smoke test binary at `cmd/fullflow-smoke/main.go`. It exercises the full Kernel middleware stack in-process:

### What It Tests

1. **HTTP server** — creates a Kernel HTTP server with full middleware chain
2. **gRPC server** — creates a Kernel gRPC server with full middleware chain
3. **Authn middleware** — injects a test Principal into context
4. **Access guard** — runs authz checks via `accessx.Guard`
5. **Rate limiting** — uses a demo count-based limiter
6. **Circuit breaker** — opens after configurable failures
7. **Admission webhook** — mutates requests (e.g., default title)
8. **Audit recording** — verifies audit events are recorded
9. **Todo CRUD** — creates, reads, updates, lists, and deletes todos
10. **Streaming** — tests WatchTodos (SSE) and SyncTodos (bidi)

### Key Patterns

- Uses in-memory stores and demo implementations (no external dependencies)
- Tests middleware chain ordering: `requestinfo → authn → access → rate limit → circuit breaker → admission → handler`
- Verifies audit records are populated after operations

## Pre-Commit Checklist

From `AGENTS.md`:

```powershell
make tools-local KERNEL_LOCAL=../kernel
make api
make proto-check
make test
make build
```

## Source Map

| File | Role |
|------|------|
| `cmd/fullflow-smoke/main.go` | Full-flow smoke test |
| `.github/workflows/go.yml` | CI test execution |