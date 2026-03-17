## Why

当前 TUI 程序使用硬编码的深色配色方案（浅灰/青色文字），在白色或浅色终端背景下可读性极差，用户几乎无法看清操作提示和非选中项目。这严重影响用户体验，需要实现自适应系统背景颜色的功能。

## What Changes

- 使用 lipgloss 的 `AdaptiveColor` 替代硬编码颜色值
- 定义深色/浅色两套配色方案：深色保持现有风格，浅色使用深灰前景色 + 浅灰背景色
- 仅修改 UI 样式定义，不涉及业务逻辑

## Capabilities

### New Capabilities
- `tui-theme-adaptive`: TUI 界面根据系统终端背景色自动适配深浅主题

### Modified Capabilities
（无）

## Impact

- 仅影响 `internal/ui/ui.go` 中的样式定义
- 无 API 变更
- 无依赖变更
