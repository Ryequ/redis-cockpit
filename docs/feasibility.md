# Redis TUI 可行性分析

## 结论

技术上可行。推荐采用 Go + Bubble Tea + Bubbles + Lipgloss + go-redis v9 实现跨平台终端工具，以“键盘优先、鼠标可选”的方式提供 Redis 管理体验。

但“操作 Redis 所有功能”不应按“为每个命令硬编码一个专属界面”理解。Redis 命令集合会随版本、模块和云厂商限制动态变化，因此必须采用以下三层覆盖策略：

1. **通用命令执行器**：用户可输入任意 Redis 命令，保证功能可达性。
2. **命令发现与注册表**：通过 `COMMAND`、`COMMAND DOCS`、`COMMAND INFO`、`MODULE LIST` 等命令动态发现可用命令，并提供静态 fallback。
3. **类型化适配器**：对高频数据结构和运维场景提供表单、分页、树形浏览和可视化结果。

首版可以覆盖绝大多数常用能力；“所有命令都有专属美观界面”应作为长期演进目标，而不是首版硬性承诺。

## 环境验证结果

| 项目 | 结果 | 影响 |
| --- | --- | --- |
| 系统 Go | `go1.19 windows/amd64` | 低于计划的 Go 1.22+ |
| 用户级 Go 工具链 | 已安装 `go1.22.5 windows/amd64` | 初始满足 Go 1.22+，不改动系统级 Go |
| GOPROXY | `https://goproxy.cn,direct` | 可拉取依赖 |
| GoOS/GoArch | `windows/amd64` | 当前构建主机为 Windows |
| `redis-server` | 未找到 | 集成测试需要外部 Redis 或后续下载 |
| Docker | 未找到 | 无法用容器启动 Cluster/Sentinel 测试环境 |

## 依赖可用性

通过 `go list -m -versions` 验证的关键依赖版本：

- `github.com/charmbracelet/bubbletea`：最新稳定版本 `v1.3.10`
- `github.com/charmbracelet/bubbles`：最新稳定版本 `v1.0.0`
- `github.com/charmbracelet/lipgloss`：最新稳定版本 `v1.1.0`
- `github.com/redis/go-redis/v9`：最新稳定版本 `v9.24.0`

实际拉取依赖时发现 `go-redis/v9 v9.24.0` 要求 Go 1.26+，Go 工具链自动切换到 `go1.26.11`。当前 `go.mod` 声明 `go 1.26.0`，README 与 CI 均以 Go 1.26 为构建基线。其他依赖版本以 `go mod tidy` 后的 `go.sum` 为准。

## 技术可行性

### 终端 UI

Bubble Tea 提供 Model-Update-View 架构，适合处理键盘事件、窗口尺寸变化和异步消息。Bubbles 提供 list、table、textinput、textarea、viewport、help 等组件，Lipgloss 可实现边框、颜色、焦点态和高密度布局。

鼠标能力通过 Bubble Tea 的 mouse cell/all motion 开启。实际可用性依赖终端模拟器和用户配置：

- Linux/macOS 常见终端通常支持鼠标事件；
- Windows Terminal 支持较好，PowerShell/CMD 传统控制台可能不支持；
- 工具应键盘优先，并在鼠标不可用时自动隐藏鼠标提示。

### Redis 客户端

go-redis v9 提供 `UniversalClient`，可统一 single、failover 和 cluster 客户端创建方式，并支持认证、TLS、数据库选择、超时、连接池等配置。可满足单机、远程、Sentinel、Cluster 与 TLS 场景的基础连接需求。

需要专门验证或额外设计的部分：

- Pub/Sub 与 RESP3 push message 的行为；
- MONITOR 这类持续流式命令的取消与背压；
- Cluster 跨槽错误和只读命令路由提示；
- 云托管 Redis 禁用命令时的降级提示；
- 模块命令的参数提示和结果渲染。

### 全命令覆盖

Redis 命令生态是动态的，原因包括：

- 不同 Redis 版本支持的命令不同；
- Redis Stack、Search、JSON、Bloom 等模块会注册新命令；
- 云厂商会禁用 `CONFIG`、`CLIENT`、`DEBUG` 等高风险命令；
- ACL 会让同一实例上的不同用户看到不同命令权限。

因此全命令覆盖应定义为：**用户能发现并调用实例当前允许的命令，且常见命令具备友好界面；未知或模块命令至少可通过原始执行器操作。**

## 推荐架构

分层如下：

