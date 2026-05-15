## ADDED Requirements

### Requirement: Delete worktree creates backup first
`Engine.DeleteWorktree` SHALL create a `.tar.gz` backup of the target feature directory after feature existence and dirty checks pass, but before removing any module worktree, main project worktree, branch, or feature directory. The backup SHALL be stored under `<worktree-root>/.modu/backups/`. If backup creation fails or the context is canceled before backup completion, delete SHALL stop and return an error without removing worktrees. On success, `DeleteWorktree` SHALL return a result containing the feature name and backup archive path.

#### Scenario: Successful delete creates backup before removal
- **WHEN** `DeleteWorktree` is called for an existing clean feature
- **THEN** the engine creates a `.tar.gz` archive under `<worktree-root>/.modu/backups/` before invoking worktree removal

#### Scenario: Backup failure blocks delete
- **WHEN** backup creation fails during `DeleteWorktree`
- **THEN** the engine returns an error and MUST NOT call worktree removal or remove the feature directory

#### Scenario: Backup cancellation blocks delete
- **WHEN** the context is canceled during backup creation
- **THEN** the engine returns an error and MUST NOT call worktree removal or remove the feature directory

#### Scenario: Dirty worktree blocks before backup
- **WHEN** `DeleteWorktree` is called without force and dirty checking detects uncommitted changes
- **THEN** the engine returns `ERR_DIRTY_WORKTREE` and MUST NOT create a backup archive

#### Scenario: Delete result includes backup path
- **WHEN** `DeleteWorktree` succeeds for feature `my-feature`
- **THEN** the returned result includes `Feature: my-feature` and the created `BackupPath`

### Requirement: Delete backup archive format
Delete backups SHALL use `.tar.gz` format and include the deleted feature directory as the top-level archive entry. Archive names SHALL include a local timestamp in `YYYYMMDD-HHMMSS` format and the feature directory slug produced by the existing feature directory naming rule. If the generated final path already exists, the engine SHALL choose a unique suffixed final path. Archives SHALL be written to a temporary file and atomically published to the final path only after successful compression without overwriting an existing archive.

#### Scenario: Archive contains feature root
- **WHEN** a feature directory named `feature-remove-backup` is backed up
- **THEN** the archive contains paths rooted at `feature-remove-backup/`

#### Scenario: Archive name includes timestamp and feature slug
- **WHEN** a backup is created for feature directory `feature-remove-backup`
- **THEN** the backup file name matches `YYYYMMDD-HHMMSS_feature-remove-backup.tar.gz`

#### Scenario: Slash feature uses directory slug
- **WHEN** a backup is created for feature `feature/remove-backup`
- **THEN** the backup file name uses the slug `feature-remove-backup`

#### Scenario: Existing archive name gets unique suffix
- **WHEN** the generated backup path already exists
- **THEN** the engine writes to a unique final path instead of overwriting the existing archive

#### Scenario: Failed archive removes temporary file
- **WHEN** archive creation fails before final publication
- **THEN** the engine removes the temporary archive file and leaves no final archive at the generated path

### Requirement: Expired delete backup cleanup
The engine SHALL provide a cleanup operation that removes generated backup archives in `<worktree-root>/.modu/backups/` whose modification time is older than `delete-backup.retention-days`. Generated backup archives are files whose names match `YYYYMMDD-HHMMSS_<feature-slug>.tar.gz` or its unique suffixed form. Cleanup SHALL retain files that are not expired, retain non-matching files, and ignore missing backup directories.

#### Scenario: Remove expired backup
- **WHEN** cleanup runs with `retention-days: 30` and a backup archive is older than 30 days
- **THEN** the engine removes that archive

#### Scenario: Keep recent backup
- **WHEN** cleanup runs with `retention-days: 30` and a backup archive is not older than 30 days
- **THEN** the engine keeps that archive

#### Scenario: Keep non-generated tarball
- **WHEN** cleanup runs and `.modu/backups` contains an expired `.tar.gz` file whose name does not match the generated backup naming pattern
- **THEN** the engine keeps that archive

#### Scenario: Missing backup directory
- **WHEN** cleanup runs and `<worktree-root>/.modu/backups/` does not exist
- **THEN** cleanup succeeds without creating or deleting anything
