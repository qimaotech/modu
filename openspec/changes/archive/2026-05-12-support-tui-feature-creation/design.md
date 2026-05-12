## Context

The current TUI (`internal/ui`) is a Bubble Tea state machine with list, menu, modules, confirm, loading, and error states. CLI feature creation already supports interactive module selection through `ui.SelectModules`, remote-branch preselection through `Engine.GetModulesWithRemoteBranch`, and worktree creation through `Engine.CreateWorktree`.

The missing path is creating a feature from inside the TUI. The TUI also needs an editable feature-name input that behaves like a text field instead of treating every `q` keypress as a quit signal.

## Goals / Non-Goals

**Goals:**

- Add a list-view action to start TUI feature creation without disrupting existing shortcuts.
- Add a dedicated feature-name input state with cursor movement and in-place editing.
- Add a project/module selection state before creation, with the main project always included and configured modules selectable.
- Reuse existing engine creation behavior, module defaults, and remote branch preselection where practical.
- Keep all user-visible TUI text in Chinese.

**Non-Goals:**

- Changing CLI `modu create` behavior.
- Adding new dependencies or replacing Bubble Tea primitives.
- Changing the worktree directory naming rules or branch creation semantics.
- Implementing multi-main-project configuration; the current workspace main project remains implicit and always included.

## Decisions

1. Use a dedicated TUI create state instead of overloading menu state.

   The list view will expose a new `n` shortcut for "新建 feature". Existing `c` is already used for copy path, so reusing it would create a shortcut conflict. A separate state keeps key handling explicit: list-level `q` still quits, while create-input `q` inserts text.

   Rejected alternative: add "创建 feature" to each selected-item operation menu. Creation is a global action rather than an operation on the selected worktree, so putting it in the per-item menu would blur menu semantics.

2. Implement feature-name editing as a small local text input model.

   Store the input as runes plus a cursor index, and handle printable runes, backspace, delete, left/right, home/end, Enter, Esc, and Ctrl+C. This avoids a dependency change and is enough for the requested cursor movement and edit behavior.

   Rejected alternative: add a new text input dependency. The project already uses Bubble Tea directly, and the required behavior is narrow.

3. Reuse `ModuleSelector` semantics for project/module selection.

   After the user confirms a feature name, the TUI should query remote branch presence for configured modules, preselect configured defaults and modules that already have the remote branch, then allow selecting modules before creation. The main project is not optional; it is created as the feature root by `Engine.CreateWorktree`.

   Rejected alternative: immediately create all modules after name input. That would not satisfy the requested project selection flow and would bypass existing default-selection behavior.

4. Execute creation through the existing engine path.

   The TUI should create a shallow copy of the loaded config with `Modules` narrowed to the selected modules, then call `CreateWorktree(ctx, featureName, DefaultBase)`. On success it reloads the list and shows a success message; on failure it enters the existing error state.

   Rejected alternative: duplicate CLI `runCreate` logic in the UI. Reusing engine methods keeps branch/worktree handling in one place.

## Risks / Trade-offs

- [Risk] Feature names containing `/` are branches but map to `-` directories, while the list currently displays directory names. → Preserve existing engine naming behavior and avoid changing list display semantics in this change.
- [Risk] Remote branch preselection may be slow or fail. → Follow CLI behavior: warn/degrade to an empty preselection map and still allow manual selection.
- [Risk] `ModuleSelector` currently treats `q` as quit. → Keep that behavior for selection screens; only the feature-name input state treats `q` as typed text.
- [Risk] Empty feature names can trigger confusing engine errors. → Validate and keep the user in the input state with a Chinese error before starting selection or creation.

## Migration Plan

No data migration is required. Existing CLI and TUI workflows remain valid. Rollback is a code revert of the TUI state additions and documentation/test updates.

## Open Questions

None.
