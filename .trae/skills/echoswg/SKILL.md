---
name: "echoswg"
description: "Builds Go APIs on the custom echoswg framework. Invoke when adding routes, auth, Swagger docs, or bootstrapping an API module in this repository."
---

# Echoswg Go API

用于当前仓库及其衍生业务项目中的 Go API 开发。它帮助你基于 `echoswg` 框架自动完成路由配置、Swagger 文档生成接入、认证授权接入，并遵循本 skill 与 `example-skeleton` 中定义的代码组织方式。

## 何时调用

在以下场景主动调用本 skill：

- 需要新增或修改基于 `echoswg` 的 HTTP API。
- 需要在 `controller` 中注册路由，并让 Swagger 自动生成接口文档。
- 需要接入或调整 `security.RequireAuth(...)` 鉴权。
- 需要按照本 skill 或 `example-skeleton` 的模式创建 `controller`、`service`、`security`、`main` 装配代码。
- 用户提到“按示例新增接口”“补框架配置”“生成 swagger 路由”“写 controller/service 模板”等类似需求。

如果任务只是修改纯业务逻辑、与路由和 API 协议无关，不必调用本 skill。

## 先理解的框架事实

- 路由不是在 `main` 中逐条挂载，而是在各个 `controller` 文件的 `init()` 中自动注册。
- `main` 通过匿名导入 `controller` 包触发这些 `init()`，例如 `_ "your/module/controller"`。
- 路由统一使用 `echoswg.NewApiGroup(...)` 创建分组。
- 每条路由在注册时同时完成两件事：
  - 注册 Echo handler
  - 根据 handler 签名自动生成 Swagger path、parameter、requestBody、schema
- Swagger 文档不是手写 JSON，而是由框架根据 handler 入参和返回值自动推导。
- `WithOperationId(...)` 是默认必写项；`WithDescription(...)`、`WithSummary(...)` 视需要补充。

## 目录与职责约定

- `controller/`
  - HTTP 协议层入口。
  - 负责声明路由、选择是否鉴权、接收请求参数、调用 `service`。
  - 不直接写数据库、Redis、事务、复杂状态机。
- `service/`
  - 业务核心层。
  - 负责数据库访问、事务、第三方调用、状态流转、权限细则判断。
  - 对外暴露包级单例，例如 `service.User`、`service.Payment`。
- `security/`
  - 负责认证、用户上下文、角色判断、token 解析。
  - controller 通过 `security.RequireAuth(...)` 接入。
- `main.go`
  - 负责初始化配置、日志、Redis、数据库、迁移、中间件、Swagger 页面、HTTP 启动。
  - 不手工逐条注册业务路由，只匿名导入 `controller` 包。

## 路由注册规范

每个资源域通常对应一个 controller 文件，并在 `init()` 中完成分组与路由注册：

```go
package controller

import (
    "your/module/security"
    "your/module/service"
    "your/module/util"

    "github.com/yb7/echoswg"
)

type OrderController struct{}

func init() {
    c := new(OrderController)
    g := echoswg.NewApiGroup(util.EchoInstance, "Order", "/api/orders")

    g.POST("", security.RequireAuth(), c.Create, echoswg.WithOperationId("createOrder"))
    g.GET("", security.RequireAuth(), c.Query, echoswg.WithOperationId("queryOrders"))
    g.GET("/:ID", security.RequireAuth(), c.GetByID, echoswg.WithOperationId("getOrderById"))
    g.PUT("/:ID", security.RequireAuth(), c.Update, echoswg.WithOperationId("updateOrder"))
    g.DELETE("/:ID", security.RequireAuth(), c.Delete, echoswg.WithOperationId("deleteOrder"))
}
```

约定如下：

- `tag` 使用资源域名称，供 Swagger tag 分组。
- `prefix` 使用资源根路径，例如 `/api/orders`。
- 每条路由都写唯一的 `WithOperationId(...)`。
- `WithDescription(...)` 用于补充接口说明；不需要时可省略。
- 不要在别处重复手工维护 Swagger path。

### 路由操作串联

`g.POST(...)`、`g.GET(...)`、`g.PUT(...)`、`g.DELETE(...)` 接收的参数里，函数类型会按顺序组成一条执行链。框架会从左到右依次调用这些函数，并把上一步返回的非 `error` 结果缓存起来，供下一步按类型自动注入。

例如：

```go
g.POST(
    "",
    security.RequireAuth(security.RoleAnonymous),
    c.Create,
    echoswg.WithOperationId("createDemo"),
    echoswg.WithDescription("create demo"),
)
```

它的实际含义是：

1. 先执行 `security.RequireAuth(security.RoleAnonymous)` 返回的函数。
2. 该函数返回 `security.AuthCtx` 后，框架会把这个值缓存到当前调用链中。
3. 再执行 `c.Create` 时，如果它的参数列表里声明了 `security.AuthCtx`，框架就会把上一步产出的 `AuthCtx` 自动注入进去。
4. `WithOperationId(...)`、`WithDescription(...)` 这类 `PathSchemaOption` 只参与 Swagger 元数据构建，不参与运行时执行链。

也就是说，下面这种写法并不是“把 `RequireAuth(...)` 当成普通中间件挂上去”，而是“把它当作链路中的前置函数，让它产出的类型值继续传给后续 handler”。

再看一个更直观的例子：

```go
g.POST(
    "",
    security.RequireAuth(),
    c.Create,
    echoswg.WithOperationId("createOrder"),
)

func (*OrderController) Create(ctx security.AuthCtx, req *struct {
    Body *service.CreateOrderVo
}) (*service.OrderVo, error) {
    return service.Order.Create(ctx, req.Body)
}
```

