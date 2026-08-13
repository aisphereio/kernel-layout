# Aisphere Kernel Layout Agent 规范

本仓库是 `kernel new` 的模板源。这里的默认结构会被复制到后续业务服务，所以模板必须体现 Kernel 推荐范式，而不是只生成一个能启动的 demo。

## 1. Kernel 是底层框架和契约真理

- 业务服务必须围绕 Kernel 能力开发：`configx`、`logx`、`metricsx`、`serverx`、`dbx`、`migrationx`、`dbrepo`、`objectstorex`、`authn`、`authz`、`accessx`、`auditx`、`transportx`。
- 业务服务可以拥有业务资源和业务 relationship 投影，例如 Hub 写入 `skill:{name}#owner@user:{uid}`；Kernel 负责 provider-neutral 接口、生成器、middleware 和 provider adapter。
- 不允许在业务服务里绕过 Kernel 直接散落 Casdoor、SpiceDB、MinIO、PostgreSQL 客户端逻辑。provider 细节只能出现在 data/provider wiring 层。
- 缺少外部中间件服务时不要自动安装、编排或偷偷拉起基础设施；通过配置显式失败，并把需要用户准备的 endpoint、token、账号写入文档。

## 2. API 必须 proto 契约驱动

- 所有 JSON HTTP/gRPC 业务 API 必须先声明在 proto 中。
- proto RPC 必须声明 `google.api.http` 和 `aisphere.access.v1.policy`。
- `make api` 生成 HTTP/gRPC binding、RequestInfoResolver、AccessResolver、Gateway Manifest/Invoker 和 ServiceModule；业务不得手写等价 glue code。
- 服务启动时只注册 generated HTTP/gRPC binding，例如 `RegisterXxxHTTPServer` / `RegisterXxxServer`。
- 禁止为普通 JSON API 在 service 层手工注册重复 HTTP route。唯一例外是 proto 无法表达的协议行为，例如浏览器 `302` redirect、raw stream、webhook 兼容入口；例外必须在 proto 注释和文档中写明原因。
- `PUBLIC` 不执行资源级 authz；`AUTHENTICATED/AUTHORIZED` 的认证、资源级授权和 access audit 由 generated resolver + Kernel `authn/authz/accessx` server chain 自动执行。handler/biz 不得再手写“调 IAM 验 token”或“调 SpiceDB CheckPermission”。
- biz 只处理真正的业务不变量、状态机和需要 admission/relationship projection 的领域语义；不要重复做框架已经完成的 endpoint access check。
- DELETE 不承载复杂 JSON body。复杂删除用 `POST /resource:delete`；简单删除条件用 path/query 参数。

## 3. Service/Biz/Data 分层

- `service` 只做 DTO 转换和调用 usecase，不写业务规则，不直接访问数据库/对象存储/SpiceDB。
- `biz` 负责用例编排、业务校验、状态机和领域关系变化，不重复实现 endpoint authn/authz。
- `data` 负责持久化和 Kernel provider adapter 调用。PostgreSQL、MinIO、Redis、SpiceDB、Casdoor 等具体依赖只能在这里接入。
- Data/repository 应由 Kernel 管理的 boot dependency 注入业务实例，禁止在 handler 内临时 `gorm.Open`、`sql.Open` 或重新构造 repository。
- 启动期 bootstrap 要幂等；但生产 schema migration 不由业务 bootstrap 临时拼 SQL。

## 4. Authz Relationship 投影规范

- Kernel 不知道业务资源关系如何变化；业务服务负责把业务事件投影成 relationship，但 endpoint 是否允许访问由 generated access policy 自动判定。
- 创建资源时可以写 owner tuple，例如 `skill:{name}#owner@user:{uid}`。
- 分享/授权接口只写允许的业务 relation，例如 `viewer`、`editor`，不能随意开放 `owner` 转移。
- 历史数据修复应从 durable source 回填，例如从 PostgreSQL `owner_id` 回填 SpiceDB tuple。
- SpiceDB relationship 写入使用 Kernel `authz.Service`/provider-neutral 接口，不直接引入 authzed SDK 到业务层。

## 5. 数据开发范式

- SQL migration 是数据库 schema 的唯一真实来源；不要把建表 SQL 塞进 proto。
- migration 文件放在 `migrations/`，生产默认由 Kernel `migrationx` 委托真实 goose 执行/校验。
- GORM AutoMigrate 只允许 dev/test 快速验证，不作为生产 schema 迁移方案。
- 普通 CRUD 优先使用 `dbrepo.ResourceRepository` 或沉淀专用 repository；handler 不到处拼 GORM 查询。
- `TenantScoped` / `OwnerScoped` repository 缺 scope 必须 fail-closed；Patch 必须受字段白名单/保护字段约束。
- 多副本服务禁止在启动阶段无锁执行 migration apply；生产服务默认 `validate`，迁移由单实例/发布 Job 完成。
- PostgreSQL/MySQL 的具体 driver、GORM 错误和 SQL 错误不能泄漏到业务层，必须经 `dbx` 归一化。

## 6. 持久化与高性能协同

- PostgreSQL 是 control plane：资源元数据、版本状态、文件索引、owner_id、manifest、审计索引。
- MinIO/S3 是 data plane：包内容、草稿文件内容、大对象、可下载产物。
- 前端频繁编辑路径必须按文件/目录增量保存，不要每次重新上传整包。
- 文件写入采用 S3-first 或 staging + metadata transaction；DB 失败要补偿删除对象，S3 失败不能写入已成功的 DB 元数据。
- 下载接口使用 ETag/sha256/If-None-Match；大文件优先走 presigned URL 或 streaming，不把大对象长期放进 PostgreSQL。
- 列表和树结构从 PostgreSQL 索引读取，文件正文按需从对象存储读取。

## 7. 本地 generator

模板 Makefile 必须同时支持：

```powershell
make tools
make tools-local KERNEL_LOCAL=../kernel
make api
make proto-check
make verify
```

修改 Kernel generator 后，必须用 `tools-local` 验证模板生成代码。业务仓库的 `make proto-check` 必须跑 `buf lint` 和 `buf-check-aisphere`。

提交的 proto 和 generated artifacts 必须保持一致。修改 proto 后必须运行 `make api`，不能只改 proto 或只手改 `.pb.go`。

## 8. 禁止模板占位符泄露

生成后的业务仓库不应继续出现未替换的：

```text
__KERNEL_FEATURES__
__KERNEL_DISABLED_FEATURES__
__KERNEL_PROFILE__
__KERNEL_VERSION__
```

如果新增占位符，必须同时更新 `kernel new` 的替换逻辑和模板说明。

## 9. 提交前检查

```powershell
make tools-local KERNEL_LOCAL=../kernel
make api
make proto-check
make test
make build
make verify
```

Kernel 的 release gate 还会真实执行 `kernel new`，再用当前 Kernel commit 对生成项目运行 `make verify`。模板不能依赖“生成后人工补代码”才能通过。
