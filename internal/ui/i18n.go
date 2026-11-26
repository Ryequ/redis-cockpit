package ui

import "strings"

var zhText = map[string]string{
	"Ready":       "就绪",
	"Connections": "连接管理",
	"Keys":        "键浏览",
	"Workbench":   "命令工作台",
	"Ops":         "运维监控",
	"Streams":     "流式任务",
	"Settings":    "设置",
	"连接档案、认证、TLS、Cluster 与 Sentinel 将在此管理。":                                    "连接档案、认证、TLS、Cluster 与 Sentinel 将在此管理。",
	"使用 SCAN 分页浏览 Key，查看类型、TTL 与值预览。":                                          "使用 SCAN 分页浏览 Key，查看类型、TTL 与值预览。",
	"搜索命令、填写参数或执行原始 Redis 命令。":                                                 "搜索命令、填写参数或执行原始 Redis 命令。",
	"INFO、CONFIG、CLIENT、SLOWLOG 与复制状态。":                                        "INFO、CONFIG、CLIENT、SLOWLOG 与复制状态。",
	"Pub/Sub、MONITOR 与 Stream 实时消息流。":                                          "Pub/Sub、MONITOR 与 Stream 实时消息流。",
	"设置语言、主题和查看快捷键帮助。":                                                         "设置语言、主题和查看快捷键帮助。",
	"↑↓/j/k 选择  •  Enter 连接  •  n 新增  •  e 编辑  •  d 删除  •  t 测试  •  鼠标点击列表项选择": "↑↓/j/k 选择  •  Enter 连接  •  n 新增  •  e 编辑  •  d 删除  •  t 测试  •  鼠标点击列表项选择",
	"↑↓/j/k 选择  •  Enter 查看值  •  a 新增 Key  •  v 修改 Value  •  t 树/平铺  •  ←/→ 折叠/展开  •  滚轮/PgUp/PgDn 滚动  •  / 筛选  •  r 刷新": "↑↓/j/k 选择  •  Enter 查看值  •  a 新增 Key  •  v 修改 Value  •  t 树/平铺  •  ←/→ 折叠/展开  •  滚轮/PgUp/PgDn 滚动  •  / 筛选  •  r 刷新",
	"  •  n 下一页": "  •  n 下一页",
	"r 刷新快照  •  脚本与事务可在 Workbench 中执行":                      "r 刷新快照  •  脚本与事务可在 Workbench 中执行",
	"Stream: 使用 Workbench 执行 XADD/XRANGE/XREAD 管理流条目。":      "Stream: 使用 Workbench 执行 XADD/XRANGE/XREAD 管理流条目。",
	"s 订阅  •  p 发布测试消息  •  m 启动 MONITOR  •  x 停止  •  输入频道名": "s 订阅  •  p 发布测试消息  •  m 启动 MONITOR  •  x 停止  •  输入频道名",
	"  •  Mouse Click 选择": "  •  鼠标点击选择",
	"Tab/1-6 切换页面  •  ↑↓/hjkl 导航  •  Enter 确认  •  ? 设置  •  Ctrl+C 退出": "Tab/1-6 切换页面  •  ↑↓/hjkl 导航  •  Enter 确认  •  ? 设置  •  Ctrl+C 退出",
}