上面的调用链中：

- `RequireAuth()` 负责从请求中解析登录态，并返回 `security.AuthCtx`
- `c.Create` 负责消费 `security.AuthCtx` 和请求体
- 请求体 `req` 则不是前一步返回的，而是框架根据 handler 参数类型自动从 path/query/body 构造出来的

生成代码时必须遵循以下规则：

- 需要前置鉴权或上下文构造时，把函数放在业务 handler 前面。
- 后置 handler 的参数类型必须能被前面步骤的返回值或框架内建注入机制满足。
- 如果前一步返回某个自定义类型，后一步可以直接声明同类型参数来接收它。
- 不要把 `WithOperationId(...)`、`WithDescription(...)` 当作可执行 handler，它们只是文档选项。
- 当链路较长时，优先保证每一步函数职责单一，例如“鉴权 -> 装配上下文 -> 业务处理”。

### 多步链路模板

作为补充特性说明，框架也支持比“鉴权 -> Handler”更长的链路，例如：

- `RequireAuth -> LoadTenant -> CheckPermission -> Handler`

这种写法适合把“认证”“租户上下文装配”“权限校验”“业务处理”拆成多个小步骤，每一步只做一件事，并把结果类型继续传给下一步。

示例：

```go
g.POST(
    "/orders",
    security.RequireAuth(),
    tenant.LoadTenant,
    permission.CheckPermission(permission.OrderWrite),
    c.CreateOrder,
    echoswg.WithOperationId("createOrder"),
)
```

可以把它理解为下面这样的类型流转：

```go
RequireAuth()        : func(*echo.Context) (security.AuthCtx, error)
LoadTenant           : func(security.AuthCtx) (*tenant.Context, error)
CheckPermission(...) : func(*tenant.Context) (*tenant.Context, error)
CreateOrder          : func(*tenant.Context, *CreateOrderReq) (*OrderVo, error)
```

上面这条链的运行语义是：

1. `RequireAuth()` 先从请求中构造 `security.AuthCtx`
2. `LoadTenant` 接收 `security.AuthCtx`，查询或装配租户信息，返回 `*tenant.Context`
3. `CheckPermission(...)` 接收 `*tenant.Context`，完成权限校验，校验通过后继续返回同一个或增强后的上下文对象
4. `CreateOrder` 最终消费 `*tenant.Context` 和框架自动构造的请求体 `*CreateOrderReq`

需要注意：

- 每一步是否能串起来，关键取决于“后一步的入参类型”能否被前一步返回值满足。
- 如果某一步只做校验、不产生新类型，最简单的方式是返回原上下文类型本身，例如 `func(ctx *tenant.Context) (*tenant.Context, error)`。
- 如果某一步需要补充更多信息，也可以返回一个新的聚合上下文类型，再让后续步骤都消费这个新类型。
- 请求体、路径参数、查询参数仍由框架根据最终 handler 的参数类型自动构造，它们不是必须从前置步骤传递。
- 这种模式更像“typed pipeline”而不是传统的 Echo middleware；设计时优先考虑“类型是否清晰”而不是“步骤是否越多越好”。

建议只在以下场景使用多步链路：

- 需要复用一段前置逻辑给多个接口
- 需要把认证、租户、权限、资源装载拆开，提升可读性
- 需要让不同阶段产出的上下文对象被后续业务直接消费

不建议为了简单接口强行拆很多步；如果只是单纯鉴权后执行业务，`RequireAuth() -> Handler` 往往已经足够清晰。

## Handler 签名规范

框架会根据 handler 参数类型自动注入 `*echo.Context`、构造请求对象、识别 path/query/body，并在返回时自动输出 JSON。

### 常见签名

受保护接口：

```go
func (*OrderController) Create(ctx security.AuthCtx, req *struct {
    Body *service.CreateOrderVo `jsonschema_description:"创建订单请求体"`
}) (*service.OrderVo, error) {
    return service.Order.Create(ctx, req.Body)
}
```

带路径参数：

```go
func (*OrderController) GetByID(ctx security.AuthCtx, req *struct {
    ID int `json:"id" jsonschema_description:"订单ID"`
}) (*service.OrderVo, error) {
    return service.Order.GetByID(ctx, req.ID)
}
```

查询接口：

```go
func (*OrderController) Query(ctx security.AuthCtx, req *service.OrderQueryVo) ([]*service.OrderVo, error) {
    return service.Order.Query(ctx, req)
}
```

需要直接操作 Cookie、上传文件或原始请求：

```go
func (*AuthController) Login(req *struct {
    Body *LoginDto
}, echoCtx *echo.Context) (*security.UserPrincipal, error) {
    // ...
}
```

### 参数识别硬规则

- 路径参数：
  - 路由中的 `/:ID` 会匹配请求结构体中的 `ID` 字段。
  - 字段名需要与路径参数名一致或大小写等价。
- 查询参数：
  - 请求结构体中除 `Body` 外、且未被识别为路径参数的字段，会被视为 query 参数。
- 请求体：
  - 字段名必须字面量写成 `Body`。
  - 推荐使用指针类型，例如 `Body *CreateOrderVo`。
  - Swagger `requestBody` 也依赖该约定生成。
- `*echo.Context`：
  - 可作为 handler 参数之一，由框架自动注入（echo v5 中 `Context` 是结构体，handler 必须用指针）。
- `security.AuthCtx`：
  - 通过前置的 `security.RequireAuth(...)` 生成并注入。

### 标签规范

优先补齐以下 tag：

- `json`
  - 用于请求绑定和 Swagger 字段名展示。
- `validate`
  - 用于参数校验；框架会自动执行校验并返回翻译后的错误信息。
