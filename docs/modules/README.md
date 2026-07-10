# 模块使用指南

本文档系列详细说明 Kernel 框架各模块在业务服务中的使用方式，包含从 `aisphere-iam` 等生产项目中脱敏的真实代码示例。

## 模块列表

| 模块 | 文档 | 核心包 | 说明 |
|------|------|--------|------|
| Config | [config.md](config.md) | `configx` | 配置加载与结构体定义 |
| Log | [log.md](log.md) | `logx` | 结构化日志 |
| Metrics | [metrics.md](metrics.md) | `metricsx` | Prometheus 指标 |
| Error | [error.md](error.md) | `errorx` | 结构化错误处理 |
| AuthN | [authn.md](authn.md) | `authn` | 认证主体 |
| AuthZ / Access | [authz.md](authz.md) | `authz` / `accessx` | 授权与访问控制 |
| DTM | [dtm.md](dtm.md) | `dtmx` | 分布式事务 |
| Audit | [audit.md](audit.md) | `auditx` | 审计日志 |

## 代码分层

```text
cmd/server/         入口：创建 Logger、Metrics、DTM、启动服务
internal/conf/      配置结构体：configx 扫描
internal/server/    HTTP/gRPC 服务端：中间件装配
internal/service/   传输层 Handler：authn.PrincipalFromContext、auditx
internal/biz/       业务逻辑层：errorx、logx、dtmx
internal/data/      数据访问层：logx、metricsx、资源初始化
```

## 通用模式

每个模块在 kernel-layout 中的使用遵循以下模式：

1. **配置定义**：在 `configs/config.yaml` 中声明配置，在 `internal/conf/conf.go` 中定义结构体
2. **资源初始化**：在 `internal/data/data.go` 的 `NewResources()` 中初始化
3. **服务装配**：在 `internal/server/` 中注入中间件
4. **业务使用**：在 `internal/service/`、`internal/biz/`、`internal/data/` 中调用