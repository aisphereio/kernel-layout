# AuthN 全流程设计参考

本文档是 Kernel layout 项目模板的 AuthN 设计参考。如果你使用本 layout 生成新服务，需要设计实现认证全流程，可以参考本文的架构、代码组织和配置方式。

## 架构总览

```text
Casdoor 签发 JWT
  ↓
Envoy Gateway OIDC 登录 / JWT 验证
  ↓
Envoy Gateway 清理伪造 X-Aisphere-* header
  ↓
Envoy Gateway claimToHeaders 注入可信身份 headers
  ↓
Envoy Gateway 转发到后端服务
  ↓
后端 Kernel middleware 自动恢复 authn.Principal
  ↓
业务 Handler 通过 authn.PrincipalFromContext(ctx) 读取
```

## 涉及模块

| 模块 | 责任 | 关键文件 |
| --- | --- | --- |
| `kernel/authn/oidcx` | 通用 OIDC/JWKS verifier，不依赖 Casdoor SDK | `config.go`, `jwks.go`, `verifier.go` |
| `kernel/authn/cached_authenticator.go` | token→Principal 验证结果缓存，SHA-256 key，TTL 受 token exp 限制 | `cached_authenticator.go` |
| `kernel/authn/trusted_headers.go` | Gateway 可信 header 定义、注入、剥离、重建 | `trusted_headers.go` |
| `kernel/authn/trusted_authenticator.go` | 信任 Gateway 注入的 Principal（gateway_trusted 模式） | `trusted_authenticator.go` |
| `Envoy Gateway` | 主认证边界，OIDC 登录、JWT 验证、claimToHeaders | Envoy Gateway CRD |
| `aisphere-iam` | 后端服务，使用 `gateway_trusted` 模式 | `internal/data/data.go`, `internal/conf/conf.go` |
| `aisphere-hub` | 后端服务，使用 `gateway_trusted` 模式 | `internal/data/data.go` |

## Kernel authn/oidcx 包

### 目录结构

```text
kernel/authn/oidcx/
├── config.go      # 配置结构体 + Normalized() 默认值
├── jwks.go        # JWKS 发现、拉取、缓存、RSA/EC 公钥解析
└── verifier.go    # JWT 签名验证 + claims 校验 + Principal 映射
```

### 配置项

| 字段 | 类型 | 说明 | 默认值 |
| --- | --- | --- | --- |
| `provider` | string | 映射到 `Principal.Provider` | `"oidc"` |
| `issuer` | string | 期望的 JWT `iss` 值 | 从 discovery 获取 |
| `discovery_url` | string | OIDC Discovery 端点 | — |
| `jwks_url` | string | JWKS 公钥端点 | 从 discovery 获取 |
| `audience` | []string | 接受的 `aud` 值列表 | — |
| `allowed_algs` | []string | 允许的签名算法 | `[RS256, RS512, ES256, ES512]` |
| `allowed_owners` | []string | 允许的 Casdoor `owner` | — |
| `jwks_cache_ttl` | duration | JWKS 进程内缓存 TTL | 10m |
| `clock_skew` | duration | exp/nbf/iat 时钟偏差 | 60s |

### 校验内容

1. **JWT 签名** — 用 JWKS 公钥验证
2. **iss** — 必须匹配配置的 issuer
3. **aud** — 至少匹配一个配置的 audience
4. **exp** — 未过期（含 clock_skew 容差）
5. **nbf** — 已生效（含 clock_skew 容差）
6. **iat** — 不超前于当前时间（含 clock_skew 容差）
7. **alg** — 必须在 allowed_algs 中
8. **owner** — 如果在 allowed_owners 中配置了，必须匹配

### Principal 映射

```go
Principal{
    SubjectID:   sub / name,
    SubjectType: "user",
    Provider:    cfg.Provider,
    ExternalID:  owner/username,
    Issuer:      iss,
    Audience:    aud,
    OrgID:       owner,
    TenantID:    owner,
    Username:    name,
    Name:        displayName,
    Email:       email,
    Groups:      groups,
    Roles:       roles,
    Scopes:      scope,
    AuthMethod:  "oidc",
    IssuedAt:    iat,
    ExpiresAt:   exp,
}
```

## 后端服务认证模式

### 推荐模式：gateway_trusted

后端信任 Envoy Gateway 注入的 `x-aisphere-*` header，不再验 JWT。

```yaml
security:
  authn:
    enabled: true
    mode: gateway_trusted
```

### 模式对比

| 特性 | gateway_trusted（推荐） | casdoor_jwt（备选） |
| --- | --- | --- |
| 后端验 JWT 签名 | ❌ | ✅ |
| 依赖 JWKS 可用性 | ❌ | ✅ |
| 适合网络隔离 | ✅ | ❌ |
| 需要 mTLS/NetworkPolicy | ✅ | ❌ |
| 性能（省一次验签） | 快 | 慢 |

## 缓存策略

| 缓存对象 | 默认 | 说明 |
| --- | --- | --- |
| JWKS 公钥 | 进程内缓存 | 必须缓存，避免每次请求拉 JWKS |
| token → Principal | 可选 Redis / 内存 | 高 QPS 时启用 |
| raw token | 不缓存 | 避免泄露 bearer token |
| private key / client secret | 永远不缓存 | 不能进入 Redis |
| 过期 token | 不缓存 / 自动失效 | TTL 被 token exp 限制 |

### 缓存 TTL 策略

```go
effective_ttl = min(configured_cache_ttl, token_exp - now)
```

所以不会从缓存里放行过期 token。

## 验证命令

### 编译检查

```powershell
# Kernel
cd E:\coding\aisphereio\kernel
go test ./authn/... -v -count=1

# IAM
cd E:\coding\aisphereio\aisphere-iam
go build ./...
go test ./... -count=1 -short

# Hub
cd E:\coding\aisphereio\aisphere-hub
go build ./...
go test ./... -count=1 -short
```

### 端到端验证

```powershell
# 1. 确保 Casdoor 运行在 http://36.137.200.194:30082
# 2. 启动 IAM
cd E:\coding\aisphereio\aisphere-iam
go run ./cmd/aisphere-iam --conf ./configs/config.local.yaml

# 3. 启动 Hub
cd E:\coding\aisphereio\aisphere-hub
go run ./cmd/aisphere-hub --conf ./configs/config.local.yaml

# 4. 获取 Casdoor JWT
# 5. 验证 Gateway 认证
curl.exe -i http://127.0.0.1:18080/v1/iam/me `
  -H "Authorization: Bearer <access_token>"
```