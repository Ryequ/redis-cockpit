# 架构说明

## 分层

```text
main.go / cmd
    └── internal/app        启动、配置加载、终端能力检测
        └── internal/ui     Bubble Tea Model、视图路由、键盘/鼠标事件
            ├── internal/command   命令元数据、解析、历史、风险分级
            ├── internal/config    配置与连接档案持久化
            ├── internal/redis     连接、通用执行器、命令发现、类型适配器
            ├── internal/ops       INFO/CONFIG/CLIENT/SLOWLOG/复制、脚本、事务
            ├── internal/stream    Pub/Sub、MONITOR、有界消息通道
            ├── internal/result    RESP 结果规范化、渲染、分页
            └── internal/security  只读策略、确认策略、脱敏、审计日志
```

UI 层不直接执行 Redis 网络操作。所有网络请求通过 Bubble Tea `tea.Cmd` 异步执行，完成后以消息形式回到 `Model.Update`，避免阻塞键盘和鼠标响应。

## 数据流

1. 启动时 `app.Run` 加载配置文件，解析 `--profile`、`--readonly` 等参数。
2. `ui.Model` 初始化连接档案列表、命令输入、历史、安全策略。
3. 用户在 Connections 视图选择档案并按 Enter，触发 `connectCmd`。
4. 连接成功后创建 `redis.Executor`、`adapters.Manager`、`ops.Manager` 和 `stream.Hub`。
5. Keys 视图使用 `SCAN` 加载 Key，通过 pipeline 批量获取 `TYPE` 和 `TTL`。
6. Workbench 使用 `command.Parse` 解析输入，经安全策略检查后执行任意命令。
7. 结果由 `result.Normalize` 统一为 `Value`，再渲染成原始文本、表格或 JSON。
8. Ops 视图读取 `INFO`、`CONFIG`、`CLIENT`、`SLOWLOG`、`MODULE LIST`、`ACL LIST`。
9. Streams 视图通过有界 channel 接收 Pub/Sub 和 MONITOR 消息，避免无界内存增长。

## 全命令覆盖策略

Redis 命令集合会随版本、模块和云厂商限制变化，因此采用三层覆盖：

1. **通用执行器**：`Executor.Execute` 可执行任意 argv，保证功能可达。
2. **命令注册表**：内置命令提供元数据；`Discoverer` 通过 `COMMAND` 动态补充，失败时回退到静态命令表。
3. **类型化适配器**：高频数据类型和运维场景提供专属交互。

模块命令若没有专属界面，仍可在 Workbench 中执行。`MODULE LIST` 的结果在 Ops 视图展示。

## 安全边界

- `command.RiskOf` 将命令分为 none/read/write/critical。
- `security.Policy` 在只读模式拦截 write 和 critical。
- critical 命令必须按 `y` 确认，按 `n` 或 `Esc` 取消。
- Key 删除使用独立确认流程。
- `security.RedactCommand` 和 `RedactText` 对认证、密码参数进行脱敏。
- `security.AuditLog` 以 JSONL 追加脱敏审计记录。
- 结果展示有分页和预览限制；Key 扫描默认使用 `SCAN` 而非 `KEYS`。

## 性能策略

- `SCAN`/`HSCAN`/`SSCAN`/`ZSCAN`/`XRANGE` 分页读取。
- `TYPE` 和 `TTL` 使用 pipeline，避免 N+1。
- 命令元数据缓存 5 分钟。
- Pub/Sub 和 MONITOR 使用有界 channel，超出容量丢弃最旧消息。
- 视图列表限制展示条数，避免终端渲染过多行。

## 跨平台

- Linux/macOS 默认启用鼠标。
- Windows 检测 Windows Terminal、ConEmu 或常见 `TERM` 值，传统 CMD 自动降级为纯键盘。
- `TERM=dumb` 时禁用鼠标提示。
- 构建目标覆盖 windows/amd64、linux/amd64、darwin/amd64 和 darwin/arm64。

## 扩展方式

### 新增命令表单

1. 在 `internal/command/registry.go` 添加 `Spec`。
2. 如需类型化渲染，在 `internal/result` 增加渲染格式。
3. 在 Workbench 中可直接通过通用执行器调用，无需新增网络代码。

### 新增数据类型页面

1. 在 `internal/redis/adapters` 增加读取/写入方法。
2. 在 `internal/ui/messages.go` 增加异步消息和命令。
3. 在 `internal/ui/update.go` 接入快捷键，在 `view.go` 渲染页面。

### 新增模块支持

1. 通过 `MODULE LIST` 和 `COMMAND` 发现命令。
2. 为高频模块命令添加 `Spec` 和可选适配器。
3. 其他模块命令保留 Workbench 原始执行入口。

## 测试

- 单元测试覆盖配置、解析、策略、结果转换、适配器和 UI Model 行为。
- 集成测试位于 `test/integration`，通过构建标签 `integration` 和环境变量启用。
- 发布前运行 `go vet ./...`、`go test ./...`、多平台交叉编译。
