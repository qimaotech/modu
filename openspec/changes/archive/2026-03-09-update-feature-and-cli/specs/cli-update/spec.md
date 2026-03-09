# CLI update 子命令规范

**版本**: 1.0 | **来源**: openspec/changes/update-feature-and-cli

## 目的

通过 CLI 子命令对主项目或指定 feature 的 worktree 执行 git fetch + rebase 更新。

## ADDED Requirements

### Requirement: update 子命令存在

CLI SHALL 提供子命令 `modu update`，用于对项目执行更新（fetch + rebase）。

#### Scenario: 无参数更新主项目
- **WHEN** 用户执行 `modu update`（无参数）
- **THEN** 对主项目（workspace 根仓库）及配置中的各模块路径并发执行 fetch + rebase，行为与 TUI 主项目“更新代码”一致

#### Scenario: 无参数更新成功输出
- **WHEN** `modu update` 执行且全部成功
- **THEN** 输出成功摘要（如“更新成功: 主项目”或“更新成功: 主项目 + N 个模块”），退出码 0

#### Scenario: 无参数更新部分失败输出
- **WHEN** `modu update` 执行且部分仓库失败
- **THEN** 输出成功数量与失败名称列表，退出码非 0

### Requirement: 指定 feature 更新

CLI SHALL 支持 `modu update <feature>`，对指定 feature 的 worktree 执行更新。

#### Scenario: 带 feature 参数更新
- **WHEN** 用户执行 `modu update <feature>`
- **THEN** 对该 feature 目录（主项目 worktree）及其下存在的各模块 worktree 并发执行 fetch + rebase

#### Scenario: feature 不存在
- **WHEN** 用户执行 `modu update <feature>` 且该 feature 目录不存在
- **THEN** 报错并退出码非 0，不执行更新

#### Scenario: 指定 feature 更新成功输出
- **WHEN** `modu update <feature>` 执行且全部成功
- **THEN** 输出成功摘要（如“更新成功: feature <name>（主项目 + N 个模块）”），退出码 0
