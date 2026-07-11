# AuthN Architecture

## Authentication Flow

The Kernel Layout uses a **gateway-trusted** authentication model. Envoy Gateway is the authentication boundary; backend services trust the headers injected by the gateway.

### Full Flow

```text
1. Casdoor issues JWT (OIDC login or direct token)
2. Client sends request with Authorization: Bearer <JWT>
3. Envoy Gateway:
   a. Verifies JWT signature via JWKS (local, cached)
   b. Validates iss, aud, exp, nbf, iat, alg, owner
   c. Strips any existing X-Aisphere-* headers (spoof protection)
   d. Injects trusted X-Aisphere-* Principal headers via claimToHeaders
   e. Injects X-Aisphere-Internal-Token
4. Backend Kernel middleware:
   a. Validates X-Aisphere-Internal-Token
   b. Requires X-Aisphere-Auth-Verified=true
   c. Restores authn.Principal from X-Aisphere-* headers
   d. Stores Principal in context.Context
5. Business code calls authn.PrincipalFromContext(ctx)
```

### Default Configuration

```yaml
security:
  authn:
    enabled: true
    mode: gateway_trusted    # Trust Gateway-injected headers
    provider: casdoor
    cache_ttl_ns: 300000000000  # 5 min
  internal_call:
    enabled: true
    header: X-Aisphere-Internal-Token
    token: "${GATEWAY_TO_APP_INTERNAL_TOKEN}"
```

## Principal Model

```go
type Principal struct {
    SubjectID      string            // User unique ID
    SubjectType    string            // user / service / robot
    Provider       string            // casdoor / oidc
    ExternalID     string
    Issuer         string
    Audience       string
    TenantID       string
    OrgID          string
    AppID          string
    ProjectID      string
    Username       string
    Name           string
    Email          string
    Phone          string
    Roles          []string
    Groups         []string
    Scopes         []string
    AuthMethod     string            // gateway / jwt / internal
    Attributes     map[string]string
    IssuedAt       time.Time
    ExpiresAt      time.Time
}
```

## Usage in Business Code

```go
// In service layer
principal, ok := authn.PrincipalFromContext(ctx)
if !ok || !principal.IsAuthenticated() {
    return nil, authn.ErrMissingCredential("principal is required")
}
userID := principal.SubjectID
```

## Security Boundary

- **Envoy Gateway** is the primary authn boundary (OIDC login, JWT verification)
- **Backend services** use `gateway_trusted` mode — no JWT parsing in handlers
- **Internal service calls** use `X-Aisphere-Internal-Token` shared secret
- **Production hardening**: Kubernetes NetworkPolicy to allow only Envoy Gateway Pods → backend; mTLS/SPIFFE as future enhancement

## Key Rules

1. Do not put JWT parsing or internal token checks in service handlers
2. If a service needs a stricter mode, change `security.authn.mode`, not business code
3. Token → Principal cache is an optimization, not authn source of truth; key must be token hash, TTL must be less than token remaining validity

## Source Map

| File | Role |
|------|------|
| `docs/architecture/authn-auto-wiring-gateway-trusted.md` | Auto-wiring guide |
| `docs/architecture/authn-gateway-internal-token.md` | Gateway boundary architecture |
| `docs/design/authn-full-flow.md` | Full authn design reference |
| `docs/modules/authn.md` | AuthN module usage guide |
| `internal/server/access.go` | Security runtime construction |
| `internal/conf/conf.go` | Authn/Authz config DTOs |