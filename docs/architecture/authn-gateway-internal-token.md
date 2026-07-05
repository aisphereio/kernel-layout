# AuthN Gateway Boundary Architecture

标准部署模式：

```text
Client
  -> Gateway: Authorization: Bearer <Casdoor JWT>
Gateway
  -> Casdoor OIDC/JWKS: 拉取公钥并缓存
  -> Gateway: 校验 JWT 签名、issuer、audience、exp、owner、alg
  -> Backend: 注入 X-Aisphere-* Principal headers
  -> Backend: 注入 X-Aisphere-Internal-Token
Backend
  -> 校验 internal token
  -> 读取可信 Principal
  -> 执行业务
```

## 包职责

- `authn/oidcx`: OIDC discovery + JWKS + JWT verifier，所有组件复用。
- `authn`: Principal、trusted headers、internal-service-token 配置与校验。
- `gatewayx`: Gateway dispatch、清理伪造 headers、注入 Principal 与 internal token。
- 业务服务: 默认 `gateway_trusted`，不处理登录/session/refresh。

## Redis 缓存

默认只使用进程内 JWKS 缓存。token -> Principal 缓存是优化项，不是认证事实来源；如启用，key 必须是 token hash，TTL 必须小于 token 剩余有效期。

## 安全边界

第一版使用 `X-Aisphere-Internal-Token`；生产建议叠加 Kubernetes NetworkPolicy，只允许 Gateway Pod 访问后端服务。mTLS/SPIFFE 可作为后续增强。
