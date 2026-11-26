# redis-cockpit

> [!WARNING]
> **重要声明 / Important Notice**
>
> **本项目仅用于本人娱乐、学习和个人测试，不面向生产环境，不提供任何稳定性、安全性或兼容性保证。请勿在重要业务系统、生产数据或不可恢复的数据上使用。**
>
> **This project is strictly for personal entertainment, learning and testing only. It is not intended for production use and comes with no stability, security or compatibility guarantees. Do not use it with critical systems, production data or irrecoverable data.**

`redis-cockpit` 是一个终端内运行的 Redis TUI 工具，提供连接管理、Key 浏览、命令执行、运维监控、Pub/Sub、MONITOR、脚本与事务等功能。工具以键盘操作为主，并支持在兼容终端中使用鼠标。

当前发布版本：`v0.2.0`

## 功能特性

- 连接管理：单机、远程、Cluster、Sentinel、TLS、密码认证、只读实例、SSH 隧道与云 Redis 模板。
- 连接页增强：搜索、排序、按模式分组、复制连接、导入/导出、健康状态、延迟与版本展示。
- 实例摘要：Connections 页并排展示 Server、Memory、Stats 三组关键指标。
- Key 浏览：`SCAN` 分页、类型/TTL 展示、筛选、树形前缀视图、值预览与滚动浏览。
- 数据类型：String、Hash、List、Set、ZSet、Stream 的分页读取。
- 命令工作台：命令解析、历史记录、任意 Redis 命令执行、危险命令确认。
- 运维视图：`INFO`、`CONFIG`、`CLIENT LIST`、`SLOWLOG`、复制状态、模块与 ACL。
- 流式任务：Pub/Sub 订阅发布、MONITOR、Stream 条目操作。
- 高级能力：Lua 脚本与 MULTI/EXEC 事务。
- 安全：只读模式、写命令二次确认、密码脱敏、JSONL 审计日志。
- 交互：中英文界面、多套主题、顶部标签、左侧导航、可点击按钮与确认弹层。

## 构建环境

从源码构建需要：

- Go 1.26 或更高版本。
- 可连接到的 Redis 实例。本地开发可使用 `127.0.0.1:6379`。
- Windows Terminal、macOS Terminal/iTerm2 或 Linux xterm 兼容终端可获得完整鼠标支持；传统 Windows CMD 会自动降级为键盘操作。

## 如何使用

### 安装

#### 方式一：Go 命令行安装

如果设备已安装 Go 1.26+：

```bash
go install github.com/Ryequ/redis-cockpit@latest
```

安装后的可执行文件位置：

```text
Linux/macOS: ~/go/bin/redis-cockpit
Windows:     %USERPROFILE%\go\bin\redis-cockpit.exe
```

确认安装：

```bash
redis-cockpit
```

如果需要安装指定版本：

```bash
go install github.com/Ryequ/redis-cockpit@v0.2.0
```

#### 方式二：从 GitHub Releases 下载预编译二进制

打开 Releases 页面：

```text
https://github.com/Ryequ/redis-cockpit/releases/latest
```

根据设备选择：

| 平台 | 文件名 |
| --- | --- |
| Windows x64 | `redis-cockpit-windows-amd64.exe` |
| Linux x64 | `redis-cockpit-linux-amd64` |
| Linux ARM64 | `redis-cockpit-linux-arm64` |
| macOS Intel | `redis-cockpit-darwin-amd64` |
| macOS Apple Silicon | `redis-cockpit-darwin-arm64` |

Linux x64 示例：

```bash
curl -L -o redis-cockpit \
  https://github.com/Ryequ/redis-cockpit/releases/latest/download/redis-cockpit-linux-amd64

chmod +x redis-cockpit
sudo mv redis-cockpit /usr/local/bin/
redis-cockpit
```

macOS Apple Silicon 示例：

```bash
curl -L -o redis-cockpit \
  https://github.com/Ryequ/redis-cockpit/releases/latest/download/redis-cockpit-darwin-arm64

xattr -cr ./redis-cockpit
chmod +x redis-cockpit
sudo mv redis-cockpit /usr/local/bin/
redis-cockpit
```

Windows PowerShell 示例：

