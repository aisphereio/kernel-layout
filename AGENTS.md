# Aisphere Kernel Layout Agent 规范

本仓库是 `kernel new` 的模板源。任何问题一旦进入这里，后续所有新业务仓库都会复制同样的问题。

## 1. 模板必须体现推荐范式

- 模板不能只生成能启动的 demo，还必须生成符合 Kernel 规范的工程骨架。
- HTTP/gRPC server 必须装配 `requestinfo + authn + access` middleware。
- proto 暴露接口必须声明 `google.api.http` 和 `aisphere.access.v1.policy`。
- 生成项目的 README/AGENTS/docs 必须告诉开发者如何使用 Kernel 工具链。

## 2. 本地 generator

模板 Makefile 必须同时支持：

```powershell
make tools
make tools-local KERNEL_LOCAL=../kernel
```

修改 Kernel generator 后，必须用 `tools-local` 验证模板生成代码。

## 3. 禁止把模板占位符泄露到业务仓库

生成后的业务仓库不应继续出现未替换的：

```text
__KERNEL_FEATURES__
__KERNEL_DISABLED_FEATURES__
__KERNEL_PROFILE__
__KERNEL_VERSION__
```

如果新增占位符，必须同时更新 `kernel new` 的替换逻辑和模板说明。

## 4. 提交前检查

```powershell
make tools-local KERNEL_LOCAL=../kernel
make api
make proto-check
make test
```
