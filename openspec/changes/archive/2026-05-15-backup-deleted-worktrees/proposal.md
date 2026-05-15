## Why

Deleting a feature worktree is currently irreversible once the worktree directories are removed. A compressed backup gives users a recovery path while keeping old backups bounded by a configurable retention period.

## What Changes

- Create a `.tar.gz` archive of the feature worktree before deletion begins.
- Store delete backups under `<worktree-root>/.modu/backups/`.
- Add configurable backup retention with `delete-backup.retention-days`, defaulting to 30 days.
- Clean expired delete backups once after each `modu` command/TUI startup that successfully loads existing configuration.
- Surface backup success/failure clearly; if backup creation fails, deletion MUST stop.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `config`: Adds delete-backup retention configuration and defaults.
- `engine`: Changes delete semantics to backup before removing worktrees and to clean expired backup archives.
- `cli`: Runs backup cleanup after configuration load and reports delete backup results in command output.
- `output`: Adds backup path fields/text to delete responses.
- `tui`: Runs backup cleanup at startup and shows delete success feedback that includes backup creation.

## Impact

- Affected code: `internal/config`, `internal/engine`, `internal/output`, `cmd/modu`, `internal/ui`, and tests.
- Uses Go standard library `archive/tar` and `compress/gzip`; no new third-party dependency is expected.
- Existing delete commands remain compatible; backup behavior is additive and enabled by default.
- Backup archives are for file-content recovery and do not guarantee that extracting the archive restores a valid Git worktree registration.
- This change only reports the backup path; product restore entry points and manual restore selection belong to a later OpenSpec change.
