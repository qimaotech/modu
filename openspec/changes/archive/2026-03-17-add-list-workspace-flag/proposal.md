## Why

目前 `modu list` 命令只显示 feature worktrees，用户无法看到主项目（workspace）及其模块的分支情况。当项目多时，用户不知道分支是否已经切换过去，也无法确认主项目的各个模块是否在正确的分支上。

## What Changes

- 给 `modu list` 命令添加 `-a` / `--all` flag
- 使用 `-a` 时显示主项目（workspace）及其所有模块的分支信息
- Workspace 信息显示在 Features 列表上方
- 输出格式：`Workspace [<分支名>]` 后面列出所有模块及其分支

## Capabilities

### New Capabilities
- `cli-list-workspace`: 在 list 命令中显示主项目及其模块的分支信息

### Modified Capabilities
无

## Impact

- 修改 `cmd/modu/main.go` - 添加 `-a` flag
- 修改 `internal/output/output.go` - 添加主项目信息的格式化方法
- 可能需要修改 `internal/engine/engine.go` - 添加获取主项目模块状态的方法
