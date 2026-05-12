## 1. TUI State And Entry

- [x] 1.1 Add App fields for feature creation input text, cursor position, validation message, and create-specific module/project selection state.
- [x] 1.2 Add a list-view `n` shortcut that enters the feature-name input state without changing existing list shortcuts.
- [x] 1.3 Update list-view help text to include the Chinese create-feature action.

## 2. Feature Name Input

- [x] 2.1 Render a Chinese feature-name input view with visible current text, cursor position, validation feedback, and cancel/confirm hints.
- [x] 2.2 Handle printable rune insertion at the cursor, including `q` as ordinary input text.
- [x] 2.3 Handle left, right, home, end, backspace, and delete without leaving the input state.
- [x] 2.4 Keep Enter on empty or whitespace-only input in the input state with a Chinese validation message.
- [x] 2.5 Handle Esc and Ctrl+C as cancellation that returns to the list view without creating anything.

## 3. Project Selection And Creation

- [x] 3.1 After a valid feature name, query remote branch presence for configured modules and initialize selection with default-selected and remote-branch modules preselected.
- [x] 3.2 Render the project/module selection step with the main project shown as always included and configured modules toggleable.
- [x] 3.3 Handle selection cancellation by returning to the list view without calling the engine.
- [x] 3.4 On selection confirmation, call `CreateWorktree` with the selected modules and the configured default base; allow zero selected modules to create only the main project worktree.
- [x] 3.5 On creation success, reload the list view and show a Chinese success message containing the feature name.
- [x] 3.6 On creation failure, enter the existing error state with the creation error.

## 4. Tests And Documentation

- [x] 4.1 Add unit tests for list `n` entry and list `q` quit behavior.
- [x] 4.2 Add unit tests for feature-name input insertion, cursor movement, deletion, empty-name validation, `q` as input, and cancellation.
- [x] 4.3 Add unit tests for project/module selection initialization and cancellation behavior.
- [x] 4.4 Add unit tests or a focused fake-engine seam for selected-module creation behavior, including zero-module creation.
- [x] 4.5 Update README TUI shortcut documentation for the create action.
- [x] 4.6 Run `gofmt` on touched Go files and `go test ./...`.
