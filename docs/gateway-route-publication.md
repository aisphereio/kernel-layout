# Gateway Route Publication Policy

`google.api.http` means an RPC has an HTTP binding. It does not automatically mean the RPC should be published through the public Envoy Gateway.

A route is generated into `GatewayManifest` only when the RPC also declares `aisphere.access.v1.policy`.

Use `gateway.publish: DISABLED` for direct-service HTTP routes that must never be published to Envoy Gateway.

```proto
option (aisphere.access.v1.policy) = {
  exposure: INTERNAL
  gateway: { publish: DISABLED tags: "dtm" }
};
```

Use profiles for routes that can be published to selected gateway planes:

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

Default production registration should use:

```go
serverx.RegisterServiceGatewayRoutesWithFilter(ctx, registry,
  gatewayx.PublicRouteFilter(),
  myv1.MyServiceKernelModule(),
)
```

Public Envoy Gateway should include product APIs and exclude internal/system/debug/ops routes. Internal service APIs should use a separate internal gateway profile or direct service discovery.