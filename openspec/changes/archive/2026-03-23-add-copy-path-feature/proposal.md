## Why

在日常开发中，经常需要获取 feature worktree 的主项目路径用于其他终端操作。目前只能通过 VS Code 打开后复制路径，操作繁琐。增加复制路径功能可以提升效率。

## What Changes

- TUI 操作菜单增加"复制路径"功能，快捷键 `c`
- 支持在菜单视图和列表视图直接触发
- 复制成功后显示临时消息提示用户
- 使用跨平台剪贴板库实现 macOS/Linux 支持

## Capabilities

### New Capabilities

- `tui-copy-path`: 在 TUI 操作菜单中提供复制主项目绝对路径的功能，支持快捷键触发和临时消息反馈

### Modified Capabilities

- (无)

## Impact

- **代码修改**: `internal/ui/ui.go` — 增加菜单项和快捷键处理
- **依赖**: 引入 `github.com/atotto/clipboard` 跨平台剪贴板库
- **用户体验**: 列表视图和菜单视图均支持 `c` 键触发
