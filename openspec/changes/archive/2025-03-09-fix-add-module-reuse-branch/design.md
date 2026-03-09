## Context

`CreateWorktree` 函数已支持复用已存在的分支，但 `AddModule` 函数（在已有 feature 上添加模块）没有此逻辑。当用户在已有 feature 上添加新模块时，如果远程仓库已存在同名分支，会报错 `module xxx already exists in feature xxx`。

## Goals / Non-Goals

**Goals:**
- 为 AddModule 函数添加分支复用逻辑
- 与 CreateWorktree 保持一致的行为

**Non-Goals:**
- 不修改其他函数

## Decisions

### D1: 复用逻辑实现位置

**选择**: 在 AddModule 函数内部添加检查逻辑

**理由**: 与 CreateWorktree 保持一致的实现方式，复用已有的 GitProxy 方法。

## Risks / Trade-offs

- [风险] 分支被其他 worktree 占用时，跳过可能导致模块不一致 → 输出提示让用户知情
