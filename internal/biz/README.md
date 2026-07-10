# Biz — 业务逻辑层

## 职责

- 定义领域模型（`Todo`、`Order` 等）
- 定义 Repository 接口（`TodoRepo`、`OrderRepo` 等）
- 实现 Use Case（`TodoUsecase`、`OrderUsecase` 等）
- 定义业务错误（`errorx`）
- 使用 `logx` 记录业务日志
- 使用 `dtmx` 编排分布式事务

## 典型结构

```go
// internal/biz/xxx.go
package biz

import (
    "context"
    "github.com/aisphereio/kernel/errorx"
)

// 1. 定义业务错误
var (
    ErrXXXNotFound = errorx.NotFound(errorx.Code("XXX_NOT_FOUND"), "xxx not found")
    ErrXXXInvalid  = errorx.BadRequest(errorx.Code("XXX_INVALID"), "invalid xxx")
)

// 2. 定义领域模型
type XXX struct {
    ID   int64
    Name string
}

// 3. 定义 Repository 接口
type XXXRepo interface {
    FindByID(context.Context, int64) (*XXX, error)
    Save(context.Context, *XXX) (*XXX, error)
}

// 4. 实现 Use Case
type XXXUsecase struct {
    repo XXXRepo
}

func NewXXXUsecase(repo XXXRepo) *XXXUsecase {
    return &XXXUsecase{repo: repo}
}

func (uc *XXXUsecase) GetXXX(ctx context.Context, id int64) (*XXX, error) {
    return uc.repo.FindByID(ctx, id)
}
```

## 详细文档

- [errorx 使用指南](../docs/modules/error.md)
- [logx 使用指南](../docs/modules/log.md)
- [dtmx 使用指南](../docs/modules/dtm.md)