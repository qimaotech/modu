## ADDED Requirements

### Requirement: Engine lists delete backups
The engine SHALL provide a delete backup listing operation that scans `<worktree-root>/.modu/backups/`, returns only generated delete backup archives, and derives each backup's ID from the archive filename. The engine SHALL derive the source feature from archived workspace metadata when available, otherwise from the archive root slug.

#### Scenario: Engine backup list is sorted
- **WHEN** multiple generated backup archives exist
- **THEN** the engine returns them sorted newest first

#### Scenario: Engine backup list skips invalid names
- **WHEN** generated and non-generated files coexist in the backup directory
- **THEN** the engine returns only generated backup archives

#### Scenario: Engine backup list prefers workspace feature
- **WHEN** a generated backup contains workspace metadata for feature `feature/restore`
- **THEN** the engine reports the backup feature as `feature/restore`

### Requirement: Engine restores delete backups
The engine SHALL provide a restore operation that resolves a generated backup by ID or path, creates a fresh destination feature worktree through existing worktree creation behavior, and extracts archive content into the destination while skipping archived `.git` metadata. When no destination feature is provided, the engine SHALL prefer the archived workspace metadata feature and fall back to the archive root slug. The engine SHALL create only configured module worktrees present in the backup archive.

#### Scenario: Engine restore creates worktree first
- **WHEN** restore is requested for a valid generated backup
- **THEN** the engine creates the destination feature worktree before writing archived content

#### Scenario: Engine restore preserves original feature name
- **WHEN** restore is requested for a backup whose workspace metadata stores feature `feature/restore`
- **THEN** the engine creates worktrees for branch `feature/restore` under directory `feature-restore`

#### Scenario: Engine restore filters modules by archive
- **WHEN** restore is requested for a backup containing only configured module `module1`
- **THEN** the engine creates a module worktree for `module1` and does not create worktrees for other configured modules

#### Scenario: Engine restore rejects existing destination
- **WHEN** the destination feature directory already exists
- **THEN** the engine returns an error and does not overwrite it

#### Scenario: Engine restore rejects unsafe extraction
- **WHEN** the archive contains a path that escapes the destination feature directory
- **THEN** the engine returns an error and does not write the unsafe entry
