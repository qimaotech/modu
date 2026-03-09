# Proposal: tui-main-project

## Why

当前 modu TUI 只显示 feature 分支列表，用户无法直接在 TUI 中查看和操作 workspace 下的主项目（主干代码）。用户需要在 TUI 中直接打开主项目的 VS Code，以及一键更新主项目和所有模块的代码。

## What Changes

1. **TUI 列表增加主项目条目**：在列表顶部固定显示主项目（workspace 目录），包含 dirty 状态和当前分支
2. **主项目专用菜单**：主项目使用简化菜单，仅包含"打开 VS Code"和"更新代码"两个操作
3. **更新代码功能**：新增 `u` 快捷键，执行 git fetch + rebase，同时更新主项目和所有模块
4. **Engine 层新增方法**：`UpdateMainProject(ctx)` 并发更新主项目 + 所有 modules

## Capabilities

### New Capabilities

- **tui-main-project**: 在 TUI 列表中展示主项目（workspace 主仓库），支持打开 VS Code 和一键更新代码

### Modified Capabilities

- **tui-operation-menu**: 修改菜单逻辑，feature 和主项目使用不同的菜单选项

## Impact

- 修改 `internal/ui/ui.go`：增加主项目展示和菜单逻辑
- 修改 `internal/engine/engine.go`：新增 `UpdateMainProject` 方法
- 新增 `openspec/specs/tui-main-project/spec.md`：定义主项目相关的 TUI 行为规范