```powershell
Invoke-WebRequest `
  -Uri "https://github.com/Ryequ/redis-cockpit/releases/latest/download/redis-cockpit-windows-amd64.exe" `
  -OutFile "redis-cockpit.exe"

.\redis-cockpit.exe
```

每个 Release 还包含 `checksums.txt`，可用于校验下载文件。

### 从源码构建

```bash
git clone https://github.com/Ryequ/redis-cockpit.git
cd redis-cockpit
go build -o bin/redis-cockpit .
./bin/redis-cockpit
```

Windows PowerShell：

```powershell
go build -o bin/redis-cockpit.exe .
.\bin\redis-cockpit.exe
```

### 启动参数

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

默认配置文件路径由 `os.UserConfigDir` 决定：

- Linux：`$XDG_CONFIG_HOME/redis-cockpit/config.yaml`
- macOS：`$HOME/Library/Application Support/redis-cockpit/config.yaml`
- Windows：`%APPDATA%\redis-cockpit\config.yaml`

首次启动时若配置文件不存在，会使用默认 `local` 连接档案（`127.0.0.1:6379`）。

完整示例见 [`config.example.yaml`](./config.example.yaml)。主要配置项：

```yaml
active_profile: local
language: zh
theme: aurora
page_size: 100
max_preview: 4096
```

- `language`：支持 `zh` / `en`
- `theme`：支持 `aurora` / `ember` / `ocean` / `graphite`

也可以在 TUI 的 Settings 页面切换语言和主题。

### 密码与环境变量

密码字段支持环境变量占位符：

```yaml
password: ${REDIS_TUI_PASSWORD}
```

占位符只在建立连接时展开；通过 TUI 保存配置会保留占位符，不会把明文密码写回配置文件。

### 连接模板

配置中的 `profiles` 示例：

```yaml
profiles:
  - id: prod
    name: prod
    addr: redis.example.com:6379
    username: default
    password: ${REDIS_TUI_PASSWORD}
    db: 0
    tls: true
    cluster: false
    sentinel: false
    read_only: true
    timeout_seconds: 10
```

Cluster 实例将 `cluster` 设为 `true`；Sentinel 实例将 `sentinel` 设为 `true` 并填写 `sentinel_addrs` 与 `master_name`。两种模式不能同时开启。

SSH 隧道通过本机 `ssh` 命令建立本地端口转发，支持 SSH agent 或 `ssh_key_path` 指定的私钥认证，不支持交互式密码输入。云 Redis 模板可在连接表单中点击 `[AWS]` / `[GCP]` / `[Azure]` / `[Aliyun]` / `[Tencent]` 快速设置 TLS、云厂商和认证方式字段；`iam` 等云 token 需要由外部工具生成后填入 `password` 字段。

## 测试

单元测试：

```bash
go test ./...
```

集成测试：

```bash
go test -tags integration ./test/integration
```

静态检查：

```bash
go vet ./...
```

## 安全说明

- 危险命令和写命令执行时会写入配置目录下的 `audit.jsonl`。
- 日志只记录时间、连接档案和脱敏后的命令，不记录密码或完整敏感参数。
- 工具包含只读模式、危险命令确认、删除/编辑确认和连接档案健康检查。
- 尽管如此，本项目仍不应被用于生产环境或不可恢复数据。

## 项目结构

```text
.
├── cmd/                    # 命令行参数
├── internal/
│   ├── app/                # 应用启动
│   ├── command/            # Redis 命令解析
│   ├── config/             # 配置读写
│   ├── ops/                # 运维数据解析
│   ├── result/             # 返回值归一化与渲染
│   ├── security/           # 策略、脱敏、审计
│   ├── stream/             # Pub/Sub 与 MONITOR
│   └── ui/                 # TUI 视图与交互
├── test/integration/       # 集成测试
├── .github/workflows/      # CI 与 Release 工作流
├── config.example.yaml
└── main.go
```

## 相关文档

- [架构说明](./docs/architecture.md)
- [可行性分析](./docs/feasibility.md)
- [配置示例](./config.example.yaml)

## 许可证

当前仓库未附带 `LICENSE` 文件。仓库公开可访问不代表任何形式的授权、担保或许可。
