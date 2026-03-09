## MODIFIED Requirements

### Requirement: AddModule reuses existing branch
When adding a module to an existing feature, if the remote repository already has a branch with the same name and it's not used by any worktree, the system SHALL reuse that branch to create the worktree.

#### Scenario: AddModule reuses existing branch
- **WHEN** user runs `modu add-module <feature> <module>` and module's remote repository already has a branch with the feature name
- **AND** the branch is not used by any worktree
- **THEN** system creates worktree from the existing branch without creating a new branch

#### Scenario: AddModule skips when branch is used by other worktree
- **WHEN** user runs `modu add-module <feature> <module>` and module's remote repository already has a branch with the feature name
- **AND** the branch is used by another worktree
- **THEN** system outputs "[SKIP] <module>: 分支 <branch> 已被其他 worktree 使用"
- **AND** returns success without creating worktree

#### Scenario: AddModule creates new branch when not exists
- **WHEN** user runs `modu add-module <feature> <module>` and module's remote repository does not have a branch with the feature name
- **THEN** system creates a new branch from base branch and creates worktree
