# modu Git 原语规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## 目的

封装 Git 命令调用，屏蔽 OS 执行细节，统一 stderr 解析与错误包装，供 engine 使用。

## 接口职责（GitClient）

- **Clone(ctx, url, path, options)**：先克隆到同级临时目录，成功后移动到指定路径；失败或取消清理本次临时目录，保留已有目标目录。默认完整克隆，可选 `Filter: "blob:none"`；失败返回带 `ERR_GIT_EXEC` 的上下文错误。
- **CreateWorktree(ctx, repoPath, branch, baseBranch, worktreePath)**：在 repoPath 仓库中先 fetch，再 `worktree add -b <branch> <worktreePath> <baseBranch>`；失败返回带上下文的 `ERR_GIT_EXEC`。
- **CreateWorktreeFromRemoteBranch(ctx, repoPath, branch, worktreePath)**：显式将 `refs/heads/<branch>` 拉取到 `origin/<branch>`，并创建同名本地 tracking 分支的 worktree；必须兼容 single-branch clone 或受限的 remote fetch refspec。
- **GetStatus(ctx, path)**：在 path 执行一次 `git status --porcelain=v2 --branch -z`，同时解析分支、脏状态与文件路径；目录不存在返回 `ERR_MODULE_NOT_FOUND`。
- **RemoveWorktree(ctx, path)**：`git worktree remove <path>`；若 remove 失败可回退为 `os.RemoveAll(path)`（实现可选）。
- **RemoveWorktreeAndBranch(ctx, repoPath, worktreePath, featureDirName)**：在移除 worktree **之前**对 `worktreePath` 调用 `GetStatus` 取得当前检出分支；仅当将该分支名中的 `/` 全部替换为 `-` 后的字符串与 `featureDirName`（与 `worktree-root` 下该 feature 的目录 basename 一致）相同时，才在 `repoPath` 上对该分支执行 `git branch -D`。若无法读状态、detached HEAD（`HEAD`）、或不一致，则仍执行 worktree remove / prune，但**不删除分支**（防误删）。`featureDirName` 规则与引擎侧「分支名 → 目录名」转换一致（`/ → -`）。
- **FetchAndSwitchBranch(ctx, repoPath, branch)**：在 repoPath 仓库中先执行 `git fetch origin` 拉取最新，再执行 `git checkout <branch>` 切换到指定分支；若分支不存在返回错误；若切换失败返回带 `ERR_GIT_EXEC` 的上下文错误。
- **GetBranchPushStatus(ctx, repoPath, branch)**：检查本地分支是否已完整包含在远端分支中；优先使用 upstream，无法解析时回退 `origin/<branch>`。

## Status 解析

- porcelain v2 的文件记录（含修改、未跟踪、删除及冲突）视为 **Dirty**；仅含分支头记录时为 **Clean**。
- 以 NUL 分隔记录，保留空格和换行文件名，并正确处理重命名与 detached HEAD。

## 错误约定

- 所有返回错误须包含：操作类型、路径/仓库/分支等上下文、原始 stderr（或摘要）与命令/上下文失败原因，并 wrap `ERR_GIT_EXEC` 或 `ERR_MODULE_NOT_FOUND`；stderr 中的 SSH warning 不得掩盖真实失败原因。

### Requirement: Remote branch worktree creation
GitProxy SHALL explicitly fetch a selected remote branch and create a usable local tracking branch for its worktree.

#### Scenario: Clone has a restricted fetch refspec
- **WHEN** the remote branch exists but the local clone's configured fetch refspec does not include it
- **THEN** GitProxy fetches that exact branch into `origin/<branch>` before creating the worktree

#### Scenario: Remote branch worktree is created
- **WHEN** the exact remote branch fetch succeeds
- **THEN** the worktree checks out a local `<branch>` whose upstream is `origin/<branch>`, rather than detached HEAD

#### Scenario: Fetch is canceled after emitting an SSH warning
- **WHEN** fetch stderr contains an SSH warning and the command is canceled or otherwise fails
- **THEN** the returned error includes both `ERR_GIT_EXEC` and the actual command or context failure cause

## Requirements

### Requirement: Branch push state inspection
GitProxy SHALL expose an operation that determines whether a local branch is fully contained in a remote branch before deletion.

#### Scenario: Branch has no commits ahead of remote
- **WHEN** the local branch resolves to a remote ref and `remote..branch` has an ahead count of zero
- **THEN** GitProxy reports the branch as pushed

#### Scenario: Branch is ahead of remote
- **WHEN** the local branch resolves to a remote ref and `remote..branch` has an ahead count greater than zero
- **THEN** GitProxy reports the branch as not pushed with the ahead count

#### Scenario: Remote branch cannot be resolved
- **WHEN** the local branch has no upstream and no fallback `origin/<branch>` remote ref
- **THEN** GitProxy reports the branch as not pushed

#### Scenario: Push-state Git command fails
- **WHEN** fetching or push-state inspection fails
- **THEN** GitProxy returns a contextual error and the engine treats the branch as requiring confirmation

## 与代码的对应

- 实现：`internal/gitproxy`（GitProxy 实现、Clone/CreateWorktree/GetStatus/RemoveWorktree/RemoveWorktreeAndBranch、parseStatus）。
