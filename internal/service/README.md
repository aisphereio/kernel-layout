# Service — 传输层 Handler

## 职责

- 实现 proto 生成的服务接口
- 从 Context 提取 `authn.Principal`
- 调用 `biz` 层的 Use Case
- 处理请求/响应转换
- 记录审计日志

## 典型结构

```go
// internal/service/xxx.go
package service

import (
    "context"
    v1 "github.com/aisphereio/kernel-layout/api/xxx/v1"
    "github.com/aisphereio/kernel-layout/internal/biz"
    "github.com/aisphereio/kernel/authn"
)

type XXXService struct {
    v1.UnimplementedXXXServiceServer
    uc *biz.XXXUsecase
}

func NewXXXService(uc *biz.XXXUsecase) *XXXService {
    return &XXXService{uc: uc}
}

func (s *XXXService) GetXXX(ctx context.Context, req *v1.GetXXXRequest) (*v1.XXX, error) {
    // 1. 获取 Principal
    principal, ok := authn.PrincipalFromContext(ctx)
    if !ok {
        return nil, authn.ErrMissingCredential("principal is required")
    }

    // 2. 调用业务层
    xxx, err := s.uc.GetXXX(ctx, req.GetId())
    if err != nil {
        return nil, err
    }

    // 3. 返回响应
    return convertXXX(xxx), nil
}
```

## 详细文档

- [authn 使用指南](../docs/modules/authn.md)
- [errorx 使用指南](../docs/modules/error.md)
- [auditx 使用指南](../docs/modules/audit.md)