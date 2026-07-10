# Data — 数据访问层

## 职责

- 初始化 Kernel 资源（DB、Cache、ObjectStore、Authn、Authz、Audit、DTM）
- 实现 Repository 接口（`TodoRepo`、`XXXRepo` 等）
- 管理资源生命周期（创建、关闭）
- 注入 Logger 和 Metrics 到各子模块

## 典型结构

```go
// internal/data/xxx.go
package data

import (
    "context"
    "github.com/aisphereio/kernel-layout/internal/biz"
    "github.com/aisphereio/kernel/logx"
)

type xxxRepo struct {
    data   *Data
    logger logx.Logger
}

func NewXXXRepo(data *Data, logger logx.Logger) biz.XXXRepo {
    return &xxxRepo{data: data, logger: logger.Named("xxx.repo")}
}

func (r *xxxRepo) FindByID(ctx context.Context, id int64) (*biz.XXX, error) {
    r.logger.Infow("查询实体", "id", id)
    // ... 数据库查询 ...
}
```

## 资源初始化

`NewResources()` 负责初始化所有 Kernel 资源：

```go
func NewResources(ctx context.Context, cfg conf.Bootstrap, opts ResourceOptions) (*Resources, func(), error) {
    // 1. 初始化 Logger
    // 2. 初始化 Metrics
    // 3. 初始化 DB（如果启用）
    // 4. 初始化 Cache（如果启用）
    // 5. 初始化 ObjectStore（如果启用）
    // 6. 初始化 Authn（如果启用）
    // 7. 初始化 Authz（如果启用）
    // 8. 初始化 Access Guard
    // 9. 初始化 Audit
    // 10. 初始化 DTM
}
```

## 详细文档

- [config 使用指南](../docs/modules/config.md)
- [logx 使用指南](../docs/modules/log.md)
- [metricsx 使用指南](../docs/modules/metrics.md)
- [authn 使用指南](../docs/modules/authn.md)
- [authz 使用指南](../docs/modules/authz.md)
- [dtmx 使用指南](../docs/modules/dtm.md)
- [auditx 使用指南](../docs/modules/audit.md)