- `jsonschema_description`（推荐）/ `desc`（兼容写法）
  - 都用于 Swagger 参数或请求体描述。
  - 框架在生成 Swagger 时会优先读取 `jsonschema_description`，便于与 `invopop/jsonschema` 等 JSON Schema 生成器复用同一份字段说明；找不到时回退到 `desc`。
  - 同一字段两个 tag 同时存在时，以 `jsonschema_description` 为准。

示例：

```go
type CreateOrderVo struct {
    Title  string `json:"title" validate:"required" jsonschema_description:"订单标题"`
    Amount int    `json:"amount" validate:"required,gte=1" jsonschema_description:"订单金额"`
}
```

旧项目里仍可继续使用 `desc:"..."`，无需立即改写。

## 鉴权规范

默认受保护接口写法：

```go
g.POST("", security.RequireAuth(), c.Create, echoswg.WithOperationId("createOrder"))
```

需要角色时：

```go
g.POST("/approve", security.RequireAuth(security.RoleFinance), c.Approve, echoswg.WithOperationId("approveOrder"))
```

规则如下：

- 认证统一走 `security.RequireAuth(...)`。
- 当前用户从 `security.AuthCtx` 读取，不要在 controller 中重复解析 token。
- `RoleAnonymous` 表示显式允许匿名，不等于已登录。
- 如果是公开接口，可以直接不加鉴权，或按现有项目规范显式使用匿名角色；生成代码前先观察当前项目已有写法。

## Controller 书写规范

- 保持 controller 轻量，只做以下事情：
  - 定义路由
  - 声明鉴权
  - 接收并校验参数
  - 调用 `service`
  - 返回结果
- 不要在 controller 中直接：
  - 操作数据库
  - 操作 Redis
  - 开事务
  - 编排复杂审批/状态流
  - 复制粘贴鉴权解析逻辑
- 一个资源域通常一个 controller 文件；如接口特别多，可按资源拆分多个文件，但都遵循 `init()` 自动注册模式。

## Service 书写规范

- service 负责核心业务实现。
- 对外优先暴露包级单例：

```go
var Order = &orderService{}

type orderService struct{}
```

- 写操作优先放进事务包装中，例如 `db.WithTx(...)`。
- 需要当前登录用户时，从 `security.AuthCtx` 获取 principal，不要让 controller 额外传裸 `userId`。
- 错误优先复用统一业务错误类型，例如 `bizerrors`，以保持全局 HTTP 返回格式一致。
- DTO/VO 命名保持稳定：
  - `CreateXxxVo`
  - `UpdateXxxVo`
  - `XxxQueryVo`
  - `XxxVo`
  - `XxxDetailVo`

## Main 与框架装配规范

如果要新建一个基于本框架的应用入口，按以下顺序组织：

1. 初始化配置。
2. 初始化日志、Redis、数据库等基础设施。
3. 进行自动迁移或其他启动检查。
4. 匿名导入 `controller` 包，触发所有路由注册。
5. 获取全局 Echo 实例并挂载通用中间件。
6. 调用 `echoswg.ServeSwagger(...)` 暴露 Swagger UI 和 API docs。
7. 启动 HTTP 服务。

不要在 `main` 里手动逐条调用 `e.GET(...)` 或 `e.POST(...)` 注册业务接口，除非当前项目明确不再遵循此模式。

## Main 示例代码

下面的模板来自本 skill 内置骨架和 `example-skeleton` 实体模板，可作为新项目的直接起点。复制后至少要替换模块路径、项目标题、配置结构和中间件细节。

```go
package main

import (
    "net/http"
    "os"

    "your/module/config"
    _ "your/module/controller"
    "your/module/db"
    "your/module/util"

    "github.com/labstack/echo/v5/middleware"
    "github.com/yb7/alilog"
    "github.com/yb7/echoswg"
)

func main() {
    if err := os.Setenv("TZ", "Asia/Shanghai"); err != nil {
        alilog.Fatal(err)
    }
    config.EnsureInitSuccess()

    alilog.SetConfig(config.C.Sls)
    alilog.StartSlsLog()
    defer alilog.CloseSlsLog(10000)

    db.InitRueidisClient()

    db.OpenDB()
    defer db.CloseDB()

    db.DbMigrate()

    e := util.EchoInstance

    echoswg.ServeSwagger(e, echoswg.SwaggerConfig{
        UrlPrefix:   "/api",
        Title:       "Your Project API",
        Description: "Your Project API",
        CdnPrefix:   "https://statics.stock001.com/swagger-ui-5.32.5",
    })

    e.Use(middleware.RequestLogger())
    e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 9}))
    e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
    }))

    e.Use(util.EchoRecover)

    alilog.Infof("rest server started at port [%s]", config.C.Ports.Http)
    if err := e.Start(config.C.Ports.Http); err != nil {
        alilog.Fatal(err)
    }
}
```

使用该模板时重点检查：

- 是否匿名导入了 `controller` 包。
- `util.EchoInstance` 是否已经初始化并挂上全局错误处理。
- `SwaggerConfig` 的 `Title`、`Description`、`UrlPrefix` 是否符合新项目。
- `db.InitRueidisClient()`、`db.OpenDB()`、`db.DbMigrate()` 是否与新项目基础设施一致。
- `middleware.RequestLogger`、CORS 策略、Recover 实现是否需要调整。
- echo v5 已移除 `middleware.Logger()` 与 `e.Logger.Fatal(...)`，请使用 `middleware.RequestLogger()` 与 `if err := e.Start(...); err != nil { alilog.Fatal(err) }`。

## Security 可复制模板

`example-skeleton/security` 可以直接作为新项目的初始模板，但复制前必须统一替换模块名、Redis key 前缀、Cookie 名、角色名和内置系统用户信息。

