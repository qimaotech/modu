## ADDED Requirements

### Requirement: TUI confirms unpushed branch deletion
The TUI SHALL require an additional confirmation before deleting a feature when the delete operation would remove local branches that are not fully pushed.

#### Scenario: Delete has no unpushed branches
- **WHEN** the user confirms feature deletion and all branch candidates are fully pushed
- **THEN** the TUI deletes the feature and returns to the refreshed list view

#### Scenario: Delete has unpushed branches
- **WHEN** the user confirms feature deletion and at least one branch candidate is not fully pushed
- **THEN** the TUI displays a second confirmation view listing the affected project or module branches

#### Scenario: User confirms unpushed deletion
- **WHEN** the unpushed-branch confirmation view is visible and the user presses `y` or Enter
- **THEN** the TUI deletes the feature with explicit unpushed-branch permission and refreshes the list

#### Scenario: User cancels unpushed deletion
- **WHEN** the unpushed-branch confirmation view is visible and the user presses `n` or Esc
- **THEN** the TUI cancels deletion and returns to the list without removing worktrees or branches