var enText = map[string]string{
	"设置":    "Settings",
	"语言":    "Language",
	"主题":    "Theme",
	"帮助":    "Help",
	"当前语言":  "Current language",
	"当前主题":  "Current theme",
	"中文":    "Chinese",
	"英文":    "English",
	"连接管理":  "Connections",
	"键浏览":   "Keys",
	"命令工作台": "Workbench",
	"运维监控":  "Ops",
	"流式任务":  "Streams",
	"就绪":    "Ready",
	"连接档案、认证、TLS、Cluster 与 Sentinel 将在此管理。":                                    "Manage profiles, authentication, TLS, Cluster and Sentinel here.",
	"使用 SCAN 分页浏览 Key，查看类型、TTL 与值预览。":                                          "Browse keys with SCAN paging, including type, TTL and value preview.",
	"搜索命令、填写参数或执行原始 Redis 命令。":                                                 "Search commands, fill arguments or execute raw Redis commands.",
	"INFO、CONFIG、CLIENT、SLOWLOG 与复制状态。":                                        "INFO, CONFIG, CLIENT, SLOWLOG and replication status.",
	"Pub/Sub、MONITOR 与 Stream 实时消息流。":                                          "Pub/Sub, MONITOR and live Stream activity.",
	"设置语言、主题和查看快捷键帮助。":                                                         "Configure language, theme and view keyboard help.",
	"↑↓/j/k 选择  •  Enter 连接  •  n 新增  •  e 编辑  •  d 删除  •  t 测试  •  鼠标点击列表项选择": "↑↓/j/k select  •  Enter connect  •  n new  •  e edit  •  d delete  •  t test  •  click a row to select",
	"↑↓/j/k 选择  •  Enter 查看值  •  a 新增 Key  •  v 修改 Value  •  t 树/平铺  •  ←/→ 折叠/展开  •  滚轮/PgUp/PgDn 滚动  •  / 筛选  •  r 刷新": "↑↓/j/k select  •  Enter view value  •  a add key  •  v edit value  •  t tree/flat  •  ←/→ collapse/expand  •  wheel/PgUp/PgDn scroll  •  / filter  •  r refresh",
	"  •  n 下一页": "  •  n next page",
	"r 刷新快照  •  脚本与事务可在 Workbench 中执行":                      "r refresh snapshot  •  run scripts and transactions in Workbench",
	"Stream: 使用 Workbench 执行 XADD/XRANGE/XREAD 管理流条目。":      "Stream: use Workbench to run XADD/XRANGE/XREAD for stream entries.",
	"s 订阅  •  p 发布测试消息  •  m 启动 MONITOR  •  x 停止  •  输入频道名": "s subscribe  •  p publish test message  •  m start MONITOR  •  x stop  •  type a channel",
	"  •  Mouse Click 选择": "  •  mouse click to select",
	"Tab/1-6 切换页面  •  ↑↓/hjkl 导航  •  Enter 确认  •  ? 设置  •  Ctrl+C 退出": "Tab/1-6 switch view  •  ↑↓/hjkl navigate  •  Enter confirm  •  ? settings  •  Ctrl+C quit",
	"Redis Cockpit 快捷键":               "Redis Cockpit shortcuts",
	"1-6 / Tab / Shift+Tab   切换页面":    "1-6 / Tab / Shift+Tab   switch view",
	"↑↓ / j k                移动选择":    "↑↓ / j k                move selection",
	"Enter                   确认 / 执行": "Enter                   confirm / execute",
	"Esc                     返回 / 关闭": "Esc                     back / close",
	"?                       显示帮助":    "?                       show help",
	"Ctrl+C / Ctrl+Q         退出":      "Ctrl+C / Ctrl+Q         quit",
	"鼠标：点击顶部标签、左侧导航、按钮和列表项；滚轮滚动长结果。": "Mouse: click top tabs, left navigation, buttons and list rows; scroll wheel for long results.",
	"快捷键": "Shortcuts",
	"点击按钮切换语言。界面文本会立即更新并保存到配置文件。": "Click a button to switch language. Interface text updates immediately and is saved to the config file.",
	"点击按钮切换主题配色。按钮颜色会随主题同步更新。":    "Click a button to switch the color theme. Button colors update with the theme.",
}

var zhButtons = map[string]string{
	"[1 Connections]": "[1 连接管理]",
	"[2 Keys]":        "[2 键浏览]",
	"[3 Workbench]":   "[3 命令工作台]",
	"[4 Ops]":         "[4 运维监控]",
	"[5 Streams]":     "[5 流式任务]",
	"[6 Settings]":    "[6 设置]",
	"[Connect]":       "[连接]",
	"[N new]":         "[新增]",
	"[E edit]":        "[编辑]",
	"[D delete]":      "[删除]",
	"[C clone]":       "[复制]",
	"[I import]":      "[导入]",
	"[X export]":      "[导出]",
	"[T test]":        "[测试]",
	"[Test all]":      "[全部测试]",
	"[Search]":        "[搜索]",
	"[Sort]":          "[排序]",
	"[Group]":         "[分组]",
	"[Confirm]":       "[确认]",
	"[Y] Delete":      "[Y] 删除",
	"[N] Cancel":      "[N] 取消",
	"[Input]":         "[输入]",
	"[Execute]":       "[执行]",
	"[Y] Execute":     "[Y] 执行",
	"[Refresh]":       "[刷新]",
	"[Next]":          "[下一个]",
	"[New key]":       "[新建 Key]",
	"[Edit value]":    "[编辑 Value]",
	"[Del]":           "[删除]",
	"[Tree]":          "[树]",
	"[Flat]":          "[平铺]",
	"[Copy value]":    "[复制 Value]",
	"[OK]":            "[确认]",
	"[Cancel]":        "[取消]",
	"[Save]":          "[保存]",
	"[Channel]":       "[频道]",
	"[Sub]":           "[订阅]",
	"[Pub]":           "[发布]",
	"[Monitor]":       "[监控]",
	"[Stop]":          "[停止]",
	"[Close]":         "[关闭]",
}

var enButtons = map[string]string{
	"[中文]": "[Chinese]",
}

func buttonLabel(language, canonical string) string {
	if language == "en" {
		if label, ok := enButtons[canonical]; ok {
			return label
		}
		return canonical
	}
	if label, ok := zhButtons[canonical]; ok {
		return label
	}
	return canonical
}

func (m Model) btnLabel(canonical string) string {
	return buttonLabel(m.language, canonical)
}

func (m Model) button(label string) string {
	return button(m.btnLabel(label))
}

func (m Model) activeButton(label string) string {
	return activeButton("▸" + m.btnLabel(label))
}

func (m Model) buttonRow(labels ...string) string {
	parts := make([]string, 0, len(labels))
	for _, label := range labels {
		parts = append(parts, m.button(label))
	}
	return strings.Join(parts, " ")
}

func tr(language, text string) string {
	if language == "en" {
		if translated, ok := enText[text]; ok {
			return translated
		}
		return text
	}
	if translated, ok := zhText[text]; ok {
		return translated
	}
	return text
}

func (m Model) txt(text string) string {
	return tr(m.language, text)
}

func (m Model) viewName(view ViewID) string {
	return m.txt(view.String())
}
