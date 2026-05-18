## Why

`modu delete` and TUI deletion remove both worktrees and matching local branches. If a branch has local commits that have not reached a remote branch, deletion can permanently discard work that is not recoverable from the remote.

Deletion already protects dirty worktrees, but clean committed local-only work is still at risk. The delete flow needs an explicit pushed-state check before removing local branches.

## What Changes

- Before deleting a local branch as part of feature or module worktree removal, determine whether the branch is fully pushed to its tracked or remote branch.
- If every branch that would be removed is already pushed, delete proceeds without extra confirmation.
- If any branch is not pushed, CLI deletion fails with a dedicated unpushed-branch error unless `--allow-unpushed` is provided.
- TUI deletion includes a risk confirmation view that lists unpushed branches before executing deletion.
- `--force` continues to skip dirty checks only; it does not silently bypass unpushed-branch protection.

## Capabilities

### New Capabilities

### Modified Capabilities

- `cli`: `modu delete` must guard local branch deletion when commits are not pushed and require `--allow-unpushed` to continue.
- `engine`: delete and module-removal flows must preflight branch push state before destructive removal.
- `gitproxy`: Git primitives must expose branch push-state inspection for local branches.
- `tui`: TUI delete must require an additional confirmation when unpushed branches would be removed.
- `errors`: add a machine-readable error for deletion blocked by unpushed local branches.

## Impact

- `cmd/modu/main.go`: CLI delete confirmation and non-interactive behavior.
- `internal/engine`: delete preflight orchestration for features and modules.
- `internal/gitproxy`: branch push-state Git commands and tests.
- `internal/ui`: risk confirmation state and rendering.
- `internal/errors` and `internal/output`: error code and formatted failure behavior.
- Tests in `internal/engine`, `internal/gitproxy`, `internal/ui`, and CLI/output areas.
