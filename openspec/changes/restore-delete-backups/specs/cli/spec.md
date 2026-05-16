## ADDED Requirements

### Requirement: CLI lists delete backups
The CLI SHALL provide `modu backup list` to display generated delete backups. The command SHALL support text and JSON output.

#### Scenario: Text backup list
- **WHEN** `modu backup list` succeeds with text output
- **THEN** stdout lists each backup with ID, feature slug, creation time, size, and path

#### Scenario: JSON backup list
- **WHEN** `modu backup list -o json` succeeds
- **THEN** stdout contains a JSON response with a `backups` array

### Requirement: CLI restores delete backups
The CLI SHALL provide `modu backup restore <backup>` where `<backup>` is either a generated backup ID or an archive path. The command SHALL support `--feature <feature>` to choose a destination feature and `--base <branch>` to choose the base branch used when recreating the worktree.

#### Scenario: Restore by backup ID
- **WHEN** `modu backup restore <backup-id>` succeeds
- **THEN** the CLI restores the selected backup and prints the destination feature and path

#### Scenario: Restore by backup path
- **WHEN** `modu backup restore /path/to/archive.tar.gz --feature recovered-feature` succeeds
- **THEN** the CLI restores the archive to `recovered-feature`

#### Scenario: Restore failure uses normal error output
- **WHEN** `modu backup restore <backup>` fails
- **THEN** the CLI prints the normal formatted error response and exits non-zero

### Requirement: Backup management commands preserve candidates
CLI backup management commands SHALL load configuration without running automatic delete backup cleanup before listing or restoring backups.

#### Scenario: Backup list does not clean first
- **WHEN** `modu backup list` runs successfully
- **THEN** automatic delete backup cleanup is not invoked before the list operation

#### Scenario: Backup restore does not clean first
- **WHEN** `modu backup restore <backup>` runs successfully
- **THEN** automatic delete backup cleanup is not invoked before the restore operation
