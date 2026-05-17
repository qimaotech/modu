# modu Git 原语规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## Purpose

封装 Git 命令调用，屏蔽 OS 执行细节，统一 stderr 解析与错误包装，向 engine 提供可测试、可识别失败阶段的 Git 原语。

## 接口职责（GitClient）

- **Clone(ctx, url, path)**：克隆仓库到指定路径；失败返回带 `ERR_GIT_EXEC` 的上下文错误。
- **CreateWorktree(ctx, repoPath, branch, baseBranch, worktreePath)**：在 repoPath 仓库中先 fetch，再 `worktree add -b <branch> <worktreePath> <baseBranch>`；失败返回带上下文的 `ERR_GIT_EXEC`。
- **CreateWorktreeNoFetch(ctx, repoPath, branch, baseBranch, worktreePath)**：不执行 fetch，直接 `worktree add -b <branch> <worktreePath> <baseBranch>`；失败返回带上下文的 `ERR_GIT_EXEC`。
- **GetStatus(ctx, path)**：在 path 执行 `git status --porcelain`，解析为 Status（IsDirty、Branch）；目录不存在返回 `ERR_MODULE_NOT_FOUND`。
- **RemoveWorktree(ctx, path)**：`git worktree remove <path>`；若 remove 失败可回退为 `os.RemoveAll(path)`（实现可选）。
- **RemoveWorktreeAndBranch(ctx, repoPath, worktreePath, featureDirName)**：在移除 worktree **之前**对 `worktreePath` 调用 `GetStatus` 取得当前检出分支；仅当将该分支名中的 `/` 全部替换为 `-` 后的字符串与 `featureDirName`（与 `worktree-root` 下该 feature 的目录 basename 一致）相同时，才在 `repoPath` 上对该分支执行 `git branch -D`。若无法读状态、detached HEAD（`HEAD`）、或不一致，则仍执行 worktree remove / prune，但**不删除分支**（防误删）。`featureDirName` 规则与引擎侧「分支名 → 目录名」转换一致（`/ → -`）。
- **FetchAndSwitchBranch(ctx, repoPath, branch)**：在 repoPath 仓库中先执行 `git fetch origin` 拉取最新，再执行 `git checkout <branch>` 切换到指定分支；若分支不存在返回错误；若切换失败返回带 `ERR_GIT_EXEC` 的上下文错误。
- **CreateWorktreeFromExistingBranchNoFetch(ctx, repoPath, branch, worktreePath)**：不执行 fetch，直接从现有本地分支创建 worktree。

## Status 解析

- `git status --porcelain` 有输出（含 M、??、D 等）视为 **Dirty**；无输出视为 **Clean**。
- Branch 可通过 `git rev-parse --abbrev-ref HEAD` 或等价方式获取（在 path 下执行）。

## 错误约定

- 所有返回错误须包含：操作类型、路径/仓库/分支等上下文、原始 stderr（或摘要），并 wrap `ERR_GIT_EXEC` 或 `ERR_MODULE_NOT_FOUND`。
- pre-create fetch 失败 MUST 可通过 gitproxy 的 fetch 错误识别函数区分，同时保留 `ERR_GIT_EXEC` wrapping。

## Requirements

### Requirement: Worktree creation can skip fetch

Gitproxy SHALL provide worktree creation behavior that can add a worktree without performing a preceding fetch when the caller explicitly requests local-only creation.

#### Scenario: Create new worktree without fetch

- **WHEN** gitproxy is asked to create a worktree with fetch disabled
- **THEN** it runs `git worktree add -b <branch> <worktreePath> <baseBranch>` without first running `git fetch`

#### Scenario: Reuse existing branch without fetch

- **WHEN** gitproxy is asked to create a worktree from an existing local branch with fetch disabled
- **THEN** it runs `git worktree add <worktreePath> <branch>` without first running `git fetch`

#### Scenario: Fetch failure remains distinguishable

- **WHEN** pre-create fetch fails before any `git worktree add` command runs
- **THEN** gitproxy returns an error that callers can identify as a fetch failure while still wrapping `ERR_GIT_EXEC`

#### Scenario: Worktree add failure keeps context

- **WHEN** local-only worktree creation fails
- **THEN** gitproxy returns an error containing the operation, repository path, target branch or base branch, worktree path, and git output

## 与代码的对应

- 实现：`internal/gitproxy`（GitProxy 实现、Clone/CreateWorktree/GetStatus/RemoveWorktree/RemoveWorktreeAndBranch、parseStatus）。
