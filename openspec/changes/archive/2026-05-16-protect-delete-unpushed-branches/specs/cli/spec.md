## ADDED Requirements

### Requirement: Delete protects unpushed local branches
The `modu delete` command SHALL check whether local branches that would be deleted are fully pushed before removing worktrees or local branches.

#### Scenario: All branches are pushed
- **WHEN** the user runs `modu delete <feature>` and every branch that would be deleted has no local commits ahead of its remote branch
- **THEN** deletion proceeds without an additional unpushed-branch confirmation

#### Scenario: Delete has unpushed branches without explicit allow flag
- **WHEN** the user runs `modu delete <feature>` and at least one branch that would be deleted is not fully pushed
- **THEN** the command fails with `ERR_UNPUSHED_BRANCH` without deleting worktrees or local branches
- **THEN** the error message tells the user to rerun with `--allow-unpushed` to confirm deletion

#### Scenario: Delete has unpushed branches with explicit allow flag
- **WHEN** the user runs `modu delete <feature> --allow-unpushed` and at least one branch that would be deleted is not fully pushed
- **THEN** deletion proceeds without interactive confirmation

#### Scenario: Force does not bypass unpushed protection
- **WHEN** the user runs `modu delete <feature> --force` and at least one branch that would be deleted is not fully pushed
- **THEN** the command fails with `ERR_UNPUSHED_BRANCH` unless `--allow-unpushed` is also provided
