# QesyWeb Go Web 框架

这是一个面向业务快速开发的 Go Web 框架示例项目。项目使用一个简单新闻系统展示框架的常用能力：QesySQL 单表 CRUD、MySQL、Redis、HTML 模板、日志、定时任务、Worker，以及 Protobuf WebSocket。

## 环境要求

- Go 1.26 或更高版本
- MySQL 8.x（启用新闻示例时需要）
- Redis（启用缓存示例时需要）
- `protoc` 和 `protoc-gen-go`（重新生成 Protobuf 代码时需要）

## 快速启动

1. 复制 `.env.example` 为 `.env`，再填写自己的配置。注意文件名是 `.env`，不是 `.evn`。
2. 默认 `MYSQL_ENABLED=false`、`REDIS_ENABLED=false`，可以先直接查看框架首页。
3. 运行：

```bash
go run ./cmd/web
```

打开 `http://localhost:8080`，健康检查地址是 `http://localhost:8080/health`。

## 启用新闻系统

在 `.env` 中填写 MySQL 配置，并设置：

```text
MYSQL_ENABLED=true
REDIS_ENABLED=true
```

先执行迁移：

```bash
go run ./cmd/migrate
```

再启动 Web 服务：

```bash
go run ./cmd/web
```

前台新闻地址：`http://localhost:8080/news`，只展示已发布新闻；后台新闻管理地址：`http://localhost:8080/admin/news`。

## 后台管理员登录

管理员用户名、密码和邮箱都保存在 `admin_users` 数据库表中，不放在 `.env`。执行以下命令后按提示输入；密码输入时不会回显，密码会以 bcrypt 哈希保存：

```bash
go run ./cmd/seedadmin
```

后台登录地址：`http://localhost:8080/admin/login`。登录页包含图片验证码，登录成功后才能访问新闻管理。Session 优先保存到 Redis；Redis 关闭时仅使用进程内存，适合本地演示，不适合多实例生产环境。

## 邮件发送示例

项目使用 Resend HTTP API 发送邮件，在 `.env` 中填写：

```text
RESEND_ENABLED=true
RESEND_API_KEY=re_xxxxxxxxx
RESEND_FROM=QesyWeb <noreply@example.com>
```

登录后台后，点击“发送 Resend 测试邮件”即可验证邮件配置。邮件通过 Resend API 发送，封装在 `internal/mail` 中；管理员收件邮箱来自数据库中的管理员记录。
如果没有设置 `RESEND_ENABLED=true`，系统会明确返回“邮件发送未启用”，不会把未发送误报为成功。

## 多语言示例

语言包放在 `locales/` 下，先按语言分目录，再按页面拆分 JSON；每个语言目录都有一个通用的 `common.json`，页面文案放在对应页面文件中：

```text
locales/
├── zh-CN/
│   ├── common.json
│   ├── admin.json
│   └── news.json
└── en-US/
    ├── common.json
    ├── admin.json
    └── news.json
```

语言选择顺序为：URL 的 `lang` 参数、语言 Cookie、`Accept-Language` 请求头、默认英文（`en-US`）。`go run ./cmd/seedadmin` 的管理员初始化提示也使用英文。新增业务时，在每种语言目录中增加同名页面文件，Key 使用业务前缀。

```text
/admin/login?lang=en-US
/admin/login?lang=zh-CN
```

模板通过 `{{call .T "admin.login.title"}}` 读取语言包，业务代码不直接写死页面文案。

## 代码应该放在哪里

- `internal/news/frontend`：前台新闻列表和详情 Handler，只展示已发布内容。
- `internal/news/handler`：后台新闻管理 Handler，负责 CRUD 和表单。
- `handler`：处理 HTTP 参数、校验和模板，不直接访问数据库。
- `service`：处理业务规则、事务边界和缓存失效。
- `repository`：处理数据库持久化。普通单表操作使用 QesySQL 链式 API。
- `model`：定义实体、请求结构和页面数据结构。
- `internal/cache`：Redis 连接和 JSON 缓存辅助。
- `internal/logging`：按天日志文件和日志清理。
- `internal/scheduler`：定时任务注册和停止。
- `internal/worker`：后台任务队列和 Worker。
- `internal/websocket`：连接、心跳、Protobuf 消息和房间广播。
- `internal/auth`：Session、登录态和管理员保护中间件。
- `internal/captcha`：一次性图片验证码。
- `internal/i18n`：JSON 语言包加载和请求语言识别。
- `internal/mail`：Resend 邮件发送封装。
- `templates/`：Go `html/template` 页面模板。
- `static/images/`、`static/scripts/`、`static/styles/`：图片、脚本和样式等静态资源。
- `web/`：Web 层目录约定说明；程序入口仍在 `cmd/web`。

