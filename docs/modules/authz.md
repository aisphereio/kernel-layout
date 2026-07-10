# AuthZ / Access 授权与访问控制模块

## 概述

Kernel 提供两层授权能力：

- **`authz`**：授权 provider contract，定义 `Authorizer` 接口，回答"能不能访问资源"
- **`accessx`**：运行时访问控制守卫，编排 authn + authz + audit，提供统一的 `accessx.Guard`

## 配置

```yaml
# configs/config.yaml
security:
  authz:
    enabled: false
    provider: spicedb
    dev_allow_all: true          # 开发阶段跳过实际授权检查
    spicedb:
      endpoint: 127.0.0.1:50051
      token: ""
      transport: grpc
      insecure: true
      timeout: 5000000000
      fully_consistent: true
      metrics_enabled: true
  access:
    public_operations: ["*"]     # 公开操作列表，跳过认证
    skip_operations: []          # 跳过操作列表
```

## 使用方式

### 1. 初始化 accessx.Guard

```go
// internal/data/data.go
r.Access = accessx.New(r.Authn, r.Authz, r.Audit)
// accessx.New() 将认证器、授权器和审计记录器组合成一个统一的访问控制守卫
```

### 2. 将 Guard 注入到中间件链

```go
// internal/server/access.go
func todoServerMiddlewares(resources *data.Resources, cfg conf.SecurityConfig) []middleware.Middleware {
    securityRuntime := mustSecurityRuntime(cfg)
    return serverx.ServerMiddlewareFromProviders(context.Background(), serverx.RuntimeProviders{
        Security:    securityRuntime,
        AccessGuard: &resources.Access,
        RequestInfoResolver: v1.TodoServiceRequestInfoResolver,
        AccessResolver:      todoAccessResolver,
    })
}
```

### 3. 定义 AccessResolver

```go
// internal/server/access.go
func todoAccessResolver(ctx context.Context, operation string, req any) (accessx.Check, bool, error) {
    // 先尝试使用 proto 生成的 resolver
    check, ok, err := v1.TodoServiceAccessResolver(ctx, operation, req)
    if err != nil || ok {
        return check, ok, err
    }
    // 回退：无匹配规则
    return accessx.Check{}, false, nil
}
```

### 4. 定义 SkipPolicy

```go
// internal/server/access.go
func iamSkipPolicyResolver(catalog serverx.ServiceCatalog) accessmw.SkipPolicyResolver {
    return func(operation string) accessx.SkipPolicy {
        op := strings.TrimPrefix(operation, "/")
        switch op {
        case "healthz", "readyz", "metrics":
            return accessx.SkipAll          // 跳过认证和授权
        case "iam.v1.IAMAuthService/ExternalAuthorize":
            return accessx.SkipAll          // 公开接口
        }
        return accessx.SkipDefault          // 使用 proto 注解的策略
    }
}
```

### 5. Proto 注解定义访问策略

```protobuf
// api/todo/v1/todo.proto
rpc CreateTodo(CreateTodoRequest) returns (Todo) {
  option (google.api.http) = {
    post: "/v1/todos"
    body: "*"
  };
  option (aisphere.access.v1.policy) = {
    exposure: AUTHORIZED
    authz: {
      action: "create"
      resource: "app:todo:*"
      audience: "user"
      mode: CHECK_ONLY
    }
    audit: {
      event: "todo.created"
      risk: MEDIUM
    }
  };
}
```

### 6. 生成的 AccessResolver

proto 注解会自动生成 `AccessResolver`，将注解规则映射为 `accessx.Check`：

```go
// api/todo/v1/todo_authz.pb.go（自动生成）
func TodoServiceAccessResolver(ctx context.Context, operation string, req any) (accessx.Check, bool, error) {
    rule, ok := TodoServiceAuthzRules[operation]
    if !ok || rule.Action == "" {
        return accessx.Check{}, false, nil
    }
    resource, err := (authz.RuleResolver{}).ResolveResource(rule, req)
    if err != nil {
        return accessx.Check{}, true, err
    }
    check := accessx.Check{
        Permission:  rule.Action,
        Resource:    resource,
        AuditAction: rule.AuditEvent,
        Metadata:    map[string]any{"authz_rule": rule.FullMethod, "authz_mode": string(rule.Mode)},
    }
    return check, true, nil
}
```

