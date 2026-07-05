# AuthN Auto Wiring in Generated Services

Generated services should use Kernel's automatic AuthN boundary instead of
hand-written token/header parsing.

## Default backend mode

```yaml
security:
  authn:
    enabled: true
    mode: gateway_trusted
  internal_call:
    enabled: true
    header: X-Aisphere-Internal-Token
    token: "${GATEWAY_TO_APP_INTERNAL_TOKEN}"
```

The framework middleware automatically:

1. Validates `X-Aisphere-Internal-Token`.
2. Requires `X-Aisphere-Auth-Verified=true`.
3. Restores `authn.Principal` from Gateway-injected `X-Aisphere-*` headers.
4. Stores the Principal in `context.Context`.

Business code only calls:

```go
principal, ok := authn.PrincipalFromContext(ctx)
```

## Gateway mode

Gateway uses:

```yaml
security:
  authn:
    enabled: true
    mode: casdoor_jwt
    oidc:
      issuer: "https://casdoor.example.com"
      discovery_url: "https://casdoor.example.com/.well-known/openid-configuration"
      jwks_url: "https://casdoor.example.com/.well-known/jwks"
      audience: ["hub-web"]
      allowed_owners: ["aisphere"]
  internal_call:
    enabled: true
    header: X-Aisphere-Internal-Token
    token: "${GATEWAY_TO_APP_INTERNAL_TOKEN}"
```

Gateway verifies the Casdoor JWT locally with JWKS, strips spoofable headers,
injects trusted Principal headers and injects the internal token before
forwarding to the backend.

## Implementation rule

Do not put JWT parsing or internal token checks in service handlers. If a service
needs a stricter mode, change `security.authn.mode`, not business code.
