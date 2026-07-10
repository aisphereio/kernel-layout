# AuthN Gateway Boundary Architecture

Standard deployment model:

```text
Client
  -> Envoy Gateway: Authorization: Bearer <Casdoor JWT>
Envoy Gateway
  -> Casdoor OIDC/JWKS: fetch public key and cache
  -> Envoy Gateway: verify JWT signature, issuer, audience, exp, owner, alg
  -> Backend: inject X-Aisphere-* Principal headers
  -> Backend: inject X-Aisphere-Internal-Token
Backend
  -> verify internal token
  -> read trusted Principal
  -> execute business logic
```

## Package responsibilities

- `authn/oidcx`: OIDC discovery + JWKS + JWT verifier, shared by all components.
- `authn`: Principal, trusted headers, internal-service-token config and validation.
- Business services: default `gateway_trusted`, no login/session/refresh handling.

## Redis cache

Default uses only in-process JWKS cache. token -> Principal cache is an optimization, not the authn source of truth; if enabled, key must be token hash, TTL must be less than token remaining validity.

## Security boundary

First version uses `X-Aisphere-Internal-Token`; production should add Kubernetes NetworkPolicy to allow only Envoy Gateway Pods to access backend services. mTLS/SPIFFE can be a subsequent enhancement.