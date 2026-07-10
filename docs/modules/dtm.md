# DTM 分布式事务模块

## 概述

Kernel 使用 `dtmx` 包提供分布式事务能力，支持 Saga 模式（推荐）和 TCC 模式。DTM 是一个外部服务，通过 HTTP 协议与业务服务通信。

## 配置

```yaml
# configs/config.yaml
dtm:
  enabled: false
  driver: dtm
  protocol: http
  server: http://127.0.0.1:36789/api/dtmsvr
  service_base_url: http://127.0.0.1:8000
  branch_prefix: /internal/dtm
  branch_secret: change-me-only-for-local-dev
  wait_result: true
  timeout_ns: 10000000000
  metrics_enabled: true
```

## 使用方式

### 1. 创建 DTM Manager

```go
// cmd/server/main.go
func newDTMManager(bc conf.Bootstrap, logger logx.Logger, metrics metricsx.Manager) (dtmx.Manager, error) {
    cfg := bc.DTM
    cfg.Logger = logger.Named("dtmx")
    cfg.Metrics = metrics
    cfg.MetricsEnabled = cfg.MetricsEnabled && bc.Metrics.Enabled
    return dtmx.New(cfg)
}
```

### 2. 在 Resources 中注入 DTM

```go
// internal/data/data.go
type ResourceOptions struct {
    Logger  logx.Logger
    Metrics metricsx.Manager
    DTM     dtmx.Manager
}

func NewResources(ctx context.Context, cfg conf.Bootstrap, opts ResourceOptions) (*Resources, func(), error) {
    r := &Resources{
        DTM: dtmx.FromContextOr(ctx, opts.DTM),
    }
    // ...
}
```

### 3. Saga 模式（推荐）

Saga 模式将分布式事务拆分为多个步骤，每个步骤有对应的补偿操作。任一失败时自动执行已成功步骤的补偿。

```go
// internal/biz/xxx.go
func (b *Biz) Transfer(ctx context.Context, from, to string, amount int64) error {
    // 检查 DTM 是否可用
    if b.dtm == nil || !b.dtm.Enabled() {
        return errors.New("dtm not available")
    }

    // 生成全局事务 ID
    gid, err := b.dtm.NewGID(ctx)
    if err != nil {
        return err
    }

    // 创建 Saga 事务
    saga := dtmx.NewSaga(gid, "transfer").
        AddStep(
            // 正向操作：扣减
            func(ctx context.Context) error {
                return b.accountRepo.Deduct(ctx, from, amount)
            },
            // 补偿操作：加回
            func(ctx context.Context) error {
                return b.accountRepo.Add(ctx, from, amount)
            },
        ).
        AddStep(
            // 正向操作：增加
            func(ctx context.Context) error {
                return b.accountRepo.Add(ctx, to, amount)
            },
            // 补偿操作：扣减
            func(ctx context.Context) error {
                return b.accountRepo.Deduct(ctx, to, amount)
            },
        )

    // 提交 Saga
    _, err = b.dtm.SubmitSaga(ctx, saga)
    return err
}
```

### 4. Saga 模式（HTTP 分支）

当 Saga 步骤需要调用 HTTP 端点时，使用 `AddHTTP`：

```go
// internal/biz/xxx.go
func (m *Manager) ProjectAuthz(ctx context.Context, eventID string, payload Payload) error {
    if m.dtm != nil && m.dtm.Enabled() {
        gid, err := m.dtm.NewGID(ctx)
        if err != nil {
            return err
        }

        // 创建 Saga，使用 HTTP 分支
        saga := dtmx.NewSaga(gid, "authz-projection").
            AddHTTP(
                "project-authz",
                m.dtm.BranchURL("myapp/projection/apply"),       // 正向 URL
                m.dtm.BranchURL("myapp/projection/compensate"), // 补偿 URL
                payload,
            )

        if _, err := m.dtm.SubmitSaga(ctx, saga); err != nil {
            return err
        }
        return nil
    }

    // 降级：DTM 不可用时直接执行
    _, err := m.ApplyBranch(ctx, payload)
    return err
}
```

### 5. 实现 HTTP 分支端点

```go
// internal/server/dtm.go
func registerDTMHandlers(srv *khttp.Server, biz *biz.MyBiz) {
    // 正向操作端点
    srv.HandleFunc("/internal/dtm/myapp/projection/apply", func(w http.ResponseWriter, r *http.Request) {
        var payload Payload
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        if err := biz.ApplyBranch(r.Context(), payload); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusOK)
    })

    // 补偿操作端点
    srv.HandleFunc("/internal/dtm/myapp/projection/compensate", func(w http.ResponseWriter, r *http.Request) {
        var payload Payload
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        if err := biz.CompensateBranch(r.Context(), payload); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusOK)
    })
}
```

## 真实代码示例（脱敏自 aisphere-iam）

### 示例 1：授权关系投影 Saga

```go
// internal/biz/projection/manager.go
func (m *Manager) DispatchProjection(ctx context.Context, eventID string, payload Payload) error {
    if m.dtm != nil && m.dtm.Enabled() {
        gid, err := m.dtm.NewGID(ctx)
        if err != nil {
            return err
        }

        saga := dtmx.NewSaga(gid, "authz-projection").
            AddHTTP(
                "project-authz",
                m.dtm.BranchURL("iam/projection/apply"),
                m.dtm.BranchURL("iam/projection/compensate"),
                payload,
            )

        if _, err := m.dtm.SubmitSaga(ctx, saga); err != nil {
            return err
        }
        return nil
    }

    // 降级：直接执行
    _, err := m.ApplyBranch(ctx, payload)
    return err
}
```

### 示例 2：身份投影分发器

```go
// internal/data/identity_mode.go
func (d *IdentityProjectionDispatcher) submit(ctx context.Context, eventID string, payload Payload) error {
    if d.dtm != nil && d.dtm.Enabled() {
        gid, err := d.dtm.NewGID(ctx)
        if err != nil {
            return err
        }

        saga := dtmx.NewSaga(gid, "identity-authz-projection").
            AddHTTP(
                "identity-authz",
                d.dtm.BranchURL("iam/identity-authz/apply"),
                d.dtm.BranchURL("iam/identity-authz/compensate"),
                payload,
            )

        _, err = d.dtm.SubmitSaga(ctx, saga)
        return err
    }

    // 降级：直接执行
    _, err := d.ApplyBranch(ctx, payload)
    return err
}
```

## 关键要点

1. **`dtmx.New(cfg)`**：创建 DTM Manager，注入 Logger 和 Metrics
2. **`dtmx.FromContextOr(ctx, opts.DTM)`**：从 context 或选项中获取 DTM Manager，支持上下文传递
3. **`dtm.NewGID(ctx)`**：生成全局事务 ID
4. **`dtmx.NewSaga(gid, topic)`**：创建 Saga 事务，`topic` 用于日志和追踪
5. **`saga.AddHTTP(name, applyURL, compensateURL, payload)`**：添加 HTTP 分支，包含正向和补偿 URL
6. **`dtm.SubmitSaga(ctx, saga)`**：提交 Saga 事务到 DTM 服务端
7. **`dtm.BranchURL(path)`**：生成完整的分支 URL（`service_base_url + branch_prefix + path`）
8. **降级策略**：当 DTM 不可用时，应提供降级方案（直接执行或报错）
9. **分支端点**：需要实现 HTTP 端点处理正向和补偿操作，返回 `200 OK` 表示成功