## 1. Engine

- [x] 1.1 Add delete backup metadata types and generated backup filename parsing.
- [x] 1.2 Implement engine delete backup listing with newest-first sorting and invalid-file filtering.
- [x] 1.3 Implement backup resolution by ID or generated archive path.
- [x] 1.4 Implement restore flow that creates a fresh worktree, extracts archive content, skips `.git` metadata, and validates extraction paths.
- [x] 1.5 Add engine tests for listing, sorting, invalid-name filtering, missing backup directory, restore success, existing destination failure, `.git` skipping, and unsafe archive rejection.

## 2. CLI And Output

- [x] 2.1 Add output formatting for backup list and backup restore in text and JSON modes.
- [x] 2.2 Add `modu backup list` and `modu backup restore <backup>` commands with `--feature` and `--base` flags.
- [x] 2.3 Ensure backup management commands load config without automatic retention cleanup.
- [x] 2.4 Add CLI/output tests for backup list/restore rendering and cleanup avoidance.
- [x] 2.5 Show main project dirty state in list/info text output.

## 3. TUI

- [x] 3.1 Add TUI backup selection state and restore completion message handling.
- [x] 3.2 Add list-view `r` shortcut and operation-menu "恢复备份" entry.
- [x] 3.3 Add TUI tests for entering the backup picker, empty backup list behavior, and restoring a selected backup.
- [x] 3.4 Add TUI force-delete confirmation shortcut for restored dirty worktree cleanup.
- [x] 3.5 Show main project dirty state in TUI list summaries.

## 4. Documentation And Verification

- [x] 4.1 Update README with backup list and restore command usage.
- [x] 4.2 Run OpenSpec validation/status for `restore-delete-backups`.
- [x] 4.3 Run Go formatting and targeted tests.
