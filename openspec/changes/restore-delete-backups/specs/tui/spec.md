## ADDED Requirements

### Requirement: TUI exposes manual backup restore
The TUI SHALL provide a manual restore entry point from the list view and operation menu so users can restore backups even when the deleted feature is not present in the feature list.

#### Scenario: Start restore from list view
- **WHEN** the user is in the list view and presses `r`
- **THEN** the TUI enters a delete backup selection view

#### Scenario: Start restore from operation menu
- **WHEN** the user is in the operation menu and selects "恢复备份"
- **THEN** the TUI enters a delete backup selection view

#### Scenario: Restore shortcut is shown
- **WHEN** the TUI renders the list view or operation menu
- **THEN** the visible help or menu includes the restore action

### Requirement: TUI backup selection restores selected backup
The TUI SHALL list generated delete backups in a selection view and restore the selected backup after confirmation.

#### Scenario: Backup picker displays backups
- **WHEN** generated delete backups exist
- **THEN** the TUI backup selection view shows each backup's feature slug and creation time

#### Scenario: Confirm selected backup restore
- **WHEN** the user selects a backup and confirms restore
- **THEN** the TUI restores that backup and returns to the refreshed list view with a success message

#### Scenario: Empty backup list
- **WHEN** no generated delete backups exist
- **THEN** the TUI shows an empty-state message and lets the user return to the list view

### Requirement: TUI delete confirmation supports force delete
The TUI delete confirmation SHALL allow users to force delete a feature worktree from the confirmation view while preserving the existing normal delete behavior.

#### Scenario: Force delete from confirmation
- **WHEN** the user is in delete confirmation and presses `f`
- **THEN** the TUI calls delete with force enabled and refreshes the list after success

#### Scenario: Force delete shortcut is shown
- **WHEN** the TUI renders delete confirmation
- **THEN** the confirmation help includes the force delete shortcut

### Requirement: TUI list includes main project dirty state
The TUI feature list SHALL count a dirty main project worktree in the visible dirty summary for a feature.

#### Scenario: Feature main project dirty
- **WHEN** a feature's main project worktree is dirty and its modules are clean
- **THEN** the TUI list shows that feature as dirty
