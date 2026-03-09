# TUI feature 更新规范

**版本**: 1.0 | **来源**: openspec/changes/update-feature-and-cli

## 目的

在 TUI 中选中 feature 时，支持通过快捷键或菜单执行“更新代码”，对该 feature 的 worktree 执行 fetch + rebase。

## ADDED Requirements

### Requirement: feature 菜单包含更新代码

当用户选中 feature 并进入操作菜单时，TUI SHALL 提供“更新代码”选项。

#### Scenario: feature 菜单显示更新代码
- **WHEN** 用户在 feature 条目上按 Enter 进入菜单
- **THEN** 菜单包含“更新代码 (u)”选项（与打开 VS Code、Modules 管理、删除并列或按设计顺序）

#### Scenario: feature 菜单选择更新代码
- **WHEN** 用户在 feature 菜单中选中“更新代码”并按 Enter 或 u
- **THEN** TUI 进入 loading 状态，对当前选中的 feature worktree（主项目 + 该 feature 下存在的模块）执行 fetch + rebase，完成后返回列表视图并显示结果消息

### Requirement: feature 列表 u 键更新

在列表视图中选中 feature 时，用户 MUST 可通过 u 键直接触发该 feature 的更新。

#### Scenario: 列表中按 u 更新 feature
- **WHEN** 用户选中某 feature 且在列表视图下按 u
- **THEN** 对该 feature worktree 执行更新（同菜单“更新代码”），进入 loading 后显示结果并刷新列表

#### Scenario: 选中主项目时 u 仍仅更新主项目
- **WHEN** 用户选中主项目条目并按 u
- **THEN** 行为与现有主项目更新一致，不调用 feature 更新

### Requirement: 更新完成后刷新

feature 更新完成后，TUI SHALL 刷新列表并展示结果消息。

#### Scenario: 更新成功后的消息
- **WHEN** feature 更新全部成功
- **THEN** 显示成功摘要（如“更新成功: feature <name>（主项目 + N 个模块）”），列表数据刷新

#### Scenario: 更新部分失败后的消息
- **WHEN** feature 更新部分仓库失败
- **THEN** 显示成功数量与失败名称，列表数据刷新
