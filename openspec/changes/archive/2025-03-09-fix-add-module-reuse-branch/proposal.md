## Why

之前实现了 `CreateWorktree` 支持复用已存在的分支，但 `AddModule` 函数（为已有 feature 添加新模块）没有同样的逻辑，导致在已有 feature 上添加模块时，如果远程仓库已存在同名分支，会报错失败。

## What Changes

- 修改 `AddModule` 函数：添加分支存在检查逻辑
- 分支存在 + 未被 worktree 使用 → 复用现有分支创建 worktree
- 分支存在 + 已被 worktree 使用 → 输出跳过提示并返回成功

## Capabilities

### New Capabilities
- (无)

### Modified Capabilities
- `reuse-existing-branch`: 扩展复用逻辑，支持 AddModule 场景

## Impact

- 影响代码：`internal/engine/engine.go` 的 AddModule 函数
- 无 API 变更
- 无依赖变更
