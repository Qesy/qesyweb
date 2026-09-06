# Go Web 框架开发规划

## 目标

开发一个面向业务快速交付的可复用 Go Web 框架。框架以一个小型新闻管理系统作为完整参考实现，用真实业务代码说明各层代码应该放在哪里，以及 MySQL、Redis、模板等基础能力如何使用。

## 开发清单

说明：`[x]` 表示已完成且已验证，`[-]` 表示正在开发，`[ ]` 表示尚未开始。后续每完成一个可验收项，都必须在此处同步更新状态。

### 规划

- [x] 确定整体技术栈和业务优先的框架方向。
- [x] 确定新闻管理系统作为参考业务模块。
- [x] 确定 QesySQL 为单表 CRUD 的默认数据访问方式。
- [x] 确定 Redis 缓存、MySQL 事务、定时任务、Worker、WebSocket、Protobuf 和按天日志的设计要求。
- [x] 创建并维护本开发规划文档。

### 项目基础

- [x] 初始化 Go Module 和依赖版本。
- [x] 创建约定目录结构。
- [x] 将 Protobuf 目录命名为 `proto/`，避免 `api` 命名歧义。
- [x] 将语言包按语言目录和页面文件拆分，并为每种语言提供 `common.json`。
- [x] 创建 `templates/`、`static/images/`、`static/scripts/`、`static/styles/` 和 `web/` 目录约定。
- [x] 注册 `/proto/` 和 `/static/` 静态资源路由，并在模板中接入示例资源。
- [x] 将 `cmd/web/main.go` 的路由注册拆分到 `cmd/web/router.go`。
- [x] 将定时任务注册和生命周期拆分到 `cmd/web/scheduler.go`。
- [x] 配置加载与 `.env.example`。
- [x] 应用生命周期、依赖注入、优雅关闭和健康检查。
- [x] `slog` 日志组件、按天文件切换和日志保留清理。

### 基础设施

- [x] 接入 QesySQL/MySQL，提供连接检查和事务辅助，并已用真实 MySQL 启动迁移。
- [x] 接入 Redis，提供连接检查和缓存辅助，并已用真实 Redis 启动验证。
- [x] 增加 `internal/cache/keys.go`，集中管理带业务命名空间和版本的 Redis Key。
- [x] 配置 MySQL 迁移及初始表结构。
- [x] 配置 Chi、通用中间件、错误处理、静态资源和模板渲染。

### 新闻示例

- [x] 新闻模型和数据库迁移。
- [x] QesySQL 链式列表、分页、详情、新增、更新和删除示例，并已通过真实 MySQL 集成测试。
- [x] 新闻 Service、Handler、路由和 HTML 模板，并已通过真实数据库列表验证。
- [x] 增加前台新闻列表和详情，并将后台新闻 CRUD 调整到 `/admin/news` 前缀。
- [x] 新闻缓存、写入后缓存失效和缓存测试，并已通过真实 Redis 集成测试。
- [x] 新闻与审计记录的事务示例和测试，并已通过真实 MySQL 集成测试。

### 后台能力

- [x] `robfig/cron/v3` 调度器、时区、超时、防重入和恢复。
- [x] 内存任务队列、Worker、取消、超时和优雅停止。
- [x] 增加 `internal/worker/example.go`，提供可复制的业务 Worker 任务示例。
- [x] 定时任务投递 Worker 缓存刷新任务的完整示例。

### WebSocket 与 Protobuf

- [x] 配置 `protoc`、`protoc-gen-go` 和生成命令。
- [x] 定义游戏消息信封与心跳、房间相关 `.proto` 协议。
- [x] 实现 `/ws/game`、连接管理、鉴权、心跳和 Protobuf 编解码。
- [x] 实现消息路由、房间加入/离开和广播示例。
- [x] 提供浏览器二进制帧和 Protobuf 客户端示例。

### 质量与文档

- [x] 补充 Service 单元测试、Repository 集成测试和 Redis 集成测试。
- [x] 编写安装、配置、迁移、QesySQL、Redis、日志、任务、Worker、WebSocket 和新建模块文档。
- [x] 完成默认配置启动、健康检查、首页和日志的端到端验证。

### 后台认证扩展

- [x] 管理员表迁移和 bcrypt 密码哈希。
- [x] Redis/内存 Session、登录、退出和后台保护中间件。
- [x] 图片验证码生成、Cookie 关联和单次校验。
- [x] Resend 邮件发送封装和后台测试入口。
- [x] JSON 语言包、请求语言识别和模板翻译调用。
- [ ] 交互式执行 `go run ./cmd/seedadmin` 写入真实管理员并完成登录冒烟验证（需要人工输入密码）。

## 固定技术栈

