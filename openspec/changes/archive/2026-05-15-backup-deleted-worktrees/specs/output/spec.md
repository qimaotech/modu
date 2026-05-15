## ADDED Requirements

### Requirement: Delete response includes backup path
Delete responses SHALL include the backup archive path produced by the engine when deletion succeeds. Text output SHALL display the backup path in Chinese user-facing copy, and JSON output SHALL include a `backupPath` field.

#### Scenario: Text delete response
- **WHEN** formatting a successful delete response with backup path `/worktrees/.modu/backups/20260515-153012_my-feature.tar.gz`
- **THEN** text output includes `备份文件: /worktrees/.modu/backups/20260515-153012_my-feature.tar.gz`

#### Scenario: JSON delete response
- **WHEN** formatting a successful delete response with backup path `/worktrees/.modu/backups/20260515-153012_my-feature.tar.gz`
- **THEN** JSON output includes `"backupPath": "/worktrees/.modu/backups/20260515-153012_my-feature.tar.gz"`
