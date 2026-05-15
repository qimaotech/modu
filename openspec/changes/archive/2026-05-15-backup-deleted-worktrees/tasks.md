## 1. Config

- [x] 1.1 Add `DeleteBackupConfig` with `retention-days` YAML support and wire it into `config.Config`.
- [x] 1.2 Default missing or zero `delete-backup.retention-days` to 30 in both normal and scan config loading.
- [x] 1.3 Validate negative `delete-backup.retention-days` as `ERR_CONFIG_INVALID`.
- [x] 1.4 Add config unit tests for default, custom, and negative retention values.

## 2. Engine Backup And Cleanup

- [x] 2.1 Add a `DeleteResult` containing at least `Feature` and `BackupPath`.
- [x] 2.2 Implement `.tar.gz` delete backup creation under `<worktree-root>/.modu/backups/` with timestamped feature-slug filenames and unique suffix handling.
- [x] 2.3 Ensure the archive contains the feature directory as its top-level entry and handles regular files, directories, and symlinks.
- [x] 2.4 Integrate backup creation into `DeleteWorktree` after dirty checks and before any worktree removal.
- [x] 2.5 Make backup creation failure stop deletion before module/main worktree removal.
- [x] 2.6 Write backups through a temporary file, atomically publish on success without overwriting existing archives, remove temp files on failure, and honor context cancellation before deletion.
- [x] 2.7 Implement expired backup cleanup based on `delete-backup.retention-days`, archive modification time, and generated filename matching.
- [x] 2.8 Add engine tests for successful backup-before-delete, returned backup path, backup failure blocking delete, cancellation blocking delete, dirty check before backup, archive shape, unique naming, temp cleanup, and retention cleanup.

## 3. CLI, Output, And TUI

- [x] 3.1 Update delete output formatting so text output shows `备份文件: <path>` and JSON output includes `backupPath`.
- [x] 3.2 Update `modu delete` to use the new `DeleteResult` and pass the backup path to output formatting.
- [x] 3.3 Run cleanup once after successful existing config load for CLI commands that use `LoadConfig` or `LoadConfigForScan`, warning to stderr without blocking when cleanup fails.
- [x] 3.4 Run cleanup once in `ui.StartTUI` after config loading and before starting the Bubble Tea program, warning/message without blocking on cleanup failure.
- [x] 3.5 Update TUI delete confirmation handling to display the backup path after a successful delete and preserve error-state behavior on backup failure.
- [x] 3.6 Add output, CLI, and TUI tests for backup path rendering and startup cleanup behavior.

## 4. Documentation And Verification

- [x] 4.1 Update README configuration examples and delete command documentation for `delete-backup.retention-days` and backup path behavior.
- [x] 4.2 Run `gofmt` on changed Go files.
- [x] 4.3 Run focused unit tests for config, engine, output, and UI packages.
- [x] 4.4 Run `go test ./...`.
- [x] 4.5 Run OpenSpec status/validation for `backup-deleted-worktrees`.
