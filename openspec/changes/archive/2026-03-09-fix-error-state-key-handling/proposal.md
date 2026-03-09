## Why

TUI 在显示错误页面时，用户按任意键无法继续操作，程序卡死。这是因为 `ui.go` 中 error 状态的按键处理缺失，导致用户体验受阻。

## What Changes

- 在 `internal/ui/ui.go` 的按键消息处理中，为 `error` 状态添加按键监听
- 用户按任意键后可返回到 list 状态继续操作

## Capabilities

### New Capabilities
无

### Modified Capabilities
无

## Impact

- 修改文件: `internal/ui/ui.go`
- 影响范围: TUI 错误页面交互
