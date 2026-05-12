## ADDED Requirements

### Requirement: TUI feature creation entry
The TUI SHALL provide a list-view action for starting feature creation without changing existing list navigation, open, update, copy, module-management, delete, or quit shortcuts.

#### Scenario: Start creation from list view
- **WHEN** the user is in the list view and presses `n`
- **THEN** the TUI enters the feature-name input state for creating a new feature

#### Scenario: Existing quit shortcut remains in list view
- **WHEN** the user is in the list view and presses `q`
- **THEN** the TUI exits as before

#### Scenario: Create shortcut is shown
- **WHEN** the TUI renders the list view
- **THEN** the shortcut help includes the Chinese create-feature action

### Requirement: Editable feature-name input
The TUI SHALL let users enter and edit the feature name with cursor movement before selecting projects/modules.

#### Scenario: Printable character insertion
- **WHEN** the user types printable characters in the feature-name input state
- **THEN** the characters are inserted at the current cursor position and the cursor moves after the inserted text

#### Scenario: Cursor movement
- **WHEN** the user presses left, right, home, or end in the feature-name input state
- **THEN** the input cursor moves within the feature-name text without leaving the input state

#### Scenario: In-place deletion
- **WHEN** the user presses backspace or delete in the feature-name input state
- **THEN** the TUI removes text adjacent to the cursor without leaving the input state

#### Scenario: q is normal input text
- **WHEN** the user presses `q` in the feature-name input state
- **THEN** the TUI inserts `q` into the feature name and MUST NOT quit or cancel input

#### Scenario: Empty name is rejected
- **WHEN** the user presses Enter while the feature-name input is empty or whitespace-only
- **THEN** the TUI remains in the feature-name input state and shows a Chinese validation message

#### Scenario: Cancel feature-name input
- **WHEN** the user presses Esc or Ctrl+C in the feature-name input state
- **THEN** the TUI cancels creation and returns to the list view without creating a worktree

### Requirement: Project selection before creation
The TUI SHALL provide a project/module selection step after the user confirms a feature name and before any worktree creation starts.

#### Scenario: Open selection after valid name
- **WHEN** the user enters a non-empty feature name and presses Enter
- **THEN** the TUI displays a selection view for configured projects/modules before creating the feature

#### Scenario: Main project is included
- **WHEN** the user reaches the project/module selection step
- **THEN** the main project is treated as always included for feature creation

#### Scenario: Configured modules are selectable
- **WHEN** the project/module selection step is displayed
- **THEN** each configured module is shown as selectable and can be toggled before confirmation

#### Scenario: Default and remote branch preselection
- **WHEN** configured default-selected modules or modules with the target remote branch exist
- **THEN** those modules are preselected in the project/module selection step

#### Scenario: Selection cancellation
- **WHEN** the user cancels the project/module selection step
- **THEN** the TUI returns to the list view without creating a worktree

### Requirement: Create feature from TUI selection
The TUI SHALL create the feature worktree after the user confirms project/module selection.

#### Scenario: Create selected projects
- **WHEN** the user confirms project/module selection for a feature name
- **THEN** the TUI creates the main project worktree and only the selected module worktrees

#### Scenario: Create with no selected modules
- **WHEN** the user confirms creation with no modules selected
- **THEN** the TUI creates the main project feature worktree without module worktrees

#### Scenario: Creation success feedback
- **WHEN** feature creation succeeds
- **THEN** the TUI reloads the list view and shows a Chinese success message containing the feature name

#### Scenario: Creation failure feedback
- **WHEN** feature creation fails
- **THEN** the TUI enters the error state and shows the creation error
