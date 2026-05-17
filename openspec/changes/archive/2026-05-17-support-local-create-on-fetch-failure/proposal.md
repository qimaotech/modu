## Why

Creating a feature currently depends on fetching remote branches first. When the network is unavailable or `git fetch` fails, users cannot create a worktree even if their local base branch is good enough to continue development.

## What Changes

- Add an explicit CLI parameter for local-only feature creation that skips remote fetch.
- Keep CLI fetch failure behavior compatible with the current flow: report the original error, then add a Chinese hint telling users to rerun with the local-only parameter when appropriate.
- In TUI create, show the concrete fetch error first and then ask whether to retry creation from local repository state.
- Make the existing `auto-fetch` configuration meaningful for create flows: enabled keeps the current fetch-first behavior, disabled skips pre-create fetch and uses local branches directly.
- Keep rollback behavior unchanged if local fallback creation later fails.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `config`: Clarify `auto-fetch` behavior for feature creation.
- `cli`: Add an explicit local-only create parameter and fetch-failure hint.
- `engine`: Update `CreateWorktree` behavior to support explicit local-only creation.
- `gitproxy`: Add/create-without-fetch behavior needed by the engine while preserving contextual git errors.
- `tui`: Add a confirmation prompt and feedback path for local fallback during TUI feature creation.

## Impact

- Affected code: `internal/config`, `internal/engine`, `internal/gitproxy`, `internal/ui`, `cmd/modu`.
- Tests: unit tests for no-fetch create decisions, gitproxy no-fetch worktree creation, config behavior, CLI hint output, and TUI confirmation states.
- User-visible behavior: CLI fetch failures keep the current error path with an added local-create hint; TUI fetch failures show the error and offer an in-place local retry.
