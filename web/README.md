# Web 层目录说明

- `cmd/web` 是 Web 程序入口，负责组装配置、数据库、缓存、路由和依赖。
- `web/templates/pages/`、`web/templates/admin/` 按页面范围保存 Go `html/template` 页面；每个范围可用 `layouts/` 和 `partials/` 保存布局及局部模板。
- `web/static/` 保存浏览器直接访问的静态资源：`images/`、`scripts/`、`styles/`。
- `internal/infra/view/` 按模板范围解析与渲染页面，生产环境启动时缓存，development 环境请求时重新解析；支持 `capitalize`、`upper`、`lower`、`add`、`sub`、`mul`、`div`、`floor`、`round`。
- `internal/infra/httpx/` 保存通用 HTTP 中间件，例如请求 ID和访问日志。
- `proto/` 保存 WebSocket 使用的 Protocol Buffers 协议定义和生成代码。

新增页面时，把模板放到对应范围目录，并通过 `view.Renderer.Render` 使用范围化名称，例如 `pages/home.html` 或 `admin/admin_login.html`。业务代码按 `internal/handler`、`internal/service`、`internal/repository`、`internal/model` 分层，再由 `cmd/web` 组装。
