## Context

`GitProxy.RemoveWorktreeAndBranch` currently removes the worktree and then force-deletes the matching local branch when the branch slug matches the feature directory name. `Engine.DeleteWorktree` calls this for every configured module and the main project. The TUI calls the same engine method after its existing delete confirmation.

Dirty worktree checks protect uncommitted files, but they do not protect committed local work that has not been pushed. The existing branch slug guard prevents deleting the wrong branch, but it does not verify that the branch is recoverable from a remote.

## Goals / Non-Goals

**Goals:**

- Detect local branches that would be deleted and are ahead of, or missing from, their remote branch.
- Require explicit `--allow-unpushed` confirmation in CLI before deleting those branches.
- Keep TUI's second confirmation before deleting those branches.
- Keep deletion behavior unchanged when all candidate branches are fully pushed.
- Keep rollback cleanup from failed create operations working without user prompts.

**Non-Goals:**

- Automatically push branches.
- Change the feature-name to directory-name slug rule.
- Protect non-branch detached HEAD worktrees beyond the existing "skip branch delete" behavior.
- Introduce new Git remotes or infer URLs outside Git's existing local remote configuration.

## Decisions

### Decision 1: Engine-level preflight, not gitproxy-level blocking

Deletion should collect branch risk before removing any worktree. The engine will expose a delete preflight that returns candidate branches whose local commits are not confirmed on the remote. CLI and TUI can then ask the human once before destructive work starts.

Alternative considered: make `RemoveWorktreeAndBranch` refuse to delete unpushed branches. This would also affect rollback paths where newly-created local branches are intentionally cleaned up after a failed create, and it would discover risk one repository at a time after deletion had already started.

### Decision 2: Explicit delete options after confirmation

`DeleteWorktree` keeps the safe default and blocks when unpushed branches exist. A new options-based deletion path allows CLI to pass `AllowUnpushedBranches` only when `--allow-unpushed` is provided, and allows TUI to pass it only after the second confirmation.

Alternative considered: let `--force` bypass the check. Rejected because the current `--force` meaning is limited to dirty checks; silently expanding it to cover committed unpushed work would be surprising.

### Decision 3: Ahead-count push status

Git push status is determined by fetching remotes, resolving the branch upstream when available, otherwise falling back to `origin/<branch>`, and checking the ahead count with `remoteRef..branch`. Ahead count `0` means the local branch is safe to delete. Missing upstream/remote, fetch failure, or ahead count greater than zero means the branch requires confirmation.

Alternative considered: only check whether a remote branch exists. Rejected because a remote branch can exist while local HEAD still contains newer unpushed commits.

### Decision 4: TUI adds a second risk confirmation state

The current TUI delete confirmation remains the first confirmation. After the user confirms deletion, the TUI runs the unpushed-branch preflight. If risk exists, the TUI shows the affected project/modules and requires a second `y` before deletion.

Alternative considered: include all risk details in the original confirmation view. Rejected because the risk list requires Git inspection and may be slow; keeping it after the first confirmation avoids doing network work while merely browsing.

## Risks / Trade-offs

- [Risk] Fetching remotes during delete may add latency. -> Mitigation: only run the check after the user initiates deletion and keep the result list concise.
- [Risk] Repositories with non-`origin` remotes and no upstream may be treated as unpushed. -> Mitigation: this is conservative and requires confirmation rather than allowing silent deletion.
- [Risk] CLI users may expect an interactive prompt. -> Mitigation: return `ERR_UNPUSHED_BRANCH` with branch details and tell callers to rerun with `--allow-unpushed`.
- [Risk] TUI module-management removal also uses branch deletion. -> Mitigation: engine defaults block unpushed module deletion; feature delete gets the richer TUI confirmation in this change.