### authctx.go 模板

用于定义统一的 `AuthCtx` 接口、全局系统上下文和用户 principal 注入方式。

```go
package security

import (
    "context"
    "time"
)

type AuthCtx interface {
    context.Context
    GetUserPrincipal() *UserPrincipal
    WithValue(key, val any)
}

var GlobalCtx AuthCtx

func init() {
    globalCtx := context.WithValue(context.Background(), KEY_USER_PRINCIPAL, &UserPrincipal{
        UserID:      0,
        Name:        "SYSTEM_ROOT",
        Roles:       make([]string, 0),
        PhoneNumber: 18900000000,
        AccessToken: "",
    })
    GlobalCtx = &authCtxImpl{
        Context: globalCtx,
    }
}

type authCtxImpl struct {
    context.Context
}

const KEY_USER_PRINCIPAL = "user_principal"

func (ctx *authCtxImpl) GetUserPrincipal() *UserPrincipal {
    userPrincipal, ok := ctx.Value(KEY_USER_PRINCIPAL).(*UserPrincipal)
    if !ok {
        return nil
    }
    return userPrincipal
}

func (ctx *authCtxImpl) WithValue(key, val any) {
    ctx.Context = context.WithValue(ctx.Context, key, val)
}

func (ctx *authCtxImpl) Deadline() (deadline time.Time, ok bool) {
    return ctx.Context.Deadline()
}

func (ctx *authCtxImpl) Done() <-chan struct{} {
    return ctx.Context.Done()
}

func (ctx *authCtxImpl) Err() error {
    return ctx.Context.Err()
}

func (ctx *authCtxImpl) Value(key any) any {
    return ctx.Context.Value(key)
}
```

### roles.go 模板

用于定义角色常量。新项目应只保留实际需要的角色，不要无脑复制旧项目的业务角色。

```go
package security

type Role string

const (
    RoleSystemAdmin Role = "SystemAdmin"
    RoleAnonymous  Role = "user:anonymous"
)
```

### auth.go 模板

用于从请求中提取 token、校验 Redis 中的登录态、生成 `AuthCtx` 并校验角色。下面保留了当前骨架的核心写法，适合作为新项目初版。

```go
package security

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"

    "your/module/bizerrors"
    "your/module/db"

    "github.com/labstack/echo/v5"
    "github.com/redis/rueidis"
    "github.com/yb7/alilog"
)

const COOKIE_ACCESSTOKEN = "flashnews_access_token"

func getAccessTokenInRequest(ctx *echo.Context) string {
    authorization := ctx.Request().Header.Get("Authorization")
    if len(authorization) != 0 {
        if !strings.HasPrefix(authorization, "Bearer ") {
            return ""
        }
        return strings.TrimPrefix(authorization, "Bearer ")
    }
    cookie, err := ctx.Cookie(COOKIE_ACCESSTOKEN)
    if err == nil && cookie != nil {
        return cookie.Value
    }
    return ""
}

func RequireAuth(requiredRoles ...Role) func(ctx *echo.Context) (AuthCtx, error) {
    return func(ctx *echo.Context) (AuthCtx, error) {
        if len(requiredRoles) == 1 && requiredRoles[0] == RoleAnonymous {
            return &authCtxImpl{
                Context: ctx.Request().Context(),
            }, nil
        }

        accessToken := getAccessTokenInRequest(ctx)
        if len(accessToken) == 0 {
            return nil, bizerrors.MissingAccessToken
        }

        userPrincipal, err := getUserPrincipalByAccessToken(GlobalCtx, accessToken)
        if err != nil {
            return nil, bizerrors.Unauthorized
        }

        missingRoles := make([]string, 0)
        for _, requiredRole := range requiredRoles {
            _, ok := userPrincipal.FindRoleWithPrefix(string(requiredRole))
            if !ok {
                missingRoles = append(missingRoles, string(requiredRole))
            }
        }
        if len(missingRoles) > 0 {
            return nil, bizerrors.MissingPermissions(missingRoles...)
        }

        log := alilog.WithTraceId(
            fmt.Sprintf("requestUserId=%d", userPrincipal.UserID),
            fmt.Sprintf("requestUserPhone=%d", userPrincipal.PhoneNumber),
        )
        authCtx := log.CreateLogContext(context.Background())
        authCtx = context.WithValue(authCtx, KEY_USER_PRINCIPAL, userPrincipal)

        return &authCtxImpl{
            Context: authCtx,
        }, nil
    }
}

func AccessTokenKeyInRedis(token string) (string, error) {
    parts := strings.Split(token, "@")
    if len(parts) < 2 {
        return "", alilog.Errorf("bad access token format: %s", token)
    }
    return fmt.Sprintf("flashnews:access-token:%s", parts[0]), nil
}

type UserPrincipal struct {
    UserID      int      `json:"userId"`
    PhoneNumber int64    `json:"phoneNumber"`
    Department  string   `json:"department"`
    Name        string   `json:"name"`
    Character   string   `json:"character"`
    Roles       []string `json:"roles"`
    AccessToken string   `json:"accessToken"`
}

func (user UserPrincipal) FindRole(role Role) (Role, bool) {
    r, ok := user.FindRoleWithPrefix(string(role))
    return Role(r), ok
}

func (user UserPrincipal) FindRoleWithPrefix(rolePrefix string) (string, bool) {
    for _, role := range user.Roles {
        if strings.HasPrefix(role, rolePrefix) {
            return role, true
        }
    }
    return "", false
}

func getUserPrincipalByAccessToken(ctx context.Context, accessToken string) (*UserPrincipal, error) {
    tokenKey, err := AccessTokenKeyInRedis(accessToken)
    if err != nil {
        return nil, err
    }

    cmd := db.RedisClient.B().Get().Key(tokenKey).Build()
    dat, err := db.RedisClient.Do(ctx, cmd).AsBytes()
    if err != nil {
        if rueidis.IsRedisNil(err) {
            return nil, bizerrors.Unauthorized
        }
        return nil, bizerrors.Unauthorized
    }
    if len(dat) == 0 {
        return nil, bizerrors.Unauthorized
    }

    userPrincipal := &UserPrincipal{}
    if err := json.Unmarshal(dat, &userPrincipal); err != nil {
        return nil, alilog.Error(err)
    }
    if userPrincipal.AccessToken != accessToken {
        return nil, bizerrors.Unauthorized
    }
    return userPrincipal, nil
}
```

