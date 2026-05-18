## ADDED Requirements

### Requirement: Engine preflights unpushed branch deletion
The engine SHALL identify all local branches that would be deleted by a feature deletion before removing any worktree.

#### Scenario: Candidate branch is fully pushed
- **WHEN** a candidate branch has no local commits ahead of its remote branch
- **THEN** the engine treats that branch as safe for deletion

#### Scenario: Candidate branch is not fully pushed
- **WHEN** a candidate branch has no remote branch, no usable upstream, a failed push-state lookup, or local commits ahead of its remote branch
- **THEN** the engine reports that branch as requiring confirmation before deletion

#### Scenario: Delete without confirmation has unpushed branches
- **WHEN** `DeleteWorktree` is called for a feature with unpushed branch candidates
- **THEN** the engine returns `ERR_UNPUSHED_BRANCH` without removing worktrees or local branches

#### Scenario: Delete with explicit confirmation has unpushed branches
- **WHEN** an options-based delete call explicitly allows unpushed branch deletion
- **THEN** the engine may remove the worktrees and matching local branches after normal dirty-check rules pass

#### Scenario: Branch is not a deletion candidate
- **WHEN** a worktree is detached, has no readable branch, or its branch slug does not match the feature directory name
- **THEN** the engine does not require unpushed-branch confirmation for that worktree because the branch will not be deleted
