# Config 配置模块

## 概述

Kernel 使用 `configx` 包加载配置，支持 YAML 文件 + 环境变量覆盖。配置结构体定义在 `internal/conf/conf.go` 中，由 `configx.LoadConfig()` 扫描加载。

## 配置结构体定义

### 完整结构体

```go
// internal/conf/conf.go
package conf

import (
    "time"

    "github.com/aisphereio/kernel/accessx"
    "github.com/aisphereio/kernel/authn"
    "github.com/aisphereio/kernel/authn/casdoor"
    "github.com/aisphereio/kernel/authn/oidcx"
    "github.com/aisphereio/kernel/authz/spicedb"
    "github.com/aisphereio/kernel/cachex"
    "github.com/aisphereio/kernel/dbx"
    "github.com/aisphereio/kernel/dtmx"
    "github.com/aisphereio/kernel/logx"
    "github.com/aisphereio/kernel/objectstorex"
    khttp "github.com/aisphereio/kernel/transportx/http"
)

type Bootstrap struct {
    Service  ServiceConfig  `json:"service" yaml:"service"`
    Server   ServerConfig   `json:"server" yaml:"server"`
    Log      logx.Config    `json:"log" yaml:"log"`
    Data     DataConfig     `json:"data" yaml:"data"`
    Security SecurityConfig `json:"security" yaml:"security"`
    Audit    AuditConfig    `json:"audit" yaml:"audit"`
    Metrics  MetricsConfig  `json:"metrics" yaml:"metrics"`
    DTM      dtmx.Config    `json:"dtm" yaml:"dtm"`
}
```

### 各子结构体

```go
type ServiceConfig struct {
    Name    string `json:"name" yaml:"name"`
    Version string `json:"version" yaml:"version"`
    Env     string `json:"env" yaml:"env"`
}

type ServerConfig struct {
    HTTP HTTPConfig `json:"http" yaml:"http"`
    GRPC GRPCConfig `json:"grpc" yaml:"grpc"`
}

type HTTPConfig struct {
    Addr    string           `json:"addr" yaml:"addr"`
    Timeout time.Duration    `json:"timeout_ns" yaml:"timeout_ns"`
    CORS    khttp.CORSConfig `json:"cors" yaml:"cors"`
}

type GRPCConfig struct {
    Addr    string        `json:"addr" yaml:"addr"`
    Timeout time.Duration `json:"timeout_ns" yaml:"timeout_ns"`
}

type DataConfig struct {
    Database    DatabaseConfig    `json:"database" yaml:"database"`
    Cache       CacheConfig       `json:"cache" yaml:"cache"`
    ObjectStore ObjectStoreConfig `json:"object_store" yaml:"object_store"`
}

type DatabaseConfig struct {
    Enabled bool       `json:"enabled" yaml:"enabled"`
    Config  dbx.Config `json:"config" yaml:"config"`
}

type CacheConfig struct {
    Enabled bool          `json:"enabled" yaml:"enabled"`
    Config  cachex.Config `json:"config" yaml:"config"`
}

type ObjectStoreConfig struct {
    Enabled bool                `json:"enabled" yaml:"enabled"`
    Config  objectstorex.Config `json:"config" yaml:"config"`
}

type SecurityConfig struct {
    Authn        AuthnConfig                      `json:"authn" yaml:"authn"`
    Authz        AuthzConfig                      `json:"authz" yaml:"authz"`
    Access       accessx.AccessConfig             `json:"access" yaml:"access"`
    InternalCall authn.InternalServiceTokenConfig `json:"internal_call" yaml:"internal_call"`
}

type AuthnConfig struct {
    Enabled  bool           `json:"enabled" yaml:"enabled"`
    Mode     string         `json:"mode" yaml:"mode"`
    Provider string         `json:"provider" yaml:"provider"`
    OIDC     oidcx.Config   `json:"oidc" yaml:"oidc"`
    Casdoor  casdoor.Config `json:"casdoor" yaml:"casdoor"`
    CacheTTL time.Duration  `json:"cache_ttl_ns" yaml:"cache_ttl_ns"`
}

type AuthzConfig struct {
    Enabled     bool           `json:"enabled" yaml:"enabled"`
    Provider    string         `json:"provider" yaml:"provider"`
    DevAllowAll bool           `json:"dev_allow_all" yaml:"dev_allow_all"`
    SpiceDB     spicedb.Config `json:"spicedb" yaml:"spicedb"`
}

type AuditConfig struct {
    Enabled bool   `json:"enabled" yaml:"enabled"`
    Store   string `json:"store" yaml:"store"`
}

type MetricsConfig struct {
    Enabled bool   `json:"enabled" yaml:"enabled"`
    Addr    string `json:"addr" yaml:"addr"`
    Path    string `json:"path" yaml:"path"`
    Pprof   bool   `json:"pprof" yaml:"pprof"`
    Runtime bool   `json:"runtime" yaml:"runtime"`
}
```