- Go 1.26 或更高版本（QesySQL 当前版本要求 Go 1.26）
- MySQL 8.x
- Redis
- `github.com/go-chi/chi`：HTTP 路由
- Go `html/template`：服务端页面渲染
- `github.com/Qesy/QesySQL`：默认 MySQL 数据访问库
- `github.com/robfig/cron/v3`：定时任务调度
- `github.com/coder/websocket`：WebSocket 服务端连接管理
- Protocol Buffers：WebSocket 二进制消息协议

## QesySQL 使用规范

常规单表数据操作必须优先使用 QesySQL 的链式 API，覆盖以下场景：

1. 带条件的列表、排序、分页和总数统计。
2. 根据主键或业务字段查询详情。
3. 新增、选择性更新和删除。
4. 在事务中执行 CRUD。

以上场景不直接手写 SQL。只有 QesySQL 不适合自然表达的复杂多表聚合、报表或数据库特定能力，才能通过 Repository 内部受控的 `sqlx`/原生 SQL 兜底；所有 SQL 必须参数化，并在方法旁说明为什么不能使用 QesySQL。

## 预期目录结构

```text
cmd/web/                  程序入口和依赖组装
config/                   强类型、环境驱动的配置
internal/app/             应用容器和生命周期管理
internal/database/        QesySQL 初始化、迁移和事务辅助
internal/cache/           Redis 客户端和缓存辅助
internal/http/            Chi 路由、中间件、模板渲染、HTTP 错误处理
internal/logging/         slog 日志初始化、按天文件切换和脱敏
internal/admin/           管理员账号、登录和后台页面
internal/auth/            Session 与登录态保护中间件
internal/captcha/         图片验证码
internal/i18n/             JSON 语言包和语言识别
internal/mail/             Resend 邮件发送封装
internal/scheduler/       定时任务注册、调度和生命周期管理
internal/worker/          后台任务队列、Worker 和任务定义
internal/websocket/       连接管理、消息分发、房间和心跳
internal/news/handler/    HTTP 请求处理和模板渲染
internal/news/frontend/   前台新闻列表和详情处理
internal/news/service/    业务规则、缓存失效和事务边界
internal/news/repository/ 基于 QesySQL 的数据持久化
internal/news/model/      数据库实体、请求 DTO、视图 DTO
internal/shared/          错误、常量、校验和通用工具
migrations/               版本化 MySQL 迁移脚本
proto/                    Protobuf 协议定义和生成代码
templates/                html/template 布局、局部模板和页面
static/images/            图片静态资源
static/scripts/            JavaScript 静态资源
static/styles/             CSS 静态资源
web/                      Web 层目录约定说明
docs/                     开发与使用文档
```

## 分层职责

- `handler`：解析并校验 HTTP 输入，调用 Service，渲染模板或返回 HTTP 响应。不直接访问数据库。
- `service`：负责业务规则、缓存策略、跨 Repository 协作和事务边界。不放 HTTP 细节。
- `repository`：只负责数据持久化。普通单表操作必须使用 QesySQL；少量例外的原生 SQL 也只能在这一层。
- `model`：定义数据库实体、请求 DTO、视图 DTO 和分页结构。
- `config`：定义从环境变量加载的强类型配置。业务代码中不应散落读取环境变量的逻辑。
- `internal/shared/constants`：存放稳定、跨模块使用的常量。请求状态和业务状态不使用可变全局变量保存。
- MySQL/QesySQL 与 Redis 客户端只在启动时创建一次，再通过构造函数注入到需要它们的组件中，避免包级可变全局变量。
- `scheduler`：只负责按计划触发任务；任务本身实现明确的 `Run(context.Context)`，不将业务代码直接写入注册函数。
- `worker`：负责消费后台任务、处理超时和关闭信号；HTTP Handler 只投递任务，不在请求协程中执行耗时工作。
- `websocket`：负责连接生命周期、认证、心跳、二进制消息读取、Protobuf 编解码、消息路由和房间广播；游戏规则仍放在独立 Service 中。

## 框架能力范围

1. 启动基础：配置加载、依赖注入、日志和优雅关闭。
2. MySQL：QesySQL 初始化、连接检查、迁移、事务辅助和 Repository 编写规范。
3. Redis：连接检查、集中 Key 定义、命名空间规范、读取/写入/删除辅助和 TTL 策略。
4. HTTP：Chi 路由分组、健康检查、异常恢复、请求日志、请求 ID、静态资源、模板渲染和错误处理。
5. 模板：布局、局部模板、页面模板和明确的视图模型。
6. 日志：基于标准库 `log` 的文本日志、按天文件切换、访问日志、保留期限和敏感字段脱敏。
7. 定时任务：集中注册、时区设置、任务超时、异常恢复、同一任务防重入、优雅停止和执行日志。
8. Worker：任务接口、内存队列示例、固定数量 Worker、取消、超时、重试策略接口和优雅停止。
9. WebSocket：基于二进制帧的 Protobuf 通讯、连接管理、鉴权、心跳、请求响应、房间广播和关闭处理。

