# Web 层目录说明

- `cmd/web` 是 Web 程序入口，负责组装配置、数据库、缓存、路由和依赖。
- `web/templates/` 保存 Go `html/template` 页面模板。
- `web/static/` 保存浏览器直接访问的静态资源：`images/`、`scripts/`、`styles/`。
- `internal/infra/httpx/` 保存通用 HTTP 中间件，例如请求 ID和访问日志。
- `proto/` 保存 WebSocket 使用的 Protocol Buffers 协议定义和生成代码。

新增业务页面时，通常在 `internal/<业务>/handler`、`service`、`repository`、`model` 中按职责拆分，再在 `cmd/web` 或业务路由注册函数中组装。
