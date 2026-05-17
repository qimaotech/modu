## ADDED Requirements

### Requirement: Worktree creation can skip fetch
Gitproxy SHALL provide worktree creation behavior that can add a worktree without performing a preceding fetch when the caller explicitly requests local-only creation.

#### Scenario: Create new worktree without fetch
- **WHEN** gitproxy is asked to create a worktree with fetch disabled
- **THEN** it runs `git worktree add -b <branch> <worktreePath> <baseBranch>` without first running `git fetch`

#### Scenario: Reuse existing branch without fetch
- **WHEN** gitproxy is asked to create a worktree from an existing local branch with fetch disabled
- **THEN** it runs `git worktree add <worktreePath> <branch>` without first running `git fetch`

#### Scenario: Fetch failure remains distinguishable
- **WHEN** pre-create fetch fails before any `git worktree add` command runs
- **THEN** gitproxy returns an error that callers can identify as a fetch failure while still wrapping `ERR_GIT_EXEC`

#### Scenario: Worktree add failure keeps context
- **WHEN** local-only worktree creation fails
- **THEN** gitproxy returns an error containing the operation, repository path, target branch or base branch, worktree path, and git output