10. 后台管理：管理员账号、密码哈希、Session、登录、退出、验证码、邮件测试入口和后台路由保护。
11. 多语言：JSON 语言包、请求语言识别、模板翻译函数和新增语言包约定。

## 日志规范

使用 Go 标准库 `log`，输出可直接阅读的文本格式。每一行包含时间、消息和 `key=value` 字段，例如：

```text
2026/09/02 17:30:00 component=news 新闻已创建 news_id=42 request_id=abc123
```

日志目录和文件名：

1. `logs/app-YYYY-MM-DD.log`：应用启动、业务、MySQL、Redis、缓存、定时任务、Worker、WebSocket 和错误日志。
2. `logs/access-YYYY-MM-DD.log`：每一条 HTTP 请求的访问日志，包含请求 ID、方法、路径、状态码、耗时、响应字节数和客户端 IP。
3. 本地开发默认同时输出到控制台；生产环境可由配置决定是否保留控制台输出。

按天切换规则：

1. 使用配置时区判断日期，默认 `Asia/Shanghai`。
2. 应用启动时打开当天文件；每次写入前发现日期变化时，在线关闭旧文件并打开新文件。
3. 文件切换和写入必须并发安全；应用优雅关闭时关闭全部日志文件。
4. 每次启动和切换时清理超过保留期限的日志文件，默认保留 30 天，配置可调整。
5. 当前阶段只按日期切换，不按文件大小分割；若单日日志量超过可接受范围，再增加大小分割策略。

日志配置项：

```text
LOG_DIR=./logs
LOG_TIMEZONE=Asia/Shanghai
LOG_RETENTION_DAYS=30
LOG_CONSOLE=true
LOG_ADD_SOURCE=false
```

日志字段与安全规则：

1. 每条业务日志尽量包含 `component`；请求链路额外包含 `request_id`。
2. HTTP 访问日志在请求结束后记录，保证包含最终状态码和耗时。
3. 错误使用 `logger.Printf`，并通过 `error=...` 字段记录错误内容。
4. 不记录密码、Authorization、Cookie、Token、完整身份证号、支付信息和原始 WebSocket 消息体。
5. `handler`、`service`、`repository`、Worker 和 WebSocket Service 使用标准库 `*log.Logger`；通过 `logging.Logger.With` 预置模块前缀，避免可变全局 Logger。

## 新闻参考模块

1. 新闻列表：支持筛选、排序、页码、每页数量和总数。
2. 新闻详情页。
3. 新闻创建与编辑表单，并能将校验错误返回到模板。
4. 新闻删除操作。
5. 使用 QesySQL 演示列表、查询、新增、更新、删除和分页。
6. 对新闻列表和详情使用 Redis 缓存；新增、更新、删除后明确清理相关缓存。
7. 在 Service 层实现事务示例：新闻记录与关联审计记录必须同时成功或同时回滚。
8. 前台使用 `/news` 和 `/news/{id}` 只展示已发布内容；后台使用 `/admin/news` 前缀完成受保护的新闻 CRUD。

## 定时任务与 Worker 示例

1. 使用 `robfig/cron/v3` 每分钟触发一个新闻发布状态检查任务。
2. 调度器以指定时区运行；任务默认包含超时、panic 恢复和同一任务不重入保护。
3. 定时任务只投递“刷新新闻缓存”任务到 Worker 队列，演示调度和实际耗时处理分离。
4. Worker 示例包含固定 Worker 数量、任务上下文取消、关闭时停止接收新任务并等待正在执行任务完成。
5. `internal/worker/example.go` 提供可直接复制的 `ExampleJob`，演示任务实现、日志、耗时处理和上下文取消。
6. 内存队列只作为框架示例；需要跨进程可靠投递时，后续通过 Redis Stream 或专业消息队列实现适配器。

## WebSocket 与 Protobuf 示例

1. WebSocket 路径暂定为 `/ws/game`，升级后只接受二进制帧，不使用 JSON 作为游戏消息格式。
2. 在 `proto/game/v1` 定义 `.proto` 文件，并使用 `protoc` 与 `protoc-gen-go` 生成 Go 代码；生成文件不手动编辑。
3. 定义统一消息信封：协议版本、消息编号、请求关联 ID，以及 `oneof` 业务负载。
4. 提供最小消息集合：客户端心跳、服务端心跳确认、加入房间、离开房间、房间广播和错误响应。
5. 连接管理器维护连接与房间关系；每个连接的写入通过单一路径或串行写队列完成，避免消息写入交叉。
6. Handler 只完成升级与鉴权；Protobuf 消息按编号路由至 WebSocket Service，游戏状态和规则不放在网络层。
7. 浏览器客户端使用 `ArrayBuffer` 接收二进制帧，并使用对应的 JavaScript/TypeScript Protobuf 生成代码编解码。

