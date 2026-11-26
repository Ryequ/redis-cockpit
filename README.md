# redis-cockpit

一个在终端中操作 Redis 的 TUI 工具，支持键盘优先操作和鼠标选择，覆盖数据浏览、命令执行、运维监控、Pub/Sub、MONITOR、脚本与事务等能力。

## 功能

- 连接管理：单机、远程、Cluster、Sentinel、TLS、密码认证、只读实例。
- Key 浏览：`SCAN` 分页、类型/TTL 展示、筛选、值预览、删除确认。
- 数据类型：String、Hash、List、Set、ZSet、Stream 的分页读取。
- 命令工作台：命令解析、历史记录、任意 Redis 命令执行、危险命令确认。
- 运维视图：`INFO`、`CONFIG`、`CLIENT LIST`、`SLOWLOG`、复制状态、模块与 ACL。
- 流式任务：Pub/Sub 订阅发布、MONITOR、Stream 条目操作。
- 高级能力：Lua 脚本和 MULTI/EXEC 事务。
- 安全：只读模式、危险命令二次确认、密码脱敏、JSONL 审计日志。
- 交互：Tab/数字键切换视图，方向键或 hjkl 导航，鼠标点击导航和列表，终端不支持鼠标时自动降级。

## 环境要求

- Go 1.26 或更高版本。当前 `go-redis/v9 v9.24.0` 要求 Go 1.26+。
- 可连接到的 Redis 实例。本地开发可使用 `127.0.0.1:6379`。

## 构建与运行

```bash
go build -o bin/redis-cockpit .
./bin/redis-cockpit
```

Windows PowerShell:

```powershell
go build -o bin/redis-cockpit.exe .
.\bin\redis-cockpit.exe
```

启动参数：

```text
--config      指定配置文件路径
--profile     启动时选中的连接档案名称
--readonly    强制只读模式
--debug       启用调试日志
```

示例：

```bash
./bin/redis-cockpit --config ./config.example.yaml --profile local
```

## 配置

默认配置文件路径由 `os.UserConfigDir` 决定，通常为：

- Linux: `$XDG_CONFIG_HOME/redis-cockpit/config.yaml`
- macOS: `$HOME/Library/Application Support/redis-cockpit/config.yaml`
- Windows: `%APPDATA%\redis-cockpit\config.yaml`

首次启动时若配置文件不存在，会使用默认 `local` 连接档案（`127.0.0.1:6379`）。

完整示例见 `config.example.yaml`。密码字段支持环境变量展开，例如：

```yaml
password: ${REDIS_TUI_PASSWORD}
```

### 新增远程连接

当前 TUI 不提供连接档案的新增/编辑表单，请直接编辑配置文件中的 `profiles` 列表。例如：

```yaml
  - id: prod
    name: prod
    addr: redis.example.com:6379
    username: default
    password: ${REDIS_TUI_PASSWORD}
    db: 0
    tls: true
    insecure_skip_verify: false
    cluster: false
    sentinel: false
    read_only: true
    timeout_seconds: 10
```

密码建议通过环境变量注入，避免明文写入配置文件。修改配置后需重启工具；启动时可用 `--profile prod` 直接选择档案，也可在 Connections 页面用 `↑` / `↓` 选择后按 `Enter` 连接。

Cluster 实例将 `cluster` 设为 `true`；Sentinel 实例将 `sentinel` 设为 `true` 并填写 `sentinel_addrs` 与 `master_name`。两种模式不能同时开启。

## 快捷键

| 按键 | 功能 |
| --- | --- |
| `1` - `5` | 切换 Connections / Keys / Workbench / Ops / Streams |
| `Tab` / `Shift+Tab` | 循环切换视图 |
| `↑` / `↓` 或 `k` / `j` | 移动选择 |
| `Enter` | 连接、查看值、执行命令、确认 |
| `Esc` | 清除输入、关闭弹层、取消确认 |
| `?` | 显示帮助 |
| `Ctrl+C` / `Ctrl+Q` | 退出 |
| `/` | 在 Keys 视图编辑筛选条件 |
| `r` | 刷新 Keys 或 Ops 快照 |
| `n` | 加载下一页 Key |
| `d` | 删除选中的 Key（需确认） |
| `s` / `p` / `m` / `x` | 在 Streams 视图订阅、发布、启动 MONITOR、停止 |

鼠标可点击左侧导航、连接列表和 Key 列表。传统 Windows CMD 可能不支持鼠标事件，工具会自动隐藏鼠标提示并保留键盘操作。

## 测试

单元测试：

```bash
go test ./...
```

集成测试需要真实 Redis，默认跳过：

```bash
go test -tags integration ./test/integration
```

集成测试环境变量：

| 变量 | 用途 |
| --- | --- |
| `REDIS_TUI_TEST_ADDR` | 单机 Redis 地址 |
| `REDIS_TUI_TEST_CLUSTER_ADDR` | Cluster 节点地址 |
| `REDIS_TUI_TEST_SENTINEL_ADDRS` | Sentinel 地址列表，逗号分隔 |
| `REDIS_TUI_TEST_MASTER_NAME` | Sentinel master 名称 |
| `REDIS_TUI_TEST_TLS_ADDR` | TLS Redis 地址 |
| `REDIS_TUI_TEST_TLS_PASSWORD` | TLS Redis 密码，可选 |

## 审计日志

危险命令和写命令执行时会写入配置目录下的 `audit.jsonl`。日志只记录时间、连接档案和脱敏后的命令，不记录密码或完整敏感参数。

## 架构

见 `docs/architecture.md`。可行性分析见 `docs/feasibility.md`。