复制 `security` 模板时必须替换：

- `COOKIE_ACCESSTOKEN` 的名称。
- `AccessTokenKeyInRedis()` 的 Redis key 前缀。
- `UserPrincipal` 中与业务无关或不适用的字段。
- `RoleSystemAdmin`、`RoleAnonymous` 之外的项目角色定义。
- `bizerrors` 中缺失的错误类型。
- `getUserPrincipalByAccessToken()` 里对 Redis 的具体依赖和序列化结构。

如果新项目要支持：

- JWT：保留 `RequireAuth(...)` 的返回约定，但把 token 校验逻辑改成验签和 claims 解析。
- 多租户：在 `AuthCtx` 中增加 tenant 信息，并在 principal 或 context 中写入租户字段。
- 更严格 RBAC：保留 controller 侧 `RequireAuth(role...)` 的用法，重写 `FindRoleWithPrefix` 或角色判断策略。

## 完整新项目骨架模板

当用户要求“按 echoswg 新建一个项目”或“直接起一套可运行 API 骨架”时，优先按下面的目录和文件模板生成。该骨架追求最小可运行、职责清晰、便于后续扩展。

### 建议目录

```text
your-project/
  bizerrors/
    bizerrors.go
  config/
    config.go
  controller/
    demo.go
  db/
    db.go
  security/
    auth.go
    authctx.go
    roles.go
  service/
    demo.go
  util/
    echo_common.go
    echo_recover.go
    shared.go
  main.go
  go.mod
  config-template.toml
```

如果项目尚未实现 `bizerrors`、`echo_recover.go`、`shared.go` 等公共模块，生成代码时要一并补齐最小版本，否则 `main.go` 和错误处理链无法闭环。

### 1. config/config.go

目标：

- 提供全局配置对象 `config.C`
- 在启动时自动读取配置文件和环境变量
- 校验核心配置是否完整

最小模板：

```go
package config

import (
    "fmt"
    "reflect"
    "strings"

    "github.com/go-playground/validator/v10"
    "github.com/spf13/viper"
)

var C = &Configuration{
    Redis: &RedisConfig{
        Database: 0,
    },
}

type Configuration struct {
    Ports *PortsConfig `validate:"required"`
    Db    *DbConfig    `validate:"required"`
    Redis *RedisConfig `validate:"required"`
}

type PortsConfig struct {
    Http string `validate:"required"`
}

type DbConfig struct {
    Url string `validate:"required"`
}

type RedisConfig struct {
    Host     string `validate:"required"`
    Password string
    Database int `validate:"gte=0,lte=16"`
}

var configInitError error

func EnsureInitSuccess() {
    if configInitError != nil || C == nil {
        panic(fmt.Sprintf("config init failed: %v", configInitError))
    }
}

func init() {
    viper.SetConfigName("config")
    viper.SetConfigType("toml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("/etc/app/")
    viper.AddConfigPath("$HOME/.app")

    bindAllKeys(reflect.TypeOf(*C))

    if err := viper.ReadInConfig(); err != nil {
        configInitError = err
        return
    }
    if err := viper.Unmarshal(C); err != nil {
        configInitError = err
        C = nil
        return
    }

    validate := validator.New()
    if err := validate.Struct(C); err != nil {
        configInitError = err
        C = nil
        return
    }
}

func bindAllKeys(t reflect.Type) {
    for i := 0; i < t.NumField(); i++ {
        f := t.Field(i)
        walkThroughFields(envKeyName(f), f.Type)
    }
}

func envKeyName(f reflect.StructField) string {
    envKey, ok := f.Tag.Lookup("env")
    if ok {
        return envKey
    }
    return strings.ToLower(f.Name[0:1]) + f.Name[1:]
}

func walkThroughFields(fieldKey string, t reflect.Type) {
    if t.Kind() == reflect.Ptr {
        walkThroughFields(fieldKey, t.Elem())
        return
    }
    if t.Kind() == reflect.Struct {
        for i := 0; i < t.NumField(); i++ {
            f := t.Field(i)
            walkThroughFields(fieldKey+"."+envKeyName(f), f.Type)
        }
        return
    }
    alterKey := strings.ReplaceAll(fieldKey, ".", "_")
    upperAlterKey := strings.ToUpper(alterKey)
    _ = viper.BindEnv(fieldKey, fieldKey, alterKey, upperAlterKey)
}
```

### 2. db/db.go

目标：

- 统一初始化数据库客户端
- 提供关闭数据库和事务包装能力
- 让 service 层可直接复用

最小模板：

