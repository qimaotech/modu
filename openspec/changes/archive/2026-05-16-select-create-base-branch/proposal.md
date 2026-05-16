## Why

`modu create` should make the base branch explicit and predictable for second-stage development or feature testing. Today the CLI flag exists but defaults to `develop` in code, while the TUI silently uses `.modu.yaml` `default-base`, so users cannot consistently confirm or change the branch before creating a worktree.

## What Changes

- Change CLI `modu create <feature>` so an omitted `--base` uses `.modu.yaml` `default-base`.
- Keep CLI `--base <branch>` as the explicit override for the created worktree base branch.
- Add a TUI creation step that lists available branches and lets users select or confirm the base branch after entering the feature name and before selecting modules.
- Preserve existing module-level `base-branch` behavior: a configured module `base-branch` continues to override the selected global base for that module.

## Capabilities

### New Capabilities

- None

### Modified Capabilities

- `cli`: `modu create` default base branch behavior changes from a hard-coded default to config-driven default-base.
- `tui`: TUI feature creation adds an explicit base branch selection/confirmation step before project/module selection.

## Impact

- `cmd/modu/main.go`: create command flag default and run-time base resolution.
- `internal/engine/engine.go`, `internal/gitproxy/gitproxy_impl.go`: branch-list discovery for TUI base selection.
- `internal/ui/ui.go`: feature creation state machine, rendering, and create execution.
- `internal/ui/ui_test.go`, `internal/engine/engine_test.go`, `internal/gitproxy/gitproxy_test.go`, `cmd/modu/main_test.go`: coverage for CLI/TUI base branch behavior.
- `README.md` and OpenSpec specs for CLI/TUI behavior.
