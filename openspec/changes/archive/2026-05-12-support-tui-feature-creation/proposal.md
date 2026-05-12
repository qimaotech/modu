## Why

Users can create features from the CLI, but the TUI currently focuses on listing and managing existing worktrees. Adding feature creation to the TUI closes that workflow gap and lets users stay in the interactive interface for the common create-and-select-modules path.

## What Changes

- Add a TUI entry point for creating a new feature from the list view.
- Add a TUI project/module selection step for the feature being created, reusing configured modules and existing selection semantics.
- Add a feature-name input state that supports cursor movement and in-place edits.
- Ensure `q` does not exit while the user is typing the feature name; quit/cancel behavior remains available outside the input state.
- Refresh the list and show a Chinese success or error message after creation.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `tui`: TUI feature creation, project/module selection, editable feature-name input, and input-state key handling requirements change.

## Impact

- `internal/ui`: Bubble Tea state machine, rendering, key handling, module/project selection flow, and unit tests.
- `internal/engine`: Existing `CreateWorktree`, remote-branch preselection, and configured module filtering may be reused by the TUI flow.
- `README.md`: TUI shortcut documentation may need to mention the create action.
