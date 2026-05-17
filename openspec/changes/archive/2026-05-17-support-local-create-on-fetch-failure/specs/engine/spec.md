## ADDED Requirements

### Requirement: CreateWorktree supports explicit local-only creation
The engine SHALL support feature creation with an explicit option to skip pre-create fetch and use local repository state.

#### Scenario: Local-only creation skips fetch
- **WHEN** `CreateWorktree` is invoked with fetch disabled
- **THEN** the engine creates the main project and selected module worktrees without pre-create fetch attempts

#### Scenario: Fetch-enabled creation fails on fetch failure
- **WHEN** `CreateWorktree` is invoked with fetch enabled and a repository fetch fails before worktree creation
- **THEN** the engine returns an error and rolls back any worktrees already created in the current operation

#### Scenario: Auto-fetch disabled skips pre-create fetch
- **WHEN** the loaded config has `auto-fetch: false`
- **THEN** `CreateWorktree` creates main project and selected module worktrees without pre-create fetch attempts

#### Scenario: Local-only creation preserves rollback
- **WHEN** fetch is disabled but a later `git worktree add` fails
- **THEN** the engine rolls back successfully created worktrees using the existing rollback behavior

#### Scenario: Remote branch reuse remains preferred when known
- **WHEN** remote branch detection succeeds and reports that a selected module has the target branch
- **THEN** the engine continues to create that module worktree from the remote branch instead of creating a new local branch
