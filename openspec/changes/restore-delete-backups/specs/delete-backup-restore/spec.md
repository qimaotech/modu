## ADDED Requirements

### Requirement: Delete backups are discoverable
The system SHALL allow generated delete backup archives to be listed with stable metadata that identifies the backup, source feature, archive path, creation time, size, and modification time. Listing SHALL include only generated delete backup archives and SHALL ignore missing backup directories. When archived workspace metadata contains an original feature name, listing SHALL use it instead of only the directory slug.

#### Scenario: List generated backups
- **WHEN** generated delete backup archives exist under `<worktree-root>/.modu/backups/`
- **THEN** the system lists those backups newest first with ID, feature, path, creation time, size, and modification time

#### Scenario: List backup with original feature metadata
- **WHEN** a generated delete backup contains workspace metadata for feature `feature/restore`
- **THEN** the backup list shows feature `feature/restore` even though the archive root directory is `feature-restore`

#### Scenario: Ignore non-generated files
- **WHEN** the backup directory contains files that do not match the generated delete backup naming rule
- **THEN** the system omits those files from the backup list

#### Scenario: Missing backup directory is empty list
- **WHEN** the backup directory does not exist
- **THEN** the system returns an empty backup list without error

### Requirement: Delete backups can be restored manually
The system SHALL allow a user to restore a generated delete backup manually. Restore SHALL create a fresh feature worktree for the destination feature, then overlay archived file content into that feature directory. Restore SHALL create only configured module worktrees that are present in the backup archive. If the destination feature already exists, restore MUST fail without overwriting it.

#### Scenario: Restore backup to default feature
- **WHEN** a user restores a backup without specifying a destination feature
- **THEN** the system restores it to the original feature from archived workspace metadata, falling back to the feature slug derived from the backup filename

#### Scenario: Restore preserves archived module shape
- **WHEN** a backup archive contains `module1` but not `module2`
- **THEN** restore creates the main worktree and `module1` worktree but does not create `module2`

#### Scenario: Restore backup to explicit feature
- **WHEN** a user restores a backup and specifies a destination feature
- **THEN** the system restores the archive content into a freshly created worktree for that destination feature

#### Scenario: Existing feature blocks restore
- **WHEN** the destination feature directory already exists
- **THEN** restore fails without overwriting the existing directory

### Requirement: Restore protects worktree integrity
Restore SHALL NOT restore archived Git metadata entries named `.git` or below `.git/`. Restore MUST reject archive entries whose normalized path would escape the destination feature directory.

#### Scenario: Git metadata is skipped
- **WHEN** a backup archive contains `.git` metadata entries
- **THEN** restore skips those entries and keeps the freshly created worktree Git metadata

#### Scenario: Unsafe archive path is rejected
- **WHEN** a backup archive contains a path traversal entry
- **THEN** restore fails and does not write that entry outside the destination feature directory
