## 1. CLI Create Base Resolution

- [x] 1.1 Change the `create --base` flag default to empty and resolve an omitted base from loaded config `default-base`.
- [x] 1.2 Add or adjust CLI tests to cover default config base, explicit `--base`, and module `base-branch` precedence.

## 2. TUI Base Selection Flow

- [x] 2.1 Add TUI create-flow state and fields for a selectable base branch list initialized from config `default-base` and workspace branches.
- [x] 2.2 Route valid feature-name confirmation to base branch selection/confirmation before module selection.
- [x] 2.3 Pass the confirmed TUI base branch to `CreateWorktree` without mutating `Config.DefaultBase`.
- [x] 2.4 Add or adjust TUI tests for preselected default base, branch-list navigation, fallback/cancel validation, and selected base execution.

## 3. Documentation And Validation

- [x] 3.1 Update user-facing docs for default base behavior and TUI create flow.
- [x] 3.2 Run OpenSpec validation and Go test/format checks.