## 真实代码示例（脱敏自 aisphere-iam）

### 示例 1：在 Service 中手动检查权限

```go
// internal/service/iam.go
func (s *IAMDirectoryService) requireZonePermission(ctx context.Context, orgID string, permission string) error {
    principal, ok := authn.PrincipalFromContext(ctx)
    if !ok || !principal.IsAuthenticated() {
        return authn.ErrMissingCredential("gateway principal is required")
    }

    // 构造 SpiceDB 检查请求
    subject := &authz.Subject{
        Type: principal.SubjectType,
        ID:   principal.SubjectID,
    }
    resource := &authz.Resource{
        Type: "zone",
        ID:   orgID,
    }

    // 执行授权检查
    ok, err := s.authz.Check(ctx, subject, permission, resource)
    if err != nil {
        return authz.ErrBackendFailed("authz check failed", err)
    }
    if !ok {
        return authz.ErrPermissionDenied("permission denied")
    }
    return nil
}
```

### 示例 2：在 HTTP Handler 中检查

```go
// internal/server/identity_membership.go
func requireIdentityGroupMembershipPermission(r *http.Request, resources *data.Resources, orgID string) error {
    principal, ok := authn.PrincipalFromContext(r.Context())
    if !ok || !principal.IsAuthenticated() {
        return authn.ErrMissingCredential("gateway principal is required")
    }

    // 使用 resources.Authz 做 SpiceDB 检查
    subject := &authz.Subject{
        Type: principal.SubjectType,
        ID:   principal.SubjectID,
    }
    allowed, err := resources.Authz.Check(r.Context(), subject, "manage", &authz.Resource{
        Type: "zone",
        ID:   orgID,
    })
    if err != nil {
        return err
    }
    if !allowed {
        return authz.ErrPermissionDenied("insufficient permissions")
    }
    return nil
}
```

### 示例 3：SkipPolicy 控制

```go
// internal/server/access.go
func iamSkipPolicyResolver(catalog serverx.ServiceCatalog) accessmw.SkipPolicyResolver {
    return func(operation string) accessx.SkipPolicy {
        op := strings.TrimPrefix(operation, "/")
        switch op {
        case "healthz", "readyz", "metrics":
            return accessx.SkipAll
        case "iam.v1.IAMAuthService/ExternalAuthorize":
            return accessx.SkipAll
        case "iam.v1.IAMAuthService/GetMe":
            return accessx.SkipAuthz  // 需要认证但不需要额外授权
        }
        return accessx.SkipDefault
    }
}
```

## Exposure 级别

| 级别 | 值 | 说明 | Gateway 行为 |
|------|-----|------|-------------|
| PUBLIC | 0 | 公开接口，无需认证 | 不挂 SecurityPolicy |
| AUTHENTICATED | 1 | 需要认证，无需额外授权 | 挂 OIDC/JWT |
| AUTHORIZED | 2 | 需要认证 + 授权检查 | 挂 OIDC/JWT |
| INTERNAL | 3 | 仅限内部服务间调用 | 不对外发布 |
| SYSTEM | 4 | 仅限系统级调用 | 不对外发布 |

## SkipPolicy 策略

| 策略 | 说明 |
|------|------|
| `accessx.SkipAll` | 跳过认证和授权（公开接口） |
| `accessx.SkipAuthn` | 跳过认证，但保留授权 |
| `accessx.SkipAuthz` | 跳过授权，但需要认证 |
| `accessx.SkipDefault` | 使用 proto 注解的默认策略 |

## 关键要点

1. **`accessx.New(authn, authz, audit)`**：将认证、授权、审计组合为一个 Guard，自动编排
2. **`accessx.Check`**：包含 `Permission`（操作）、`Resource`（资源）、`AuditAction`（审计事件）
3. **`dev_allow_all: true`**：开发阶段跳过实际授权检查，所有请求通过
4. **Proto 驱动**：访问策略在 proto 注解中声明，代码生成器自动生成 `AccessResolver`
5. **SkipPolicy**：通过 `SkipPolicyResolver` 控制哪些操作跳过认证/授权
6. **手动检查**：对于复杂授权逻辑，可以直接调用 `resources.Authz.Check()`