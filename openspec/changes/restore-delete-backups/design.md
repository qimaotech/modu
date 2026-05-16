## Context

`modu delete` now creates a `.tar.gz` archive under `<worktree-root>/.modu/backups/` before removing a feature worktree. The archive preserves file content, but it does not guarantee a valid Git worktree after raw extraction because the archived `.git` entries can point at Git worktree metadata that was removed or pruned during delete.

Recovery therefore needs to be a product flow, not just documentation that tells users to extract a tarball. The engine should own backup discovery and restore semantics so CLI and TUI behave consistently.

## Goals / Non-Goals

**Goals:**

- Let users list generated delete backups with enough metadata to choose the right archive.
- Let users restore a backup manually from CLI or TUI.
- Recreate Git worktrees through existing creation logic before overlaying archived file content.
- Avoid restoring stale `.git` metadata from archives.
- Reject unsafe archive paths and avoid overwriting an existing feature directory.

**Non-Goals:**

- Restoring arbitrary user-created tarballs outside the generated delete-backup naming scheme.
- Merging a backup into an existing feature worktree.
- Recovering local branches that were intentionally deleted and cannot be recreated from the configured base branch.
- Adding a background restore service, retention scheduler, or configurable backup directory.

## Decisions

1. Add backup discovery to the engine layer.

   The engine already knows `worktree-root`, backup naming rules, and retention behavior. A new listing method should read `<worktree-root>/.modu/backups/`, accept only generated backup names, derive creation timestamps from filenames, prefer archived workspace metadata for the original feature name, and return newest-first metadata for CLI/TUI/output.

   Rejected alternative: let CLI/TUI scan backup files independently. That duplicates filename parsing and path filtering.

2. Restore by creating a fresh worktree before overlaying files.

   Restore should call existing worktree creation semantics for the destination feature, then extract archive content over the created worktree. When no destination feature is specified, restore should prefer the archived workspace metadata feature name (for example `feature/foo`) and fall back to the archive root slug. Restore should also create only configured modules that are present in the backup archive, rather than every module currently configured. This keeps Git worktree registration valid and lets existing branch/base behavior remain centralized without changing the restored worktree shape.

   Rejected alternative: raw extraction directly to `<worktree-root>/<feature>`. That can leave invalid `.git` pointers and bypass branch/worktree setup.

3. Exclude `.git` metadata during restore.

   Archive entries named `.git` or under any `.git/` path should be skipped. The recreated worktree supplies current Git metadata, while file content and untracked files from the backup are restored.

   Rejected alternative: restore everything and hope Git can recover. That is fragile and can corrupt the new worktree metadata.

4. Resolve backups by ID or path for CLI, by selected item for TUI.

   CLI should accept either a generated backup ID from `modu backup list` or an explicit archive path. TUI can present the same engine metadata as a picker. If the user does not provide a destination feature, restore should default to the original feature stored in the archived workspace metadata, falling back to the feature slug encoded in the backup filename.

   Rejected alternative: require users to copy full paths every time. IDs make common usage easier without losing path-based precision.

5. Do not run automatic cleanup before backup-management commands.

   A user may run list/restore specifically because a backup is old. Running retention cleanup before these commands could delete the candidate before it is displayed or restored. Existing cleanup remains for normal commands.

   Rejected alternative: keep unconditional cleanup on all config-loading commands. That preserves the previous pattern but surprises users during recovery.

## Risks / Trade-offs

- [Risk] Restore can fail after creating the new worktree but before all files are extracted. -> Mitigation: fail with a clear error and leave the partially created worktree for inspection instead of deleting recovered content.
- [Risk] Very old backups might not contain workspace metadata with the original branch name. -> Mitigation: fall back to the feature directory slug and allow `--feature` in CLI.
- [Risk] Large backups can make TUI restore appear slow. -> Mitigation: run restore as a Bubble Tea command and return to loading/error states through existing patterns.
- [Risk] Extracted paths can attempt traversal. -> Mitigation: normalize each tar entry, require it to stay under the destination feature directory, and reject unsafe entries.

## Migration Plan

No data migration is required. Existing generated delete backups remain valid. Rollback is to remove the new list/restore commands and TUI entry point; existing backup creation and cleanup remain unchanged.

## Open Questions

None.