## 配置文件示例

```yaml
# configs/config.yaml
service:
  name: app
  version: dev
  env: local

server:
  http:
    addr: 0.0.0.0:8000
    timeout_ns: 1000000000
    cors:
      enabled: true
      allowed_origins:
        - http://localhost:3000
      allowed_methods: [GET, POST, PUT, PATCH, DELETE, OPTIONS]
      allowed_headers: [Authorization, Content-Type, X-Request-ID]
      allow_credentials: true
  grpc:
    addr: 0.0.0.0:9000
    timeout_ns: 1000000000

log:
  service_name: app
  level: info
  format: console
  redact:
    enabled: true
    keys: [password, secret, token]
    value: "***"
  access_log:
    enabled: true
    skip_paths: [/healthz, /readyz, /metrics]
    slow_threshold: 1000000000

data:
  database:
    enabled: false
    config:
      driver: postgres
      dsn: "postgres://postgres:postgres@127.0.0.1:5432/app?sslmode=disable"
      max_open_conns: 20
      max_idle_conns: 10
      conn_max_lifetime_ns: 1800000000000  # 30 分钟
      conn_max_idle_time_ns: 300000000000  # 5 分钟
      slow_query_threshold_ns: 200000000   # 200ms
  cache:
    enabled: false
    config:
      driver: redis
      addrs: [127.0.0.1:6379]
      key_prefix: "app:"
  object_store:
    enabled: false
    config:
      driver: minio
      endpoint: 127.0.0.1:9000
      bucket: app

security:
  authn:
    enabled: false
    mode: gateway_trusted
    provider: casdoor
  internal_call:
    enabled: true
    header: X-Aisphere-Internal-Token
    token: "${GATEWAY_TO_APP_INTERNAL_TOKEN}"
  access:
    public_operations: ["*"]
  authz:
    enabled: false
    provider: spicedb
    dev_allow_all: true
    spicedb:
      endpoint: 127.0.0.1:50051
      token: ""
      insecure: true

audit:
  enabled: true
  store: memory

metrics:
  enabled: true
  addr: 127.0.0.1:9090
  path: /metrics
  runtime: true

dtm:
  enabled: false
  driver: dtm
  server: http://127.0.0.1:36789/api/dtmsvr
  wait_result: true
  timeout_ns: 10000000000
```

## 配置加载

在 `cmd/server/main.go` 中加载配置：

```go
// cmd/server/main.go
var bc conf.Bootstrap
if err := configx.LoadConfig(&bc, configx.File(path)); err != nil {
    panic(err)
}
```

## 关键要点

1. **时间单位**：所有 `_ns` 后缀字段使用纳秒（`time.Duration`），Kernel 自动解析
2. **环境变量覆盖**：`${VAR_NAME}` 语法支持从环境变量读取，如 `token: "${GATEWAY_TO_APP_INTERNAL_TOKEN}"`
3. **Enabled 开关**：每个模块有独立的 `enabled` 字段，关闭时使用 Noop 实现
4. **Kernel 包配置**：`dbx.Config`、`cachex.Config`、`dtmx.Config` 等直接嵌入，无需手写