## ADDED Requirements

### Requirement: Delete backup retention configuration
The config SHALL support an optional `delete-backup.retention-days` value that controls how many days delete backup archives are retained. When `delete-backup` or `retention-days` is omitted, or `retention-days` is zero, config loading SHALL default the value to 30 days.

#### Scenario: Load config without delete backup settings
- **WHEN** `.modu.yaml` omits `delete-backup`
- **THEN** config loading succeeds with delete backup retention set to 30 days

#### Scenario: Load config with custom retention
- **WHEN** `.modu.yaml` contains `delete-backup.retention-days: 7`
- **THEN** config loading succeeds with delete backup retention set to 7 days

#### Scenario: Reject negative retention
- **WHEN** `.modu.yaml` contains `delete-backup.retention-days: -1`
- **THEN** config loading fails with `ERR_CONFIG_INVALID`