## 后台管理与认证示例

1. `admin_users` 表保存管理员账号、bcrypt 密码哈希、邮箱和状态；管理员信息不放在 `.env`。
2. `go run ./cmd/seedadmin` 交互式输入用户名、密码和邮箱，并写入或更新管理员账号。
3. `/admin/login` 展示登录表单和图片验证码；登录成功后通过 HttpOnly Session Cookie 进入后台。
4. Session 优先保存到 Redis；未启用 Redis 时使用进程内存，仅用于单机演示。
5. `auth.Manager.Require` 保护后台和新闻管理路由；未登录请求跳转到登录页。
6. `/admin/email/test` 演示通过 Resend HTTP API 发送邮件，邮件实现放在 `internal/mail`，业务代码不依赖具体邮件厂商。
7. Resend 未启用或配置不完整时必须返回明确错误，不能重定向为成功。
8. 当前示例只演示管理员登录态；角色、权限点和多管理员授权规则作为后续扩展。

## 多语言示例

1. 语言包存放在 `locales/<语言>/`，例如 `locales/zh-CN/common.json`、`locales/zh-CN/admin.json` 和 `locales/en-US/news.json`；每个语言目录都包含 `common.json`，页面文案按页面拆分。
2. 语言选择顺序为 URL 的 `lang` 参数、语言 Cookie、`Accept-Language` 请求头、默认英文 `en-US`；管理员初始化命令的交互提示也使用英文。
3. Handler 将 `T` 函数放入模板数据，模板使用 `{{call .T "admin.login.title"}}` 获取文案。
4. 新模块的页面文案必须使用语言 Key，不直接把用户可见中文写死在模板中。

## 交付阶段

### 第一阶段：项目基础

- 初始化 Go Module 和依赖版本。
- 创建目录结构。
- 添加强类型配置与本地环境变量示例。
- 完成应用启动、Chi、日志、优雅关闭和健康检查。
- 建立按天生成 `app` 与 `access` 日志文件的 `slog` 日志组件。
- 固定 Go 最低版本、Protobuf 代码生成工具和生成命令。

验收标准：应用可按照配置启动，并正常返回健康检查结果。

### 第二阶段：基础设施

- 通过 QesySQL 连接 MySQL，并连接 Redis。
- 添加迁移、模板渲染、静态资源服务、中间件和通用错误处理。
- 接入定时任务调度器、Worker 生命周期管理和 WebSocket 基础路由。

验收标准：启动时可检查 MySQL 与 Redis；可正常返回服务端渲染页面。

### 第三阶段：新闻 CRUD

- 定义新闻表迁移和模型。
- 实现基于 QesySQL 的 Repository、Service、Handler、路由和模板。

验收标准：浏览器中可完成新闻 CRUD；正常单表 Repository 方法均使用 QesySQL 链式 API。

### 第四阶段：缓存与事务

- 实现 Cache-Aside 读取和写入后的缓存失效。
- 增加审计表和事务示例。

验收标准：写入后缓存内容保持正确；事务失败时新闻和审计记录均回滚。

### 第五阶段：质量与文档

- 为 Service 添加聚焦的单元测试，为 Repository 添加适当的集成测试。
- 编写安装、配置、迁移、QesySQL 使用方式、Redis 使用方式、事务规则和新建业务模块流程的文档。

验收标准：新开发者可按照文档启动项目，并据此创建第二个业务模块。

### 第六阶段：后台认证与多语言

- 管理员账号、Session、验证码、邮件和语言包的基础示例已经实现。
- 交互式执行 `go run ./cmd/seedadmin` 写入真实管理员并完成登录冒烟验证。

## 开发约束

- 数据持久化代码只放在 Repository。
- 事务边界定义在 Service，不放在 Handler 或 Repository。
- Redis Key 必须按模块和版本命名，例如 `news:v1:detail:{id}`。
- 例外的原生 SQL 必须使用参数化查询。
- 模板中不编写业务逻辑。
- 修改共享基础设施或跨层行为时，必须补充相应测试。
- 日志切换、过期清理和敏感字段脱敏属于共享基础设施，必须有测试。

## 暂缓功能

- 多租户。
- JSON/OpenAPI API 模块。
- 后台 UI 组件库。

在基础框架和新闻参考模块建立稳定约定之前，以上功能暂不进入当前开发范围。