```go
package db

import (
    "context"
    "database/sql"

    "your/module/config"
    "your/module/ent"

    "entgo.io/ent/dialect"
    entsql "entgo.io/ent/dialect/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/yb7/alilog"
)

var DBClient *ent.Client

func OpenDB() {
    db, err := sql.Open("pgx", config.C.Db.Url)
    if err != nil {
        alilog.Fatal(err)
    }
    drv := entsql.OpenDB(dialect.Postgres, db)
    DBClient = ent.NewClient(ent.Driver(drv))
}

func DbMigrate() {
    if DBClient == nil {
        alilog.Fatal("db client is nil")
    }
}

func CloseDB() {
    if DBClient != nil {
        DBClient.Close()
    }
}

func WithTx(ctx context.Context, client *ent.Client, fn func(ctxInTx context.Context, tx *ent.Tx) error) error {
    tx, err := client.Tx(ctx)
    if err != nil {
        return alilog.Errorf("start transaction err: %v", err)
    }
    defer func() {
        if v := recover(); v != nil {
            _ = tx.Rollback()
            panic(v)
        }
    }()

    if err := fn(ctx, tx); err != nil {
        if rerr := tx.Rollback(); rerr != nil {
            return alilog.Errorf("rollback transaction err: %v", rerr)
        }
        return err
    }
    if err := tx.Commit(); err != nil {
        return alilog.Errorf("commit transaction err: %v", err)
    }
    return nil
}
```

说明：

- 如果项目暂时未接入 `ent`，可以先保留 `OpenDB` 和 `CloseDB`，并将 `WithTx` 改为适配当前 ORM。
- 如果项目没有 Redis，也可以先不生成 Redis 初始化代码。

### 3. util/echo_common.go

目标：

- 统一 Echo 的错误输出格式
- 让 controller 和 service 只需返回 error，不需要自己拼 HTTP JSON

最小模板：

```go
package util

import (
    "net/http"

    "your/module/bizerrors"

    "github.com/labstack/echo/v5"
)

var EchoInstance = echo.New()

func init() {
    EchoInstance.HTTPErrorHandler = func(c *echo.Context, err error) {
        if resp, _ := echo.UnwrapResponse(c.Response()); resp != nil && resp.Committed {
            return
        }

        var bizErr *bizerrors.BizError
        switch err := err.(type) {
        case *echo.HTTPError:
            bizErr = &bizerrors.BizError{
                HttpStatus:   err.Code,
                Success:      false,
                ErrorCode:    "HTTP_ERROR",
                ErrorMessage: err.Error(),
                ShowType:     2,
            }
        case *bizerrors.BizError:
            bizErr = err
        default:
            bizErr = &bizerrors.BizError{
                HttpStatus:   http.StatusInternalServerError,
                Success:      false,
                ErrorCode:    "SYS_ERROR",
                ErrorMessage: err.Error(),
                ShowType:     2,
            }
        }
        _ = c.JSON(bizErr.HttpStatus, bizErr)
    }
}
```

说明：

- 若项目暂无统一错误结构，生成代码时要补最小版 `bizerrors.BizError`。
- `EchoInstance` 放在 `util` 中，便于 `main` 和 `controller` 共用。

### 4. bizerrors/bizerrors.go

目标：

- 提供统一业务错误结构
- 让 `util/echo_common.go` 和 `util/echo_recover.go` 可以直接复用
- 让 controller/service 只需要返回 error，而不用自己拼 HTTP 响应

最小模板：

```go
package bizerrors

import "net/http"

type BizError struct {
    HttpStatus   int    `json:"-"`
    Success      bool   `json:"success"`
    ErrorCode    string `json:"errorCode"`
    ErrorMessage string `json:"errorMessage"`
    ShowType     int    `json:"showType"`
}

func (e *BizError) Error() string {
    return e.ErrorMessage
}

func (e *BizError) Code() string {
    return e.ErrorCode
}

func BadRequest(msg string) *BizError {
    return &BizError{
        HttpStatus:   http.StatusBadRequest,
        Success:      false,
        ErrorCode:    "BAD_REQUEST",
        ErrorMessage: msg,
        ShowType:     2,
    }
}

var Unauthorized = &BizError{
    HttpStatus:   http.StatusUnauthorized,
    Success:      false,
    ErrorCode:    "UNAUTHORIZED",
    ErrorMessage: "unauthorized",
    ShowType:     2,
}

var MissingAccessToken = &BizError{
    HttpStatus:   http.StatusUnauthorized,
    Success:      false,
    ErrorCode:    "MISSING_ACCESS_TOKEN",
    ErrorMessage: "missing access token",
    ShowType:     2,
}

func MissingPermissions(roles ...string) *BizError {
    return &BizError{
        HttpStatus:   http.StatusForbidden,
        Success:      false,
        ErrorCode:    "MISSING_PERMISSIONS",
        ErrorMessage: "missing permissions",
        ShowType:     2,
    }
}
```

说明：

- 如果项目已有统一错误协议，优先复用现有格式，不要强行引入新字段。
- 若前端约定了 `errno`/`msg` 等字段名，生成时要同步调整 `BizError`。

### 5. util/shared.go

目标：

- 提供全局共享的 Echo 实例
- 让 `main`、`controller`、`echo_common` 使用同一个 `echo.New()`

最小模板：

```go
package util

import "github.com/labstack/echo/v5"

var EchoInstance = echo.New()
```

说明：

- 若已经在 `util/echo_common.go` 中声明了 `EchoInstance`，则不要重复定义，保留一个唯一实现即可。
- 生成项目时，`shared.go` 和 `echo_common.go` 二选一存放 `EchoInstance`，但 skill 默认推荐放在 `shared.go`，把错误处理放在 `echo_common.go`。

### 6. util/echo_recover.go

目标：

- 提供兜底 panic 恢复中间件
- 统一打印请求和堆栈信息
- 与 `bizerrors.BizError` 联动

最小模板：

