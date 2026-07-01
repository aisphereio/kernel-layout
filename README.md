# Aisphere Kernel Layout

Standalone service layout for `github.com/aisphereio/kernel/cmd/kernel`.

## Use

```powershell
go install github.com/aisphereio/kernel/cmd/kernel@latest
kernel new skill-service --repo https://github.com/aisphereio/kernel-layout.git
cd skill-service
make tools
make api
make proto-check
```

This repository is intentionally a template repository. `kernel new` copies it, replaces the module path, applies Kernel scaffold options, then renames `cmd/server` to `cmd/<project-name>`.

## Included defaults

- Features: `__KERNEL_FEATURES__`
- DB: `__KERNEL_DB_DRIVER__`
- Cache: `__KERNEL_CACHE_DRIVER__`
- Object storage: `__KERNEL_OBJECTSTORE_DRIVER__`
- Authn: `__KERNEL_AUTHN_PROVIDER__`
- Authz: `__KERNEL_AUTHZ_PROVIDER__`
- Kernel version for generated Makefile tools: `__KERNEL_VERSION__`

## Development flow

```text
kernel new <service>
  -> write proto under api/<domain>/v1
  -> declare google.api.http + aisphere.access.v1.policy
  -> make tools
  -> make api
  -> implement business service
```
