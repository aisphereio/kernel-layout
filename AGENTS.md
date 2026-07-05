# Aisphere Kernel Layout Agent 规范

本仓库是 `kernel new` 的模板源。这里的默认结构会被复制到后续业务服务，所以模板必须体现 Kernel 推荐范式，而不是只生成一个能启动的 demo。

## 1. Kernel 是底层框架和契约真理

- 业务服务必须围绕 Kernel 能力开发：`configx`、`logx`、`metricsx`、`dbx`、`objectstorex`、`authn`、`authz`、`accessx`、`auditx`、`transportx`。
- 业务服务可以拥有业务资源和业务 relationship 投影，例如 Hub 写入 `skill:{name}#owner@user:{uid}`；Kernel 负责 provider-neutral 接口、生成器、middleware 和 provider adapter。
- 不允许在业务服务里绕过 Kernel 直接散落 Casdoor、SpiceDB、MinIO、PostgreSQL 客户端逻辑。provider 细节只能出现在 data/provider wiring 层。
- 缺少外部中间件服务时不要自动安装、编排或偷偷拉起基础设施；通过配置显式失败，并把需要用户准备的 endpoint、token、账号写入文档。服务自身需要的逻辑资源必须优先复用 Kernel 幂等初始化能力，例如 `dbx.auto_create_database` 自动建业务库、`objectstorex.ensure_bucket` 自动建 bucket，避免每个业务模块重复实现。

## 2. API 必须 proto 契约驱动

- 所有 JSON HTTP/gRPC 业务 API 必须先声明在 proto 中。
- proto RPC 必须声明 `google.api.http` 和 `aisphere.access.v1.policy`。
- 服务启动时只注册 generated HTTP/gRPC binding，例如 `RegisterXxxHTTPServer` / `RegisterXxxServer`。
- 禁止为普通 JSON API 在 service 层手工注册重复 HTTP route。唯一例外是 proto 无法表达的协议行为，例如浏览器 `302` redirect、raw stream、webhook 兼容入口；例外必须在 proto 注释和文档中写明原因。
- PUBLIC API 不写 authz check，只保留 exposure、audit、rate limit。AUTHENTICATED/AUTHORIZED API 的资源级授权由 generated access rule 和 biz 层业务检查共同完成。
- DELETE 不承载复杂 JSON body。复杂删除用 `POST /resource:delete`；简单删除条件用 path/query 参数。

## 3. Service/Biz/Data 分层

- `service` 只做 DTO 转换和调用 usecase，不写业务规则，不直接访问数据库/对象存储/SpiceDB。
- `biz` 负责用例编排、业务校验、状态机、权限语义和审计事件。
- `data` 负责持久化和 Kernel provider adapter 调用。PostgreSQL、MinIO、Redis、SpiceDB、Casdoor 等具体依赖只能在这里接入。
- 启动期 bootstrap 要幂等。Schema/bootstrap/relationship projection 可以在服务启动时修复，但必须可重复执行。

## 4. Authz Relationship 投影规范

- Kernel 不知道业务资源含义；业务服务负责把业务事件投影成 relationship。
- 创建资源时写 owner tuple，例如 `skill:{name}#owner@user:{uid}`。
- 分享/授权接口只写允许的业务 relation，例如 `viewer`、`editor`，不能随意开放 `owner` 转移。
- 历史数据修复应从 durable source 回填，例如从 PostgreSQL `owner_id` 回填 SpiceDB tuple。
- SpiceDB 写入优先使用 Kernel `authz.Service.WriteRelationships`，依赖 adapter 的幂等 `TOUCH` 语义；不要直接引入 authzed SDK 到业务层。

## 5. 持久化与高性能协同

- PostgreSQL 是 control plane：资源元数据、版本状态、文件索引、owner_id、manifest、审计索引。
- MinIO/S3 是 data plane：包内容、草稿文件内容、大对象、可下载产物。
- 前端频繁编辑路径必须按文件/目录增量保存，不要每次重新上传整包。
- 文件写入采用 S3-first 或 staging + metadata transaction；DB 失败要补偿删除对象，S3 失败不能写入已成功的 DB 元数据。
- 下载接口使用 ETag/sha256/If-None-Match；大文件优先走 presigned URL 或 streaming，不把大对象长期放进 PostgreSQL。
- 列表和树结构从 PostgreSQL 索引读取，文件正文按需从对象存储读取。

## 6. 本地 generator

模板 Makefile 必须同时支持：

```powershell
make tools
make tools-local KERNEL_LOCAL=../kernel
make api
make proto-check
```

修改 Kernel generator 后，必须用 `tools-local` 验证模板生成代码。业务仓库的 `make proto-check` 必须跑 `buf lint` 和 `buf-check-aisphere`。

## 7. 禁止模板占位符泄露

生成后的业务仓库不应继续出现未替换的：

```text
__KERNEL_FEATURES__
__KERNEL_DISABLED_FEATURES__
__KERNEL_PROFILE__
__KERNEL_VERSION__
```

如果新增占位符，必须同时更新 `kernel new` 的替换逻辑和模板说明。

## 8. 提交前检查

```powershell
make tools-local KERNEL_LOCAL=../kernel
make api
make proto-check
make test
make build
```
