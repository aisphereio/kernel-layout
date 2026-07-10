# Metrics 指标模块

## 概述

Kernel 使用 `metricsx` 包提供 Prometheus 指标暴露能力，支持 Counter、Histogram、Gauge 等标准指标类型，以及带标签的指标变体。

## 配置

```yaml
# configs/config.yaml
metrics:
  enabled: true
  addr: 127.0.0.1:9090
  path: /metrics
  pprof: false
  runtime: true
```

## 使用方式

### 1. 启动时创建 Metrics Manager

```go
// cmd/server/main.go
metrics := metricsx.Noop()
if bc.Metrics.Enabled {
    metrics = metricsx.NewPrometheusManager(
        bc.Service.Name,
        bc.Service.Version,
        logger,
    )
}
```

### 2. 确保非 nil 的 Manager

```go
// internal/data/data.go
metrics := metricsx.Ensure(opts.Metrics)
// metricsx.Ensure() 保证返回非 nil 实例（传入 nil 时返回 Noop）
```

### 3. 定义业务指标

```go
// internal/data/xxx.go
var (
    // Counter：计数
    createCounter = metricsx.NewCounter(
        "entity_created_total",
        "实体创建总数",
    )

    // Histogram：耗时分布
    createDuration = metricsx.NewHistogram(
        "entity_create_duration_ms",
        "实体创建耗时(ms)",
    )

    // Gauge：当前值
    queueDepth = metricsx.NewGauge(
        "queue_depth",
        "当前队列深度",
    )
)
```

### 4. 带标签的指标

```go
// CounterVec：按标签维度统计
var apiRequestCounter = metricsx.NewCounterVec(
    "api_request_total",
    "API 请求总数",
    "method", "path", "status",
)

// 使用
apiRequestCounter.WithLabelValues("GET", "/v1/users", "200").Inc()
```

### 5. 在业务代码中记录指标

```go
// internal/data/xxx.go
func (r *Repo) Create(ctx context.Context, entity *Entity) error {
    start := time.Now()
    defer func() {
        createCounter.Inc()
        createDuration.Observe(float64(time.Since(start).Milliseconds()))
    }()
    // ... 数据库操作 ...
}
```

### 6. 将 Metrics 注入到 Kernel 模块

```go
// internal/data/data.go
if cfg.Data.Database.Enabled {
    dbCfg := cfg.Data.Database.Config
    dbCfg.Logger = logger.Named("data.dbx")
    dbCfg.Metrics = metrics
    dbCfg.MetricsEnabled = dbCfg.MetricsEnabled && cfg.Metrics.Enabled
    db, err := dbx.New(dbCfg)
    // ...
}

if cfg.Data.Cache.Enabled {
    cacheCfg := cfg.Data.Cache.Config
    cacheCfg.Logger = logger.Named("data.cachex")
    cacheCfg.Metrics = metrics
    cacheCfg.MetricsEnabled = cacheCfg.MetricsEnabled && cfg.Metrics.Enabled
    cache, err := cachex.New(cacheCfg)
    // ...
}
```

## 真实代码示例（脱敏自 aisphere-iam）

### 示例 1：在 data 层初始化 Metrics

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
    metrics := metricsx.Ensure(opts.Metrics)

    // 将 metrics 注入到各个子模块
    if cfg.Data.Database.Enabled {
        dbCfg := cfg.Data.Database.Config
        dbCfg.Logger = logger.Named("data.dbx")
        dbCfg.Metrics = metrics
        dbCfg.MetricsEnabled = dbCfg.MetricsEnabled && cfg.Metrics.Enabled
        db, err := dbx.New(dbCfg)
        // ...
    }
}
```

### 示例 2：在 DTM 中注入 Metrics

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

### 示例 3：在 Authn/Authz 中注入 Metrics

```go
// internal/data/data.go
func newAuthenticator(cfg conf.AuthnConfig, logger logx.Logger, metrics metricsx.Manager, metricsEnabled bool) (authn.Authenticator, error) {
    switch cfg.Provider {
    case "", "casdoor":
        casdoorCfg := cfg.Casdoor
        casdoorCfg.Logger = logger.Named("authn.casdoor")
        casdoorCfg.Metrics = metrics
        casdoorCfg.MetricsEnabled = casdoorCfg.MetricsEnabled && metricsEnabled
        return casdoor.New(casdoorCfg)
    }
}

func newAuthorizer(cfg conf.AuthzConfig, logger logx.Logger, metrics metricsx.Manager, metricsEnabled bool) (authz.Authorizer, func() error, error) {
    switch cfg.Provider {
    case "", "spicedb":
        spiceCfg := cfg.SpiceDB
        spiceCfg.Logger = logger.Named("authz.spicedb")
        spiceCfg.Metrics = metrics
        spiceCfg.MetricsEnabled = spiceCfg.MetricsEnabled && metricsEnabled
        return spicedb.New(spiceCfg)
    }
}
```

## 关键要点

1. **`metricsx.Ensure(m)`**：保证返回非 nil 的 Manager，传入 nil 时返回 Noop 实现，避免空指针
2. **`metricsx.Noop()`**：当 metrics 禁用时使用，所有操作无副作用
3. **`metricsx.NewPrometheusManager(name, version, logger)`**：创建 Prometheus 实现，自动注册 `/metrics` 端点
4. **`MetricsEnabled` 双重检查**：每个子模块的 `MetricsEnabled` 需要与全局 `cfg.Metrics.Enabled` 做 `&&`，确保全局关闭时子模块也不记录
5. **指标类型**：
   - `NewCounter(name, help)`：单调递增计数器
   - `NewHistogram(name, help)`：值分布统计（如耗时）
   - `NewGauge(name, help)`：可增可减的当前值
   - `NewCounterVec(name, help, labelKeys...)`：带标签的计数器