# AuthN 认证模块

## 概述

Kernel 的 `authn` 包提供认证主体（Principal）模型和认证器接口。后端服务默认使用 `gateway_trusted` 模式，信任 Envoy Gateway 注入的 `x-aisphere-*` headers，自动恢复为 `authn.Principal` 并注入 Context。

## 配置

```yaml
# configs/config.yaml
security:
  authn:
    enabled: true
    mode: gateway_trusted    # 默认模式：信任 Gateway 注入的 headers
    provider: casdoor
    cache_ttl_ns: 300000000000  # 5 分钟
  internal_call:
    enabled: true
    header: X-Aisphere-Internal-Token
    token: "${GATEWAY_TO_APP_INTERNAL_TOKEN}"
```

## Principal 结构体

```go
// kernel/authn/principal.go
type Principal struct {
    SubjectID      string            // 用户唯一标识
    SubjectType    string            // 主体类型（user / service / robot）
    Provider       string            // 认证提供商（casdoor / oidc）
    ExternalID     string            // 外部 ID
    Issuer         string            // JWT 签发者
    Audience       string            // JWT 受众
    TenantID       string            // 租户 ID
    OrgID          string            // 组织 ID
    AppID          string            // 应用 ID
    ProjectID      string            // 项目 ID
    Username       string            // 用户名
    Name           string            // 显示名称
    Email          string            // 邮箱
    Phone          string            // 手机号
    Roles          []string          // 角色列表
    Groups         []string          // 组列表
    Scopes         []string          // 权限范围
    AuthMethod     string            // 认证方法（gateway / jwt / internal）
    Attributes     map[string]string // 扩展属性
    IssuedAt       time.Time         // 签发时间
    ExpiresAt      time.Time         // 过期时间
}
```

## 使用方式

### 1. 从 Context 获取 Principal（最常用）

```go
// internal/service/xxx.go
func (s *Service) GetProfile(ctx context.Context, req *pb.GetProfileReq) (*pb.GetProfileResp, error) {
    principal, ok := authn.PrincipalFromContext(ctx)
    if !ok || !principal.IsAuthenticated() {
        return nil, authn.ErrMissingCredential("gateway principal is required")
    }

    // 使用 Principal 的字段
    userID := principal.SubjectID
    orgID := principal.OrgID
    email := principal.Email
    username := principal.Username

    // ... 业务逻辑 ...
}
```

### 2. 将 Principal 传入下层

```go
// internal/biz/xxx.go
func (b *Biz) ListMyResources(ctx context.Context, page, size int) ([]*Resource, error) {
    principal, ok := authn.PrincipalFromContext(ctx)
    if !ok || !principal.IsAuthenticated() {
        return nil, authn.ErrMissingCredential("principal is required")
    }

    // 使用 SubjectID 做数据隔离
    return b.repo.ListByOwner(ctx, principal.SubjectID, page, size)
}
```

### 3. 提取 SubjectType 和 SubjectID

```go
// 封装为工具函数
func currentPrincipalSubject(ctx context.Context) (string, string, error) {
    principal, ok := authn.PrincipalFromContext(ctx)
    if !ok || !principal.IsAuthenticated() {
        return "", "", authn.ErrMissingCredential("kernel principal is required")
    }
    subjectType := strings.TrimSpace(principal.SubjectType)
    if subjectType == "" {
        subjectType = authn.SubjectTypeUser
    }
    return subjectType, strings.TrimSpace(principal.SubjectID), nil
}
```

### 4. 在 HTTP Handler 中使用

```go
// internal/server/xxx.go
func runWithGatewayPrincipal(c khttp.Context, fn func(context.Context, authn.Principal) (any, error)) (any, error) {
    return c.Middleware(func(ctx context.Context, _ any) (any, error) {
        principal, ok := authn.PrincipalFromContext(ctx)
        if !ok || !principal.IsAuthenticated() {
            return nil, authn.ErrMissingCredential("gateway principal is required")
        }
        return fn(ctx, principal.Normalize())
    })(c, nil)
}
```

## 真实代码示例（脱敏自 aisphere-iam）

### 示例 1：GetMe 接口

```go
// internal/service/iam.go
func (s *IAMAuthService) GetMe(ctx context.Context, req *v1.GetMeRequest) (*v1.GetMeReply, error) {
    principal, ok := authn.PrincipalFromContext(ctx)
    if !ok || !principal.IsAuthenticated() {
        return nil, authn.ErrMissingCredential("gateway principal is required")
    }
    return &v1.GetMeReply{Principal: principalToProto(principal)}, nil
}
```

### 示例 2：权限检查前获取 Principal

```go
// internal/service/iam.go
func (s *IAMDirectoryService) requireZonePermission(ctx context.Context, orgID string, permission string) error {
    principal, ok := authn.PrincipalFromContext(ctx)
    if !ok || !principal.IsAuthenticated() {
        return authn.ErrMissingCredential("gateway principal is required")
    }

    // 使用 principal 的字段做 SpiceDB 权限检查
    subject := &authz.Subject{
        Type: principal.SubjectType,
        ID:   principal.SubjectID,
    }
    // ... 调用 resources.Authz.Check(...)
}
```

### 示例 3：控制面操作强制使用 Context Principal

```go
// internal/service/control_plane.go
func (s *ControlPlaneService) CreateOrganization(ctx context.Context, req *v1.CreateOrganizationRequest) (*v1.Organization, error) {
    // 强制从 Context 获取 Actor，不接受请求体中的 owner
    subjectType, subjectID, err := currentPrincipalSubject(ctx)
    if err != nil {
        return nil, err
    }

    // 使用 subjectType/subjectID 作为创建者
    org, err := s.biz.Create(ctx, req, subjectType, subjectID)
    if err != nil {
        return nil, err
    }
    return org, nil
}
```

## 认证模式

| 模式 | 配置值 | 说明 | 适用场景 |
|------|--------|------|----------|
| Gateway 信任模式 | `gateway_trusted` | 信任 Envoy Gateway 注入的 headers | **所有后端服务（默认推荐）** |
| 禁用 | `disabled` | 不进行认证 | 仅用于开发调试 |

## 关键要点

1. **`authn.PrincipalFromContext(ctx)`**：从 context 中提取 Principal，返回 `(Principal, bool)`，`bool` 表示是否存在
2. **`principal.IsAuthenticated()`**：检查 Principal 是否已认证（SubjectID 非空）
3. **`principal.Normalize()`**：返回标准化后的 Principal，填充默认值
4. **`authn.ErrMissingCredential(msg)`**：当 Principal 不存在时返回的错误
5. **SubjectID 优先级**：`x-aisphere-principal` > `x-aisphere-user-id` > `x-aisphere-external-id` > `x-aisphere-external-sub`
6. **禁止直接解析 Header**：业务 handler 不应直接读取 `x-aisphere-*` headers，应通过 `authn.PrincipalFromContext(ctx)` 获取