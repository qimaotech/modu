## Context

当前 `modu create feature` 执行时，会为每个 module 创建同名的 git 分支。当某个 module 的远程仓库已存在同名分支时（例如其他同事已创建），会创建失败并报错。

需求：支持复用已存在的分支，减少重复操作和冲突。

## Goals / Non-Goals

**Goals:**
- 支持 module 分支已存在时，复用该分支创建 worktree
- 分支已被其他 worktree 使用时，跳过该 module 并输出明确提示
- 主项目保持现有逻辑（始终创建新分支）

**Non-Goals:**
- 不修改主项目的分支创建逻辑
- 不支持强制移除其他 worktree 后使用

## Decisions

### D1: 分支状态检查逻辑

**选择**: 在 GitProxy 层新增 `CheckBranchWorktreeStatus` 方法

**理由**: 将 git 操作封装在 gitproxy 包内，保持 Engine 层的业务逻辑清晰。

**备选方案**:
- 在 Engine 层直接调用 ListWorktrees + 解析 → 缺点是暴露 git 操作细节

### D2: 分支已存在的处理策略

**选择**: 三级处理
1. 分支不存在 → 创建新分支（现有逻辑）
2. 分支存在 + 未被 worktree 使用 → 直接 checkout 现有分支
3. 分支存在 + 已被 worktree 使用 → 跳过该 module

**理由**: 符合用户需求的优先级，复用 > 跳过 > 报错。

### D3: 跳过处理

**选择**: 跳过时记录日志，继续处理其他 module，最终输出 summary

**理由**: 部分成功应该被视为成功，避免因单个模块问题导致整体失败。

## Risks / Trade-offs

- [风险] 分支被其他 worktree 占用时，跳过可能导致各 module 分支不一致 → 文档说明用户需手动处理
- [风险] 远程分支可能落后于 base 分支 → 用户需要手动处理合并
