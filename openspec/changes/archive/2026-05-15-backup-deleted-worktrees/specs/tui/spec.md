## ADDED Requirements

### Requirement: TUI runs delete backup cleanup on startup
The TUI SHALL invoke delete backup cleanup once after configuration is loaded and before entering the interactive list. Cleanup failure SHALL be surfaced as a non-blocking warning or message and MUST NOT prevent the TUI from starting.

#### Scenario: TUI startup cleanup
- **WHEN** `modu tui` successfully loads configuration
- **THEN** delete backup cleanup runs once before the TUI list is displayed

#### Scenario: TUI cleanup failure is non-blocking
- **WHEN** delete backup cleanup fails during TUI startup
- **THEN** the TUI still starts and the cleanup failure is surfaced as a warning or message

### Requirement: TUI delete success shows backup path
After a confirmed delete succeeds, the TUI SHALL show Chinese success feedback that includes the feature name and the created backup archive path.

#### Scenario: Confirmed delete succeeds
- **WHEN** the user confirms deletion of feature `my-feature` and backup creation succeeds
- **THEN** the TUI reloads the list and shows a success message containing `已删除 feature: my-feature` and the backup archive path

#### Scenario: Delete backup fails
- **WHEN** the user confirms deletion and backup creation fails
- **THEN** the TUI enters the error state and MUST NOT remove the feature worktree
