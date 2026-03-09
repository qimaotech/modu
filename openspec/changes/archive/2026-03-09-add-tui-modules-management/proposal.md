## Why

当前 TUI 只能管理整个 feature（查看、打开 VS Code、删除），但无法对 feature 下的单个模块进行管理。用户希望在 TUI 中能够为某个 feature 动态添加或删除模块，无需使用 CLI 命令。

## What Changes

- 在列表视图新增按 `m` 键直接进入 Modules 管理功能（快速入口）
- 在操作菜单中将 "Modules 管理" 作为第二个选项（顺序：打开 VS Code → Modules 管理 → 删除）
- 新增 Modules 管理视图，复用现有的 `ModuleSelector` 组件逻辑，显示配置中所有模块及其在当前 feature 中的状态
- 支持通过空格键切换模块的选中/未选中状态
- 回车确认后执行模块的增删操作：
  - 新增模块：调用 Engine 创建模块 worktree
  - 删除模块：调用 Engine 删除模块 worktree
- Engine 层新增 `AddModule` 和 `RemoveModule` 方法支持单模块操作

## Capabilities

### New Capabilities
- `tui-module-management`: TUI 模块管理功能，支持在 UI 中查看、添加、删除 feature 下的模块

### Modified Capabilities
- `tui-operation-menu`: 扩展操作菜单，增加 Modules 管理入口

## Impact

- **代码影响**:
  - `internal/ui/ui.go`: 新增 modules 状态和视图，修改操作菜单
  - `internal/engine/engine.go`: 新增 AddModule/RemoveModule 方法
- **无 API 变更**
- **无依赖变更**