1. **UI 层**：Bubble Tea Model、键盘/鼠标事件、页面布局、主题和组件。
2. **应用状态层**：当前连接、当前视图、加载状态、命令历史、结果集、分页和错误提示。
3. **命令层**：命令元数据、动态发现、参数解析、自动补全、危险命令识别。
4. **Redis 执行层**：连接档案、UniversalClient、同步/异步执行、超时取消、流式通道。
5. **结果渲染层**：RESP 结果转换为表格、树、JSON、原始文本、二进制预览和日志流。
6. **安全层**：只读模式、危险命令确认、敏感信息脱敏、本地审计。
7. **配置层**：连接档案、快捷键、主题、分页和预览上限。

网络请求必须放在 Bubble Tea 的异步 command 中，UI 线程只负责状态更新，避免阻塞输入响应。

## 关键风险与缓解

| 风险 | 影响 | 缓解方式 |
| --- | --- | --- |
| 命令数量庞大且动态变化 | 无法完整硬编码所有命令 | 通用执行器 + 动态发现 + 类型化适配器 |
| 模块命令体验不一致 | 无法预设表单和结果 | 原始执行兜底，按模块逐步增加适配 |
| Cluster 命令路由复杂 | 跨槽命令失败 | 使用 UniversalClient，识别跨槽错误并给出提示 |
| Pub/Sub/MONITOR 持续输出 | 占满内存或界面 | 有界 channel、限频刷新、暂停/清空/退出 |
| 大 Key 与大结果集 | 终端卡死或内存暴涨 | SCAN/HSCAN/SSCAN/ZSCAN/XRANGE 分页，结果截断 |
| 危险命令误操作 | 数据丢失 | 二次确认、只读模式、危险等级标记 |
| Windows 传统 CMD 鼠标兼容差 | 鼠标不可用 | 键盘优先，能力检测，鼠标不可用时降级 |
| RESP3 与 push message 差异 | 流式命令不稳定 | 明确版本验证，单独集成测试 |

## 性能约束

- Key 浏览默认使用 `SCAN`，禁止默认使用 `KEYS`。
- Hash/Set/ZSet 使用 `HSCAN`/`SSCAN`/`ZSCAN` 分页。
- Stream 使用 `XRANGE`/`XREVRANGE` 分页。
- TYPE/TTL 等元数据使用 pipeline 批量获取，避免 N+1 请求。
- 命令元数据内存缓存并设置 TTL。
- 单页结果、value 预览长度、日志条数均设置上限。
- 所有网络操作支持 context 超时和取消。

## 安全设计

- 危险命令：`FLUSHALL`、`FLUSHDB`、`DEL`、`UNLINK`、`KEYS`、`CONFIG SET`、`SHUTDOWN`、`ACL DELETE` 等需要二次确认。
- 支持只读模式，拦截写命令。
- 命令历史和配置中的密码字段脱敏。
- 危险操作在本地审计日志中记录时间、连接档案和命令名，不记录密码或完整敏感参数。

## 测试策略

- 单元测试：使用 miniredis 覆盖连接配置、命令解析、安全策略、结果转换和分页逻辑。
- TUI 测试：测试 Model.Update 的键盘和鼠标事件，测试 View 输出快照。
- 集成测试：真实 Redis 覆盖数据类型、事务、Lua、Pub/Sub、Stream、ACL。
- 拓扑测试：single、cluster、sentinel、TLS、密码认证和受限命令。
- 手工验收：Windows Terminal、PowerShell、CMD、Linux 常见终端、macOS Terminal 与 iTerm2。
- 发布验证：`go vet`、`go test ./...`、多平台交叉编译和示例配置启动验证。

当前环境未安装 `redis-server` 与 Docker，真实 Redis 集成测试应通过环境变量启用；未启用时自动跳过，避免阻塞本地开发。

## 实施边界

首版建议范围：

- 连接档案、Key 浏览、String/Hash/List/Set/ZSet/Stream 基本操作；
- 原始命令执行、命令搜索和历史；
- INFO/CONFIG/CLIENT/SLOWLOG/复制状态查看；
- 危险命令确认、只读模式、结果分页；
- Windows/Linux/macOS 构建与键盘交互验证。

后续增强：

- 专属 Pub/Sub、MONITOR、Lua、事务、ACL、Search/JSON 页面；
- Cluster 多节点诊断视图；
- 配置导入导出、主题定制、插件化命令适配器；
- 完整跨平台鼠标兼容矩阵和自动能力检测。
