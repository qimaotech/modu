## ADDED Requirements

### Requirement: Base branch selection before creation
The TUI SHALL let users select or confirm the base branch after entering a feature name and before selecting projects/modules.

#### Scenario: Default base is preselected
- **WHEN** the user enters a valid feature name and presses Enter
- **THEN** the TUI displays a base-branch selection list with `.modu.yaml` `default-base` preselected

#### Scenario: Available branches are selectable
- **WHEN** the user presses Up/Down or k/j in the base-branch selection state
- **THEN** the TUI moves the highlighted base branch without leaving the base-branch selection state

#### Scenario: Branch list failure falls back to default
- **WHEN** the TUI cannot read the workspace branch list
- **THEN** the TUI still offers `.modu.yaml` `default-base` as the selectable base branch and shows a Chinese warning

#### Scenario: Base selection advances to modules
- **WHEN** the user confirms the highlighted base branch in the base-branch selection state
- **THEN** the TUI displays the project/module selection step

#### Scenario: Cancel base selection
- **WHEN** the user presses Esc or Ctrl+C in the base-branch selection state
- **THEN** the TUI cancels creation and returns to the list view without creating a worktree

## MODIFIED Requirements

### Requirement: Project selection before creation
The TUI SHALL provide a project/module selection step after the user confirms a feature name and base branch and before any worktree creation starts.

#### Scenario: Open selection after valid name and base
- **WHEN** the user enters a non-empty feature name, confirms a non-empty base branch, and presses Enter
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
The TUI SHALL create the feature worktree from the confirmed base branch after the user confirms project/module selection.

#### Scenario: Create selected projects
- **WHEN** the user confirms project/module selection for a feature name and base branch
- **THEN** the TUI creates the main project worktree from the confirmed base branch and only the selected module worktrees

#### Scenario: Create with no selected modules
- **WHEN** the user confirms creation with no modules selected
- **THEN** the TUI creates the main project feature worktree from the confirmed base branch without module worktrees

#### Scenario: Module base branch still wins
- **WHEN** the user confirms creation with a selected module that configures `base-branch`
- **THEN** that module is created from its configured `base-branch`

#### Scenario: Creation success feedback
- **WHEN** feature creation succeeds
- **THEN** the TUI reloads the list view and shows a Chinese success message containing the feature name

#### Scenario: Creation failure feedback
- **WHEN** feature creation fails
- **THEN** the TUI enters the error state and shows the creation error