```go
package util

import (
    "fmt"
    "net/http/httputil"
    "runtime"

    "your/module/bizerrors"

    "github.com/labstack/echo/v5"
    "github.com/yb7/alilog"
)

var StackSize = 4 << 10

func EchoRecover(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c *echo.Context) (err error) {
        defer func() {
            if r := recover(); r != nil {
                if bizError, ok := r.(*bizerrors.BizError); ok {
                    err = c.JSON(bizError.HttpStatus, bizError)
                    return
                }

                recovered, ok := r.(error)
                if !ok {
                    recovered = fmt.Errorf("%v", r)
                }

                stack := make([]byte, StackSize)
                length := runtime.Stack(stack, true)
                reqDump, _ := httputil.DumpRequest(c.Request(), true)

                alilog.Errorf("[PANIC RECOVER] Request\n%s", string(reqDump))
                alilog.Errorf("[PANIC RECOVER] %v\n%s", recovered, string(stack[:length]))

                err = recovered
            }
        }()
        return next(c)
    }
}
```

说明：

- 如果项目已经使用 Echo 官方 `middleware.Recover()`，也可以先复用官方版本；但当你需要统一输出业务错误和打印请求体时，优先使用该模板。
- 若项目对日志敏感，请注意脱敏请求头和请求体中的敏感信息。
- echo v5 移除了 `c.Error(...)`：恢复后必须把 `error` 通过中间件返回值抛回给框架，由 `EchoInstance.HTTPErrorHandler` 统一渲染响应。

### 7. service/demo.go

目标：

- 提供一个最小业务服务模板
- 体现“service 负责业务逻辑、controller 只转调”的约定

最小模板：

```go
package service

import "your/module/security"

var Demo = &demoService{}

type demoService struct{}

type CreateDemoVo struct {
    Name string `json:"name" validate:"required" jsonschema_description:"演示名称"`
}

type DemoVo struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func (*demoService) Create(ctx security.AuthCtx, req *CreateDemoVo) (*DemoVo, error) {
    user := ctx.GetUserPrincipal()
    _ = user

    return &DemoVo{
        ID:   1,
        Name: req.Name,
    }, nil
}
```

说明：

- 这个文件故意保持无数据库依赖，适合项目初始化阶段先跑通接口链路。
- 真正接入数据库时，再将 `Create` 迁移为事务写入逻辑。

### 8. controller/demo.go

目标：

- 展示 `init()` 自动注册路由
- 展示 `Body` 请求体规则和 `RequireAuth()` 的标准接法

最小模板：

```go
package controller

import (
    "your/module/security"
    "your/module/service"
    "your/module/util"

    "github.com/yb7/echoswg"
)

type DemoController struct{}

func init() {
    c := new(DemoController)
    g := echoswg.NewApiGroup(util.EchoInstance, "Demo", "/api/demos")
    g.POST("", security.RequireAuth(), c.Create, echoswg.WithOperationId("createDemo"))
}

func (*DemoController) Create(ctx security.AuthCtx, req *struct {
    Body *service.CreateDemoVo `jsonschema_description:"创建Demo"`
}) (*service.DemoVo, error) {
    return service.Demo.Create(ctx, req.Body)
}
```

说明：

- 如果需要匿名接口，可以改为 `security.RequireAuth(security.RoleAnonymous)` 或按当前项目规范直接不加鉴权。
- 如果需要 query/path 参数，再在请求 struct 中增加字段，但请求体仍必须叫 `Body`。

### 9. main.go

目标：

- 按本框架模式把所有模块串起来
- 保持足够简洁，便于 AI 在新项目上直接复制

最小模板：

```go
package main

import (
    "net/http"
    "os"

    "your/module/config"
    _ "your/module/controller"
    "your/module/db"
    "your/module/util"

    "github.com/labstack/echo/v5/middleware"
    "github.com/yb7/alilog"
    "github.com/yb7/echoswg"
)

func main() {
    if err := os.Setenv("TZ", "Asia/Shanghai"); err != nil {
        alilog.Fatal(err)
    }

    config.EnsureInitSuccess()

    db.OpenDB()
    defer db.CloseDB()

    db.DbMigrate()

    e := util.EchoInstance

    echoswg.ServeSwagger(e, echoswg.SwaggerConfig{
        UrlPrefix:   "/api",
        Title:       "Your Project API",
        Description: "Your Project API",
        CdnPrefix:   "https://img.cls.cn/statics/swagger-ui-4.10.3",
    })

    e.Use(middleware.RequestLogger())
    e.Use(middleware.Gzip())
    e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
    }))
    e.Use(util.EchoRecover)

    alilog.Infof("rest server started at port [%s]", config.C.Ports.Http)
    if err := e.Start(config.C.Ports.Http); err != nil {
        alilog.Fatal(err)
    }
}
```

说明：

- 如果项目已接入 Redis、SLS、Recover 中间件，可按前文的 `main` 模板增强。
- `controller` 必须匿名导入，否则 `init()` 不会执行，路由不会注册。

### 10. go.mod 模板

为了让 AI 可以直接初始化项目，建议在骨架里同时给出最小 `go.mod` 模板。版本号不必和示例完全一致，但依赖集合应满足最小可运行需求。

```go
module your/module

go 1.24.0

require (
    github.com/go-playground/validator/v10 v10.29.0
    github.com/jackc/pgx/v5 v5.4.1
    github.com/labstack/echo/v5 v5.1.1
    github.com/redis/rueidis v1.0.69
    github.com/spf13/viper v1.19.0
    github.com/yb7/alilog v1.1.12
    github.com/yb7/echoswg v0.7.3
)
```

说明：

