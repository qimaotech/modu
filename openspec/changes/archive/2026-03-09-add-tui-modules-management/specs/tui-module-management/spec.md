# TUI 模块管理规范

**版本**: 1.0 | **来源**: proposal

## 目的

在 TUI 中提供模块管理功能，允许用户查看、添加、删除 feature 下的模块。

## ADDED Requirements

### Requirement: 列表视图按 m 进入模块管理
用户 MUST 可以在列表视图按 `m` 键直接进入模块管理视图。

#### Scenario: 快速进入模块管理
- **WHEN** 用户在列表视图按 `m` 键
- **THEN** TUI 进入模块管理视图，显示当前选中 feature 的模块列表

### Requirement: 操作菜单显示模块管理选项
操作菜单 MUST 显示"Modules 管理"选项。

#### Scenario: 操作菜单显示模块管理
- **WHEN** 用户按 Enter 进入操作菜单
- **THEN** 显示三个选项：打开 VS Code、Modules 管理、删除（从上到下顺序）

### Requirement: 操作菜单进入模块管理
用户 MUST 可以在操作菜单中选中"Modules 管理"并进入。

#### Scenario: 从操作菜单进入模块管理
- **WHEN** 用户在操作菜单选中"Modules 管理"项并按 Enter
- **THEN** TUI 进入模块管理视图

### Requirement: 模块列表显示
模块管理视图 MUST 显示配置中的所有模块及其状态。

#### Scenario: 显示已创建的模块
- **WHEN** 模块已在 feature 中创建
- **THEN** 显示 `[x] 模块名`

#### Scenario: 显示未创建的模块
- **WHEN** 模块未在 feature 中创建
- **THEN** 显示 `[ ] 模块名`

### Requirement: 模块选择切换
用户 MUST 可以使用空格键切换模块的选中状态。

#### Scenario: 选中未创建的模块
- **WHEN** 用户在未创建的模块上按空格键
- **THEN** 该模块标记为选中状态 `[x]`

#### Scenario: 取消选中已创建的模块
- **WHEN** 用户在已创建的模块上按空格键
- **THEN** 该模块标记为未选中状态 `[ ]`

### Requirement: 模块管理导航
用户 MUST 可以在模块列表中使用上下键移动光标。

#### Scenario: 向上导航
- **WHEN** 用户在模块列表按向上键且不是第一项
- **THEN** 光标向上移动一项

#### Scenario: 向下导航
- **WHEN** 用户在模块列表按向下键且不是最后一项
- **THEN** 光标向下移动一项

### Requirement: 模块管理确认执行
用户 MUST 可以按回车键确认执行选中的模块操作。

#### Scenario: 确认添加模块
- **WHEN** 用户选中未创建的模块并按 Enter
- **THEN** TUI 调用 Engine 创建选中模块的 worktree，操作完成后刷新列表

#### Scenario: 确认删除模块
- **WHEN** 用户选中已创建的模块并按 Enter
- **THEN** TUI 调用 Engine 删除选中模块的 worktree，操作完成后刷新列表

### Requirement: 退出模块管理
用户 MUST 可以按 esc 或 q 退出模块管理视图，返回操作菜单。

#### Scenario: 退出模块管理
- **WHEN** 用户在模块管理视图按 esc 或 q
- **THEN** TUI 返回操作菜单
