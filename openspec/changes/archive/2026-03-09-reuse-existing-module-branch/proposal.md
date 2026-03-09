## Why

当前 `modu create feature` 在创建 feature 时，会为每个 module 都创建同名分支。但当某个 module 的远程仓库已存在同名分支时（其他同事已创建），会创建失败。需要支持复用已存在的分支，减少冲突和重复操作。

## What Changes

- **新增** GitProxy 方法：检查分支是否已被 worktree 使用
- **修改** CreateWorktree 逻辑：module 分支存在时，复用现有分支而非创建新分支
- **新增** 跳过逻辑：分支已被其他 worktree 使用时，跳过该 module 并输出提示
- **新增** 结果 summary：显示创建成功和跳过的 module 数量

## Capabilities

### New Capabilities
- `reuse-existing-branch`: 支持复用 module 仓库中已存在的分支创建 worktree

### Modified Capabilities
- (无)

## Impact

- 影响代码：`internal/gitproxy/gitproxy.go`, `internal/engine/engine.go`
- 无 API 变更
- 无依赖变更
