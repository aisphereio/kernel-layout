# Log 日志模块

## 概述

Kernel 使用 `logx` 包提供结构化日志能力，支持等级控制、结构化字段、上下文追踪、敏感信息脱敏。

## 配置

```yaml
# configs/config.yaml
log:
  service_name: app
  env: local
  version: dev
  level: info
  format: console
  output: stdout
  add_source: true
  redact:
    enabled: true
    keys:
      - password
      - secret
      - token
      - access_key
      - secret_key
    value: "***"
  access_log:
    enabled: true
    skip_paths:
      - /healthz
      - /readyz
      - /metrics
    slow_threshold: 1000000000  # 1s
    request_id_header: X-Request-ID
    route_header: X-Route-Pattern
    log_user_agent: true
```

## 使用方式

### 1. 启动时创建 Logger

```go
// cmd/server/main.go
logger, _, err := logx.New(bc.Log)
if err != nil {
    panic(err)
}
defer func() { _ = logger.Sync() }()
```

### 2. 创建子 Logger（带组件名称）

```go
// 在 data 层初始化时
logger := opts.Logger.Named("data.dbx")

// 在 server 层
logger := logger.Named("transport.http")
```

### 3. 从 Context 获取 Logger（带追踪信息）

```go
// internal/service/xxx.go
func (s *Service) GetProfile(ctx context.Context, req *pb.GetProfileReq) (*pb.GetProfileResp, error) {
    logger := logx.FromContext(ctx)
    logger.Info("获取用户信息",
        "user_id", req.GetUserId(),
    )
    // ... 业务逻辑 ...
}
```

### 4. 使用结构化字段

```go
// 使用 logx 提供的结构化字段函数
log.WithContext(ctx).Warn("读取 schema 失败，将尝试重新初始化",
    logx.Err(err),
    logx.String("schema_path", path),
    logx.Int("retry_count", retries),
)

// 记录成功
log.WithContext(ctx).Info("schema 已安装，跳过引导",
    logx.Int("size", len(schema.Text)),
)
```

### 5. 记录错误日志

```go
// 使用 logx.DefaultLogger() 记录全局错误
logx.DefaultLogger().Error("OAuth 令牌交换失败",
    logx.Err(err),
    logx.String("redirect_uri", req.RedirectURI),
    logx.String("code_prefix", safePrefix(req.Code, 10)),
)
```

### 6. 在 Repository 层使用注入的 Logger

```go
// internal/data/xxx.go
type xxxRepo struct {
    data   *Data
    logger logx.Logger
}

func NewXXXRepo(data *Data, logger logx.Logger) *xxxRepo {
    return &xxxRepo{data: data, logger: logger.Named("xxx.repo")}
}

func (r *xxxRepo) Save(ctx context.Context, entity *Entity) error {
    r.logger.Infow("保存实体",
        "entity_id", entity.ID,
        "entity_type", entity.Type,
    )
    // ... 数据库操作 ...
}
```

## 真实代码示例（脱敏自 aisphere-iam）

### 示例 1：带 Context 的结构化日志

```go
// 在 data 层引导过程中记录日志
func BootstrapSchema(ctx context.Context, cfg Config, resources *Resources, log logx.Logger) error {
    if log == nil {
        log = logx.Noop()
    }
    log = log.Named("schema.bootstrap")

    // 读取 schema 文件
    schema, err := os.ReadFile(path)
    if err != nil {
        log.WithContext(ctx).Warn("读取 schema 文件失败，将尝试引导",
            logx.Err(err),
            logx.String("schema_path", path),
        )
        // ... 引导逻辑 ...
    }

    log.WithContext(ctx).Info("schema 已安装，跳过引导",
        logx.Int("size", len(schema.Text)),
    )
    return nil
}
```

### 示例 2：记录操作结果

```go
// 记录引导结果
logger.Info("控制面管理员关系引导完成",
    logx.Int("written", result.Written),
)
```

### 示例 3：在 data 层注入 Logger

```go
// internal/data/data.go
type ResourceOptions struct {
    Logger  logx.Logger
    Metrics metricsx.Manager
    DTM     dtmx.Manager
}

func NewResources(ctx context.Context, cfg conf.Bootstrap, opts ResourceOptions) (*Resources, func(), error) {
    logger := opts.Logger
    if logger == nil {
        logger = logx.DefaultLogger()
    }

    if cfg.Data.Database.Enabled {
        dbCfg := cfg.Data.Database.Config
        dbCfg.Logger = logger.Named("data.dbx")
        // ...
    }
}
```

## 关键要点

1. **`logx.FromContext(ctx)`**：从 context 中提取带请求追踪信息的 Logger，推荐在 service/biz 层使用
2. **`log.Named("xxx")`**：创建子 Logger，日志中会显示 `xxx` 前缀，便于定位来源
3. **`log.WithContext(ctx)`**：将 context 中的追踪信息（request_id、trace_id 等）注入到日志中
4. **`logx.Err(err)`**：结构化记录错误，自动提取 error 的 message 和 stack
5. **`logx.DefaultLogger()`**：获取全局默认 Logger，适合在初始化阶段使用
6. **Redact**：配置中的 `redact.keys` 列表中的字段值会被自动替换为 `***`，防止敏感信息泄露
7. **Access Log**：`access_log` 配置自动记录每个 HTTP/gRPC 请求的访问日志，支持慢请求阈值