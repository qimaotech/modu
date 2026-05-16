## Why

Delete backups now protect users from accidental content loss, but recovery is still manual and easy to get wrong because a raw archive extraction does not re-register Git worktrees. Users need a supported way to find delete backups and restore a deleted feature worktree safely.

## What Changes

- Add a backup listing capability for generated delete backup archives under `<worktree-root>/.modu/backups/`.
- Add a restore flow that recreates the feature worktree structure and overlays archived content while avoiding stale `.git` metadata from the archive.
- Add CLI commands to list and restore delete backups manually.
- Add a TUI operation for choosing a backup and triggering restore manually.
- Preserve existing delete backup retention cleanup behavior while ensuring backup list/restore commands do not delete candidates before the user can inspect or restore them.

## Capabilities

### New Capabilities

- `delete-backup-restore`: Listing and restoring delete backups created by `modu delete`.

### Modified Capabilities

- `engine`: Engine behavior gains backup enumeration and restore semantics shared by CLI and TUI.
- `cli`: CLI gains backup list/restore commands and startup cleanup exceptions for backup-management commands.
- `output`: Output format gains backup list/restore responses for text and JSON modes.
- `tui`: TUI gains a manual restore entry point and backup selection flow.

## Impact

- Affected code: `internal/engine`, `cmd/modu`, `internal/output`, `internal/ui`, README, and related tests.
- No new third-party dependencies are expected; archive reading can reuse Go standard library `archive/tar` and `compress/gzip`.
- Restore behavior affects filesystem and Git worktree state, so tests should cover archive selection, path safety, existing feature conflicts, and `.git` metadata exclusion.
