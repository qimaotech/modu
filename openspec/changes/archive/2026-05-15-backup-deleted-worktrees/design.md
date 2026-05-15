## Context

`modu delete` currently performs dirty checks, removes configured module worktrees, removes the main project worktree, and finally deletes the feature directory. Once this completes there is no local recovery path for accidentally deleted worktree contents.

The requested behavior is cross-cutting: configuration gains a retention setting, engine deletion gains a backup step, CLI/TUI startup gains cleanup, and output needs to expose where the backup was written.

## Goals / Non-Goals

**Goals:**

- Create a `.tar.gz` backup of the feature directory before any delete operation removes worktree content.
- Store backups at the fixed path `<worktree-root>/.modu/backups/`.
- Make retention configurable with `delete-backup.retention-days`, defaulting to 30 days.
- Run expired-backup cleanup once after each successful configuration load for normal CLI/TUI tasks.
- Show users the backup path after a successful delete.

**Non-Goals:**

- Add a restore command.
- Add CLI/TUI restore flows, backup browsers, restore menu entries, or manual backup selection for restore.
- Run a background daemon or OS-level scheduler.
- Make backup directory or archive format configurable.
- Back up individual module removals from `create`/module-management flows.
- Guarantee that extracting an archive recreates a registered Git worktree; backups are for file-content recovery.

## Decisions

1. Keep backup and cleanup in the engine layer.

   `Engine.DeleteWorktree` has the full context needed to resolve `worktree-root`, apply dirty-check semantics, and coordinate the delete sequence. The engine should own backup creation and return a `DeleteResult` containing `Feature` and `BackupPath` so CLI/TUI/output can render it.

   Rejected alternative: implement backup in CLI/TUI. That would duplicate behavior and risk one entry point deleting without backup.

2. Create the backup after dirty checks but before any worktree removal.

   Dirty worktrees already block non-forced deletes. Once those checks pass, the archive must be created before module or main project worktrees are removed. If archive creation fails, deletion stops and no worktree removal is attempted.

   Rejected alternative: back up after `git worktree remove`. At that point the source content can already be gone.

3. Use a fixed hidden backup directory and timestamped archive names.

   Backups live under `<worktree-root>/.modu/backups/`, outside feature directories, so the archive never includes prior backups. File names should use local time in `YYYYMMDD-HHMMSS` format plus the existing `featureToDirName` slug, for example `20260515-153012_feature-remove-backup.tar.gz`. If that name already exists, append a numeric suffix before `.tar.gz`.

   Rejected alternative: put backups beside the deleted feature. That clutters the worktree root and makes cleanup harder to scope safely.

4. Use Go standard library tar/gzip.

   Implement archive writing with `archive/tar`, `compress/gzip`, and `filepath.WalkDir`. The archive should preserve a top-level feature directory entry and include regular files, directories, and symlinks without following symlinks. The archive is a content backup; `.git` pointer files may be archived as ordinary files, but the archive does not promise a directly reusable Git worktree after extraction.

   Rejected alternative: shell out to `tar`. Standard library code is easier to test and avoids platform-specific command behavior.

5. Treat "timed cleanup" as startup cleanup for this CLI app.

   `modu` is not a long-running service. Cleanup should run once after an existing config is loaded by CLI commands that use `LoadConfig` or `LoadConfigForScan`, and once in `ui.StartTUI` after `LoadConfig` succeeds and before entering the Bubble Tea program. Cleanup failures should warn/log and allow the requested command to continue; CLI warnings must go to stderr so JSON stdout remains parseable.

6. Write archives through a temp file and publish them atomically.

   Archive creation should write to a temporary file in the backup directory, then publish it to the final generated name only after compression succeeds. If the generated final name already exists, choose a suffixed name without overwriting the existing archive. If compression fails or the context is canceled, remove the temp file and return an error before any worktree removal begins.

   Rejected alternative: spawn a background goroutine or daemon. Short-lived CLI commands can exit before timers fire, and a daemon would add operational complexity outside the current product shape.

## Risks / Trade-offs

- [Risk] Large worktrees can make deletion slower. → Mitigation: make the backup step explicit in output/logs and keep compression local with no network dependency.
- [Risk] Backup creation can fail because of permissions, disk space, or unsupported file types. → Mitigation: fail closed; do not delete if backup cannot be written.
- [Risk] Cleanup could delete user-created files in the backup directory. → Mitigation: only target generated backup names under the fixed `.modu/backups` directory.
- [Risk] Retention based on file modification time can be affected if files are touched manually. → Mitigation: generated filenames still carry creation time for humans; tests can validate cleanup with controlled mtimes.

## Migration Plan

No data migration is required. Existing configs continue to load; missing `delete-backup` config defaults to 30 days. Rollback is to remove the new config field and backup calls, leaving existing `.modu/backups` archives untouched.

## Open Questions

None.