## QesySQL 示例

新闻 Repository 中的标准单表操作采用这种风格：

```go
var items []model.News
err := db.Table("news").
    SetWhere(qesysql.Eq{"status": "published"}).
    SetSort("id DESC").
    SetLimit(20).
    SetOffset(0).
    SetField("id", "title", "status").
    ExecSelect(ctx, &items)
```

事务使用 QesySQL 提供的 `InTx`：

```go
err := db.InTx(ctx, nil, func(tx *qesysql.Tx) error {
    _, err := tx.Table("news").
        SetInsert(qesysql.Values{"title": "示例", "slug": "demo", "status": "draft"}).
        ExecInsert(ctx)
    return err
})
```

## Redis Key 集中管理与 Redis 示例

Redis Key 统一定义在 `internal/cache/keys.go`，业务代码不直接散落拼接 Key：

```go
key := cache.NewsDetail(newsID)
listKey := cache.NewsList(page, pageSize, status)
sessionKey := cache.Session(sessionID)
```

Key 应包含业务名称和版本，例如 `news:v1:detail:42`。修改 Key 格式时，只修改集中定义的方法，并同步考虑旧 Key 的兼容或清理。

```go
var item model.News
if err := cache.GetJSON(ctx, cache.NewsDetail(1), &item); err == nil {
    return item
}
_ = cache.SetJSON(ctx, cache.NewsDetail(1), item, 5*time.Minute)
```

## Worker 示例

业务耗时操作实现 `worker.Job`，再提交到 Worker 池；HTTP Handler 或定时任务只负责投递任务：

```go
pool := worker.New(2, 32)
pool.Start(ctx)
defer pool.Stop()

if err := pool.Submit(worker.NewExampleJob(logger, "刷新新闻缓存")); err != nil {
    logger.Error("任务投递失败", "error", err)
}
```

示例任务位于 `internal/worker/example.go`，支持上下文取消，并演示开始、完成日志。

## 日志文件

启动后会在 `LOG_DIR` 下生成：

- `app-YYYY-MM-DD.log`：应用、业务、数据库、Redis、任务和错误日志。
- `access-YYYY-MM-DD.log`：HTTP 请求日志。

日志按 `LOG_TIMEZONE` 切换，默认保留 30 天。

## Protobuf 代码生成

修改 `proto/game/v1/game.proto` 后执行：

```bash
protoc --go_out=. --go_opt=module=github.com/Qesy/QesyWeb proto/game/v1/game.proto
```

也可以执行：

```bash
make proto
```

WebSocket 服务端地址为 `/ws/game`，使用二进制帧传输 Protobuf，不使用 JSON 作为游戏消息格式。协议文件可通过 `/proto/game/v1/game.proto` 查看，浏览器演示页面是 `/game`，它会加载同一份 `.proto` 并演示心跳、加入房间和广播；页面依赖浏览器访问 CDN 获取 `protobufjs`。

## 静态资源

静态资源统一放在根目录 `static/`，程序通过 `/static/` 提供访问：

```text
static/
├── images/   图片
├── scripts/  JavaScript
└── styles/   CSS
```

例如 `static/styles/app.css` 的浏览器地址是 `/static/styles/app.css`。模板文件仍放在根目录 `templates/`。

## 按业务组织代码

每个业务模块按 `handler`、`service`、`repository`、`model` 组织。一个页面需要同时读取新闻和用户时，可以在该业务的 Service 中组合多个 Repository；如果为了页面查询方便需要在 `news/repository` 中增加用户查询方法，也可以保留，这是业务边界优先的取舍。不要为了追求目录“绝对不重复”而让 Handler 直接跨层访问数据库；重复的查询封装应在确认多个业务都复用后，再提取到共享模块。

新闻同时有前台和后台页面时，推荐复用 `model`、`service`、`repository`，分开访问端 Handler 和路由：前台使用 `/news`、`/news/{id}`，后台使用 `/admin/news`、`/admin/news/{id}`，后台路由统一挂载管理员 Session 中间件。

## 测试

```bash
go test ./...
```

完整开发清单见 [docs/DEVELOPMENT_PLAN.md](docs/DEVELOPMENT_PLAN.md)。
