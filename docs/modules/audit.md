# Audit 审计模块

## 概述

Kernel 使用 `auditx` 包提供审计日志记录能力。审计日志记录关键操作的事件、操作者、目标、结果等信息，用于安全审计和合规要求。

## 配置

```yaml
# configs/config.yaml
audit:
  enabled: true
  store: memory    # 存储后端：memory（开发用）| 后续支持 db / kafka
```

## 使用方式

### 1. 初始化审计记录器

```go
// internal/data/data.go
func NewResources(ctx context.Context, cfg conf.Bootstrap, opts ResourceOptions) (*Resources, func(), error) {
    r := &Resources{
        Audit: auditx.NewMemoryStore(),  // 默认使用内存存储
    }
    if !cfg.Audit.Enabled {
        r.Audit = auditx.Noop()          // 禁用时使用 Noop
    }
    // ...
}
```

### 2. 注入到 accessx.Guard

```go
// internal/data/data.go
r.Access = accessx.New(r.Authn, r.Authz, r.Audit)
// 将审计记录器注入到访问控制守卫中，每次鉴权操作自动记录审计日志
```

### 3. 手动记录审计事件

```go
// internal/service/xxx.go
func (s *UserService) UpdateUserRole(ctx context.Context, req *pb.UpdateRoleReq) error {
    principal, ok := authn.PrincipalFromContext(ctx)
    if !ok {
        return authn.ErrMissingCredential("principal is required")
    }

    // 执行操作
    if err := s.biz.UpdateRole(ctx, req.UserId, req.Role); err != nil {
        return err
    }

    // 记录审计日志
    s.audit.Record(ctx, &auditx.Event{
        Action:    "user.role.update",
        Operator:  principal.SubjectID,
        Target:    req.UserId,
        Detail:    fmt.Sprintf("role changed to %s", req.Role),
        Result:    "success",
    })
    return nil
}
```

### 4. 记录敏感数据变更

```go
// internal/biz/xxx.go
func (b *UserBiz) UpdateEmail(ctx context.Context, userID, newEmail string) error {
    oldEmail, err := b.repo.GetEmail(ctx, userID)
    if err != nil {
        return err
    }

    if err := b.repo.UpdateEmail(ctx, userID, newEmail); err != nil {
        return err
    }

    // 记录旧值和新值
    b.audit.Record(ctx, &auditx.Event{
        Action:   "user.email.update",
        Actor:    userID,
        Target:   userID,
        OldValue: oldEmail,
        NewValue: newEmail,
        Result:   "success",
    })
    return nil
}
```

### 5. 批量操作审计

```go
// internal/service/xxx.go
func (s *AdminService) BatchDeleteUsers(ctx context.Context, userIDs []string) error {
    principal, ok := authn.PrincipalFromContext(ctx)
    if !ok {
        return nil, authn.ErrMissingCredential("unknown")
    }

    for _, uid := range userIDs {
        if err := s.biz.DeleteUser(ctx, uid); err != nil {
            // 记录失败
            s.audit.Record(ctx, &audit.Event{
                Action: "user.batch_delete",
                Actor:  principal.SubjectID,
                Target: uid,
                Result: "failed",
                Reason: err.Error(),
            })
            continue
        }
        // 记录成功
        s.audit.Record(ctx, &audit.Event{
            Action: "user.batch_delete",
            Actor:  principal.SubjectID,
            Target: uid,
            Result: "success",
        })
    }
    return nil
}
```

## 真实代码示例（脱敏自 aisphere-iam）

### 示例 1：初始化审计记录器

```go
// internal/data/data.go
type Resources struct {
    DB          dbx.DB
    Cache       cachex.Cache
    ObjectStore objectstorex.Client
    Audit       auditx.Recorder    // 审计记录器
    Authn       authn.Authenticator
    Authz       authz.Authorizer
    Access      accessx.Guard
    DTM         dtmx.Manager
    // ...
}

func NewResources(ctx context.Context, cfg conf.Bootstrap, opts ResourceOptions) (*Resources, func(), error) {
    r := &Resources{
        Audit: auditx.NewMemoryStore(),  // 默认使用内存存储
    }
    if !cfg.Audit.Enabled {
        r.Audit = auditx.Noop()          // 禁用时无操作
    }
    // ...

    // 将审计记录器注入 accessx.Guard
    r.Access = accessx.New(r.Authn, r.Authz, r.Audit)
    // ...
}
```

### 示例 2：通过 accessx 自动审计

当使用 `accessx.Guard` 时，每次鉴权操作会自动记录审计日志，无需手动调用：

```go
// 在中间件链中自动执行
// 1. authn 认证
// 2. authz 授权检查
// 3. audit 自动记录审计事件
// 4. 业务 handler 执行
```

## 关键要点

1. **`auditx.NewMemoryStore()`**：创建内存存储的审计记录器，适合开发环境
2. **`auditx.Noop()`**：当审计禁用时使用，所有操作无副作用
3. **`auditx.Recorder` 接口**：定义 `Record(ctx, *Event)` 方法，可扩展为其他存储后端
4. **注入到 accessx**：将 `auditx.Recorder` 注入到 `accessx.New()`，每次鉴权自动记录审计
5. **手动记录**：对于需要额外审计的业务操作，手动调用 `audit.Record()`
6. **Event 结构**：包含 `Action`（操作）、`Actor`（操作者）、`Target`（目标）、`Result`（结果）、`OldValue`/`NewValue`（变更前后值）等字段