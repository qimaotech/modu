# modu Git 原语规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## 目的

封装 Git 命令调用，屏蔽 OS 执行细节，统一 stderr 解析与错误包装，供 engine 使用。

## 接口职责（GitClient）

- **Clone(ctx, url, path)**：克隆仓库到指定路径；失败返回带 `ERR_GIT_EXEC` 的上下文错误。
- **CreateWorktree(ctx, repoPath, branch, baseBranch, worktreePath)**：在 repoPath 仓库中先 fetch，再 `worktree add -b <branch> <worktreePath> <baseBranch>`；失败返回带上下文的 `ERR_GIT_EXEC`。
- **GetStatus(ctx, path)**：在 path 执行 `git status --porcelain`，解析为 Status（IsDirty、Branch）；目录不存在返回 `ERR_MODULE_NOT_FOUND`。
- **RemoveWorktree(ctx, path)**：`git worktree remove <path>`；若 remove 失败可回退为 `os.RemoveAll(path)`（实现可选）。
- **RemoveWorktreeAndBranch(ctx, repoPath, branch, worktreePath)**：先 remove worktree，再在 repoPath 删除分支 branch（若存在）。

## Status 解析

- `git status --porcelain` 有输出（含 M、??、D 等）视为 **Dirty**；无输出视为 **Clean**。
- Branch 可通过 `git rev-parse --abbrev-ref HEAD` 或等价方式获取（在 path 下执行）。

## 错误约定

- 所有返回错误须包含：操作类型、路径/仓库/分支等上下文、原始 stderr（或摘要），并 wrap `ERR_GIT_EXEC` 或 `ERR_MODULE_NOT_FOUND`。

## 与代码的对应

- 实现：`internal/gitproxy`（GitProxy 实现、Clone/CreateWorktree/GetStatus/RemoveWorktree/RemoveWorktreeAndBranch、parseStatus）。
