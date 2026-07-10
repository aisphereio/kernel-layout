# Error 错误处理模块

## 概述

Kernel 使用 `errorx` 包提供结构化错误处理，支持错误码、HTTP/gRPC 状态码映射、错误包装和错误链。

## 使用方式

### 1. 定义业务错误

在 `internal/biz/` 包中集中定义业务错误：

```go
// internal/biz/error.go
package biz

import "github.com/aisphereio/kernel/errorx"

var (
    // 使用 errorx.NotFound 创建 404 错误
    ErrTodoNotFound = errorx.NotFound(
        errorx.Code("TODO_NOT_FOUND"),
        "todo not found",
    )

    // 使用 errorx.BadRequest 创建 400 错误
    ErrTodoInvalidArgument = errorx.BadRequest(
        errorx.Code("TODO_INVALID_ARGUMENT"),
        "invalid todo argument",
    )

    // 使用 errorx.New 创建自定义错误
    ErrOrderExpired = errorx.New(
        400,
        "ORDER_EXPIRED",
        "order has expired",
    )
)
```

### 2. 在业务逻辑中返回错误

```go
// internal/biz/todo.go
func (uc *TodoUsecase) GetTodo(ctx context.Context, id int64) (*Todo, error) {
    todo, err := uc.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err  // 直接透传 errorx 错误
    }
    return todo, nil
}

func (uc *TodoUsecase) CreateTodo(ctx context.Context, todo *Todo) (*Todo, error) {
    if err := validateTodo(todo); err != nil {
        return nil, err  // 返回 ErrTodoInvalidArgument
    }
    return uc.repo.CreateTodo(ctx, todo)
}

func (uc *TodoUsecase) DeleteTodo(ctx context.Context, id int64) error {
    if id <= 0 {
        return ErrTodoInvalidArgument
    }
    return uc.repo.DeleteTodo(ctx, id)
}
```

### 3. 在 Repository 层包装错误

```go
// internal/data/todo.go
func (r *todoRepo) FindByID(_ context.Context, id int64) (*biz.Todo, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    todo, ok := r.todos[id]
    if !ok {
        return nil, biz.ErrTodoNotFound  // 返回业务层定义的 errorx 错误
    }
    return cloneTodo(todo), nil
}
```

### 4. 错误判断

```go
// 使用 errorx 提供的判断函数
if errorx.IsNotFound(err) {
    // 处理 404
}

if errorx.IsBadRequest(err) {
    // 处理 400
}

// 或者直接比较
if errors.Is(err, biz.ErrTodoNotFound) {
    // 特定错误处理
}
```

### 5. 在测试中使用

```go
// internal/service/todo_test.go
func TestTodoServiceValidation(t *testing.T) {
    svc := newTestTodoService()

    // 验证空标题返回 BadRequest
    _, err := svc.CreateTodo(ctx, &v1.CreateTodoRequest{
        Todo: &v1.Todo{Title: " "},
    })
    if !errorx.IsBadRequest(err) {
        t.Fatalf("CreateTodo(empty title) error = %v, want bad request", err)
    }

    // 验证删除不存在的记录返回 NotFound
    _, err := svc.DeleteTodo(ctx, &v1.DeleteTodoRequest{Id: 999})
    if !errorx.IsNotFound(err) {
        t.Fatalf("DeleteTodo(missing id) error = %v, want not found", err)
    }
}
```

## 真实代码示例（脱敏自 aisphere-iam）

### 示例 1：使用领域特定错误

IAM 项目中除了 `errorx`，还使用 Kernel 提供的领域特定错误类型：

```go
// internal/service/iam.go
func (s *IAMAuthService) GetMe(ctx context.Context, req *v1.GetMeRequest) (*v1.GetMeReply, error) {
    principal, ok := authn.PrincipalFromContext(ctx)
    if !ok || !principal.IsAuthenticated() {
        return nil, authn.ErrMissingCredential("gateway principal is required")
    }
    // ...
}
```

### 示例 2：授权错误

```go
// internal/service/authz_admin.go
func (s *Service) requireGlobalAuthz(ctx context.Context, permission string) error {
    if !s.authzEnabled {
        return authz.ErrBackendFailed("authz provider is not configured", nil)
    }
    // ...
}
```

### 示例 3：在控制面操作中使用

```go
// internal/service/control_plane.go
func currentPrincipalSubject(ctx context.Context) (string, string, error) {
    principal, ok := authn.PrincipalFromContext(ctx)
    if !ok || !principal.IsAuthenticated() {
        return "", "", authn.ErrMissingCredential("kernel principal is required")
    }
    return principal.SubjectType, principal.SubjectID, nil
}
```

## Kernel 提供的错误类型

| 函数 | HTTP 状态码 | gRPC 状态码 | 说明 |
|------|-------------|-------------|------|
| `errorx.NotFound(code, msg)` | 404 | NotFound | 资源不存在 |
| `errorx.BadRequest(code, msg)` | 400 | InvalidArgument | 请求参数错误 |
| `errorx.Unauthorized(code, msg)` | 401 | Unauthenticated | 未认证 |
| `errorx.Forbidden(code, msg)` | 403 | PermissionDenied | 无权限 |
| `errorx.Internal(code, msg)` | 500 | Internal | 服务器内部错误 |
| `errorx.New(httpCode, code, msg)` | 自定义 | 自定义 | 自定义错误 |

## 领域特定错误

Kernel 还提供以下领域错误类型：

| 错误 | 包 | 说明 |
|------|-----|------|
| `authn.ErrMissingCredential(msg)` | `authn` | 缺少认证凭据 |
| `authn.ErrInvalidCredential(msg)` | `authn` | 认证凭据无效 |
| `authz.ErrBackendFailed(msg, err)` | `authz` | 授权后端失败 |
| `authz.ErrPermissionDenied(msg)` | `authz` | 权限被拒绝 |

## 关键要点

1. **集中定义**：所有业务错误在 `internal/biz/` 中集中定义，便于维护和统一
2. **语义化错误码**：使用 `errorx.Code("TODO_NOT_FOUND")` 而非数字，便于前端和文档引用
3. **错误链**：`errorx` 支持 `errors.Is()` 和 `errors.As()`，可以正确判断包装后的错误
4. **自动映射**：Kernel 的 HTTP/gRPC transport 自动将 `errorx` 错误映射为对应的 HTTP 状态码和 gRPC 状态码
5. **领域错误优先**：优先使用 Kernel 提供的领域特定错误（如 `authn.ErrMissingCredential`），而非泛化错误