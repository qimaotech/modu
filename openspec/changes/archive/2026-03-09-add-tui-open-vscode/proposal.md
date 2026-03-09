## Why

modu TUI 目前仅支持删除 feature 功能，用户需要手动查找并打开项目目录。使用 TUI 选择 feature 后直接打开 VS Code，可以提升工作效率，减少上下文切换。

## What Changes

- 在 TUI 列表视图添加 `o` 快捷键支持
- 选中 feature 后按 `o` 调用 `code` 命令打开主项目目录
- 如果 feature 无主项目，显示错误提示

## Capabilities

### New Capabilities
此变更不涉及新的 capability，是对现有 TUI 交互的增强。

### Modified Capabilities
无。现有 specs 不需要修改。

## Impact

- **代码影响**: `internal/ui/ui.go` - 添加键盘事件处理和 VS Code 启动逻辑
- **无新依赖**: 使用标准库 `os/exec` 执行系统命令
- **向后兼容**: 现有交互方式不变
