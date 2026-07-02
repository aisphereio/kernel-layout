# Kernel Layout 合规说明

本模板用于生成新的 Kernel 服务项目。模板必须默认展示正确的工程实践，否则新生成项目会继续复制错误。

## 本次约束

生成项目应包含：

- `AGENTS.md`：约束 AI Agent 和人类开发者不要绕过 Kernel 范式。
- `make tools-local`：支持从本地 `../kernel` 安装 generator。
- HTTP/gRPC middleware：`requestinfo -> authn -> access`。
- proto-first Todo 示例：展示 access policy、audit、rate limit、Gateway manifest 生成入口。

## 重要边界

- 业务项目不得直接 import `cmd/protoc-gen-*`。
- 如果 `_kernel.pb.go` 编译依赖缺失，应修 Kernel generator，而不是在业务仓库长期手写 glue。
- Gateway 路由必须由 proto contract 生成，不允许模板中出现手写外部 path 清单。
