## Why

`modu update` 命令更新 workspace 主项目时，只是 fetch + rebase 到当前分支跟踪的远程分支，没有切换到 `default-base` 配置的分支（如 develop）。这导致 workspace 主项目无法保持最新状态，不利于需求分析。

## What Changes

1. **GitProxy 新增方法** `FetchAndSwitchBranch(repoPath, branch)`：
   - fetch 所有远程
   - 如果分支不存在，从 `origin/<branch>` 创建本地分支
   - 如果分支存在，checkout 到该分支
   - rebase 到 `origin/<branch>`

2. **Engine.UpdateMainProject**：
   - 主项目：使用 `FetchAndSwitchBranch` 切换到 `default-base`
   - 模块：使用 `FetchAndSwitchBranch` 切换到各自的分支（优先使用模块的 `base-branch`，其次全局 `default-base`）

## Capabilities

### Modified Capabilities
- `cli-update`: 更新主项目时，现在会将主项目切换到 default-base 分支

## Impact

- `internal/gitproxy/gitproxy.go` - 新增接口方法
- `internal/gitproxy/gitproxy_impl.go` - 实现 FetchAndSwitchBranch
- `internal/engine/engine.go` - 修改 UpdateMainProject 逻辑
- `internal/engine/engine_test.go` - Mock 适配
