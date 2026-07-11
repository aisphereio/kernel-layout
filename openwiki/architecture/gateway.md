# Gateway Route Publication

## Policy

`google.api.http` means an RPC has an HTTP binding. It does **not** automatically mean the RPC should be published through the public Envoy Gateway.

A route is generated into `GatewayManifest` only when the RPC also declares `aisphere.access.v1.policy`.

## Proto Annotations

### Disable Gateway Publication

```proto
option (aisphere.access.v1.policy) = {
  exposure: INTERNAL
  gateway: { publish: DISABLED tags: "dtm" }
};
```

### Selective Gateway Profiles

```proto
option (aisphere.access.v1.policy) = {
  exposure: AUTHORIZED
  authz: {
    action: "update"
    resource: "my:resource:{id}"
    audience: "my-service"
    mode: CHECK_ONLY
  }
  gateway: {
    profiles: "public"
    profiles: "internal"
    tags: "business"
  }
};
```

## Generated Manifests

`make deploy` runs `buf.gen.deploy.yaml`, which calls `protoc-gen-go-deploy` and writes Kubernetes Gateway API route manifests from proto annotations:

```text
PUBLIC                         -> deploy/generated/gateway/public/
AUTHENTICATED / AUTHORIZED     -> deploy/generated/gateway/authenticated/
INTERNAL / SYSTEM              -> deploy/generated/gateway/internal/
```

The generator reads both `google.api.http` and `aisphere.access.v1.policy`, so the generated route contains:
- HTTP method/path
- Upstream gRPC operation
- Exposure level
- Edge authn mode
- Authz action/resource headers

## Registration

Default production registration:

```go
serverx.RegisterServiceGatewayRoutesWithFilter(ctx, registry,
  gatewayx.PublicRouteFilter(),
  myv1.MyServiceKernelModule(),
)
```

- **Public Envoy Gateway**: product APIs only, exclude internal/system/debug/ops routes
- **Internal service APIs**: separate internal gateway profile or direct service discovery

## Source Map

| File | Role |
|------|------|
| `docs/gateway-route-publication.md` | Full policy document |
| `buf.gen.deploy.yaml` | Deploy manifest generation config |
| `api/aisphere/access/v1/access.proto` | Access policy proto with Gateway options |