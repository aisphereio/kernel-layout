# AuthN Auto Wiring in Generated Services

Generated services should use Kernel's automatic AuthN boundary instead of hand-written token/header parsing.

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

## Envoy Gateway mode

Envoy Gateway uses OIDC SecurityPolicy for browser login and JWT provider for Bearer token verification:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: myapp-oidc
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: myapp-route
  oidc:
    provider:
      issuer: "https://casdoor.weagent.cc"
    clientID: "myapp-web"
    clientSecret:
      name: casdoor-myapp-oidc
    redirectURL: "https://myapp.weagent.cc/oauth2/callback"
    logoutPath: "/logout"
    scopes: [openid, profile, email]
    refreshToken: true
    forwardAccessToken: true
    passThroughAuthHeader: true
  jwt:
    providers:
      - name: casdoor
        issuer: "https://casdoor.weagent.cc"
        audiences: ["myapp-web"]
        remoteJWKS:
          uri: "https://casdoor.weagent.cc/.well-known/jwks"
        claimToHeaders:
          - claim: sub
            header: x-aisphere-external-sub
          - claim: email
            header: x-aisphere-external-email
          - claim: name
            header: x-aisphere-external-name
          - claim: preferred_username
            header: x-aisphere-external-username
```

Envoy Gateway verifies the Casdoor JWT locally with JWKS, strips spoofable headers, injects trusted Principal headers and injects the internal token before forwarding to the backend.

## Implementation rule

Do not put JWT parsing or internal token checks in service handlers. If a service needs a stricter mode, change `security.authn.mode`, not business code.