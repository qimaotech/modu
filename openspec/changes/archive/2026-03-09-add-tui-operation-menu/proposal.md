## Why

当前 TUI 的删除操作流程繁琐：用户需要先按 Enter 进入删除确认界面，才能进行二次确认删除。用户体验不流畅，且缺乏直接的操作入口。用户希望在列表视图就能通过快捷键直接触发删除操作（仍需二次确认）。

## What Changes

- 列表视图按 Enter 进入操作菜单（删除 / 打开 VS Code）
- 列表视图按 d 直接触发删除确认
- 操作菜单内按 d 触发删除确认，按 o 打开 VS Code
- 删除确认流程保持不变（y/n 二次确认）
- 保留列表视图直接按 o 打开 VS Code 的现有功能

## Capabilities

### New Capabilities
- `tui-operation-menu`: TUI 操作菜单，支持在列表视图通过菜单选择操作

### Modified Capabilities
- 无

## Impact

- 影响代码：`internal/ui/ui.go`
- 无新增依赖
- 用户交互方式变化：新增操作菜单和删除快捷键
