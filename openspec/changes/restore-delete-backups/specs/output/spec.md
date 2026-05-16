## ADDED Requirements

### Requirement: Backup list response
Output formatting SHALL support delete backup list responses in text and JSON formats.

#### Scenario: Text backup list response
- **WHEN** formatting a backup list response in text mode
- **THEN** the output includes each backup's ID, feature slug, creation time, size, and path

#### Scenario: JSON backup list response
- **WHEN** formatting a backup list response in JSON mode
- **THEN** the output includes a `backups` array with ID, feature, path, createdAt, sizeBytes, and modTime fields

### Requirement: Backup restore response
Output formatting SHALL support delete backup restore responses in text and JSON formats.

#### Scenario: Text backup restore response
- **WHEN** formatting a successful restore response in text mode
- **THEN** the output includes the restored feature, destination path, and backup path

#### Scenario: JSON backup restore response
- **WHEN** formatting a successful restore response in JSON mode
- **THEN** the output includes restored feature, destination path, and backup path fields

### Requirement: Worktree list and info show main project dirty state
Text output for worktree list and info SHALL show the main project worktree dirty state when status information is requested or details are displayed.

#### Scenario: List shows main project dirty
- **WHEN** formatting a text list response with status enabled and a feature's main project is dirty
- **THEN** the feature row includes `main: dirty`

#### Scenario: Info shows main project dirty
- **WHEN** formatting a text info response for a feature whose main project is dirty
- **THEN** the output includes the main project section with dirty status
