## Why

当前 `modu create` 命令只创建 git worktree，用户需要在 VSCode 中手动打开多个模块目录或者使用命令行逐个打开。通过在创建 worktree 后自动生成 `.code-workspace` 文件，用户可以直接双击打开包含所有模块的 VSCode 多仓库工作区，提升开发体验。

## What Changes

- 在 `Engine.CreateWorktree` 方法成功后，自动创建 `.code-workspace` 文件
- workspace 文件包含主项目（workspace 根目录）和所有配置的模块
- workspace 文件放置在 feature 目录下，命名为 `{feature}.code-workspace`

## Capabilities

### New Capabilities
- `vscode-workspace-creation`: 在创建 feature worktree 时自动生成 VSCode workspace 文件，包含所有模块的配置

## Impact

- 修改 `internal/engine/engine.go` 的 `CreateWorktree` 方法
- 无新增依赖，使用标准库 JSON 序列化
