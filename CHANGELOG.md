# 更新日志

本项目按版本记录面向使用者的重要变更。版本遵循语义化版本，日期在实际发布时填写。

## [1.1.0] - 2026-10-14

### 新增

- 新增 `internal/infra/view` 模板渲染器，支持按 `pages`、`admin` 等范围隔离模板。
- 增加模板函数：`capitalize`、`upper`、`lower`、`add`、`sub`、`mul`、`div`、`floor`、`round`。
- 新增 `internal/infra/validate` 通用校验库，覆盖必填、相等、邮箱、电话、中国手机号、密码长度与复杂度、URL、域名和 IPv4。
- 新增 `internal/infra/apperror`，用于封装带错误码和参数的应用错误。

### 调整

- 代码组织改为按层分目录：`handler`、`service`、`repository`、`model`。
- 新闻前台和后台继续共用 Service 与 Repository，由各自 Handler 处理 HTTP 请求。
- 模板迁入 `web/templates/pages/` 和 `web/templates/admin/`，由 `view.Renderer` 按范围名渲染。
- 模板在生产环境启动时解析并缓存，在 `development` 环境按请求重新解析。
- 更新项目结构和模板使用文档。

### 移除

- 移除旧的 `internal/modules` 与 `internal/data` 目录结构。

### 发布前验证

- `go build ./cmd/web ./cmd/migrate ./cmd/seedadmin` 通过。
- Renderer 成功解析全部页面模板。
- 本版本未运行测试套件。
