## ADDED Requirements

### Requirement: Create uses configured default base
The CLI SHALL use the loaded configuration `default-base` as the base branch for `modu create <feature>` when the user does not provide `--base`.

#### Scenario: Create without base flag
- **WHEN** the user runs `modu create feature/foo` and `.modu.yaml` contains `default-base: release/1.2`
- **THEN** the system creates the main project worktree from `release/1.2`
- **AND** modules without module-level `base-branch` are created from `release/1.2`

#### Scenario: Create with explicit base flag
- **WHEN** the user runs `modu create feature/foo --base main` and `.modu.yaml` contains `default-base: develop`
- **THEN** the system creates the main project worktree from `main`
- **AND** modules without module-level `base-branch` are created from `main`

#### Scenario: Module base branch still wins
- **WHEN** the user runs `modu create feature/foo --base main`
- **AND** a selected module has `base-branch: release/module`
- **THEN** that module worktree is created from `release/module`
