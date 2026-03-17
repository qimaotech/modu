# CLI update 子命令规范

**版本**: 1.1 | **来源**: openspec/changes/update-switch-to-default-base

## 目的

通过 CLI 子命令对主项目或指定 feature 的 worktree 执行 git fetch + rebase 更新。

## ADDED Requirements

### Requirement: update 子命令存在

CLI SHALL 提供子命令 `modu update`，用于对项目执行更新（fetch + rebase）。

#### Scenario: 无参数更新主项目
- **WHEN** 用户执行 `modu update`（无参数）
- **THEN** 对主项目（workspace 根仓库）及配置中的各模块路径并发执行 fetch + rebase，行为与 TUI 主项目"更新代码"一致

#### Scenario: 无参数更新成功输出
- **WHEN** `modu update` 执行且全部成功
- **THEN** 输出成功摘要（如"更新成功: 主项目"或"更新成功: 主项目 + N 个模块"），退出码 0

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
- **THEN** 输出成功摘要（如"更新成功: feature <name>（主项目 + N 个模块）"），退出码 0

## MODIFIED Requirements

### Requirement: 主项目更新时切换到 default-base 分支

**原描述**: 对主项目（workspace 根仓库）及配置中的各模块路径并发执行 fetch + rebase

**新描述**: 对主项目（workspace 根仓库）：
1. 调用 GitProxy.FetchAndSwitchBranch(repoPath, defaultBaseBranch)，其中 defaultBaseBranch 来自配置文件的 default-base 字段
2. 如果 default-base 未配置，跳过切换步骤，执行原有的 fetch + rebase
对各模块路径：
1. 使用模块的 base-branch（如果配置了），否则使用全局 default-base
2. 调用 GitProxy.FetchAndSwitchBranch(repoPath, branch) 切换到对应分支

#### Scenario: 主项目更新切换到 default-base 分支
- **WHEN** 用户执行 `modu update` 且配置文件存在 default-base 配置（如 "develop"）
- **THEN** 主项目执行 fetch + checkout 到 origin/develop + rebase 到 origin/develop

#### Scenario: 主项目更新无 default-base 配置
- **WHEN** 用户执行 `modu update` 且配置文件不存在 default-base 配置
- **THEN** 主项目执行原有的 fetch + rebase 到当前分支跟踪的远程分支

#### Scenario: 主项目更新切换分支时本地有修改
- **WHEN** 用户执行 `modu update` 且主项目本地有未提交的修改
- **THEN** 报错并退出码非 0，不执行切换

#### Scenario: 远程 default-base 分支不存在
- **WHEN** 用户执行 `modu update` 且远程不存在 default-base 分支（如 origin/develop）
- **THEN** 报错并退出码非 0，提示远程分支不存在

### Requirement: 子模块更新时切换到对应分支

**描述**: 对各模块路径，使用模块的 base-branch（如果配置了）或全局 default-base 切换到对应分支

#### Scenario: 子模块使用模块配置的 base-branch
- **WHEN** 用户执行 `modu update` 且模块配置了 base-branch（如 "release/v1.0"）
- **THEN** 该模块执行 fetch + checkout 到 origin/release/v1.0 + rebase 到 origin/release/v1.0

#### Scenario: 子模块使用全局 default-base
- **WHEN** 用户执行 `modu update` 且模块未配置 base-branch
- **THEN** 该模块执行 fetch + checkout 到 origin/develop + rebase 到 origin/develop（使用全局 default-base）