- 如果项目当前就是 `echoswg` 仓库内的示例应用，模块名必须替换为实际模块路径。
- 如果项目未使用 Redis，可先删除 `github.com/redis/rueidis`。
- 如果项目未使用数据库，可先删除 `pgx`；但一旦保留 `db/db.go` 中的 PostgreSQL 模板，就需要它。
- 如果项目后续使用 `ent`，记得补充 `entgo.io/ent` 和生成代码依赖。

### 11. config-template.toml

为了让项目可以开箱即跑，建议同时生成最小配置模板：

```toml
[ports]
http = ":8080"

[db]
url = "postgres://postgres:postgres@127.0.0.1:5432/app?sslmode=disable"

[redis]
host = "127.0.0.1:6379"
password = ""
database = 0
```

### 骨架生成规则

当 AI 按该骨架起项目时，默认执行以下规则：

- 优先生成“能编译、能启动、能打开 Swagger”的最小版本。
- 如果用户没有要求接真实数据库表，`service/demo.go` 先返回 mock 结果。
- 如果项目未提供 `bizerrors`，应一并补一个最小错误结构。
- 如果项目未提供 `util/shared.go` 和 `util/echo_recover.go`，应一并补齐并在 `main.go` 中接线。
- 如果项目未提供 `ent schema`，不要强行生成复杂迁移逻辑。
- 如果用户只说“先起个骨架”，默认生成上述全部文件。
- 如果用户已经有部分目录，只补缺失文件，并尽量复用现有命名。

### 骨架生成后的自检

生成完整骨架后，至少检查：

- `main.go` 是否匿名导入了 `controller`
- `controller/demo.go` 是否使用 `init()` 注册路由
- `service/demo.go` 是否暴露了包级单例
- `util/shared.go` 是否提供了唯一的 `EchoInstance`
- `util/echo_common.go` 是否设置了 `EchoInstance.HTTPErrorHandler`
- `util/echo_recover.go` 是否已在 `main.go` 中挂载
- `bizerrors/bizerrors.go` 是否可被 `security` 和 `util` 共同复用
- Swagger 是否通过 `echoswg.ServeSwagger(...)` 挂载
- 请求体字段是否使用 `Body`
- `WithOperationId(...)` 是否已填写
- `go.mod` 依赖是否覆盖 Echo、Viper、alilog、echoswg
- 配置结构与 `config-template.toml` 是否一致

### AI 执行偏好

当用户说“按 echoswg 起个新项目”时，优先：

1. 生成本章节中的骨架文件。
2. 先让项目具备最小可运行能力。
3. 再根据用户业务补充具体资源、DTO、service 和鉴权。
4. 始终保持 controller 薄、service 厚、Swagger 自动生成。

## 新增接口时的执行清单

当用户要求“新增一个接口”时，按下面顺序执行：

1. 确认资源域、HTTP 方法、路由路径、是否需要鉴权、返回模型。
2. 在 `service` 定义或补充所需 DTO/VO 和业务方法。
3. 在对应 `controller` 中补充 handler。
4. 在该 controller 的 `init()` 中注册路由。
5. 为路由补充唯一的 `WithOperationId(...)`。
6. 如果涉及请求体，使用 `Body` 字段承载。
7. 如果涉及路径参数，保证结构体字段名与 `/:Param` 对应。
8. 如果涉及当前用户，使用 `security.RequireAuth(...)` + `security.AuthCtx`。
9. 保持 controller 薄、service 厚。
10. 完成后检查 Swagger 是否可自动覆盖该接口，不要手写文档。

## 输出代码时的默认要求

在使用本 skill 生成代码时，默认遵循：

- 语言使用 Go。
- 尽量复用本 skill 和 `example-skeleton` 中的命名与组织方式。
- 优先生成最小改动，而不是引入新的框架层。
- 如果项目中已有同类 controller/service，先模仿最近的现有实现。
- 若发现当前项目写法与骨架模板不一致，以当前项目现有主流写法为准，并在结果中简要说明偏差。

## 禁止事项

- 不要手写 Swagger JSON。
- 不要把业务核心堆进 controller。
- 不要绕开 `security.RequireAuth(...)` 自己复制 token 解析流程。
- 不要把请求体字段命名成 `Payload`、`Data` 等其他名称来替代 `Body`。
- 不要在路径参数是 `/:ID` 时，把结构体字段写成不匹配的 `OrderID`，除非你同时调整路由参数名。
- 不要在 `main` 中手工散落注册业务路由。

## 生成代码前的自检

在真正输出代码前，先快速检查：

- 是否放在正确目录：`controller` / `service` / `security` / `main`
- 是否使用 `init()` + `NewApiGroup(...)` 注册路由
- 是否提供唯一 `WithOperationId(...)`
- 是否正确区分 path/query/body
- 请求体字段是否命名为 `Body`
- 是否正确接入 `security.RequireAuth(...)`
- 是否把业务逻辑放在 `service`
- 是否沿用现有 DTO/VO 命名风格

## 可直接复用的最小模板

```go
package controller

import (
    "your/module/security"
    "your/module/service"
    "your/module/util"

    "github.com/yb7/echoswg"
)

type DemoController struct{}

func init() {
    c := new(DemoController)
    g := echoswg.NewApiGroup(util.EchoInstance, "Demo", "/api/demos")
    g.POST("", security.RequireAuth(), c.Create, echoswg.WithOperationId("createDemo"))
}

func (*DemoController) Create(ctx security.AuthCtx, req *struct {
    Body *service.CreateDemoVo `jsonschema_description:"创建Demo"`
}) (*service.DemoVo, error) {
    return service.Demo.Create(ctx, req.Body)
}
```

当用户没有给出更具体约束时，优先按这个模板衍生实现。
