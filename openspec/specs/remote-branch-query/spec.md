## ADDED Requirements

### Requirement: Remote branch exists query
The system SHALL provide a method to query whether a remote repository contains a specific branch.

#### Scenario: Branch exists on remote
- **WHEN** `RemoteBranchExists(ctx, "https://github.com/org/repo.git", "feature/abc")` is called and the branch exists on the remote
- **THEN** it returns `true`

#### Scenario: Branch does not exist on remote
- **WHEN** `RemoteBranchExists(ctx, "https://github.com/org/repo.git", "feature/xyz")` is called and the branch does not exist on the remote
- **THEN** it returns `false`

#### Scenario: Repository does not exist
- **WHEN** `RemoteBranchExists(ctx, "https://github.com/org/nonexistent.git", "feature/abc")` is called and the repository does not exist
- **THEN** it returns `false`

#### Scenario: Network error
- **WHEN** `RemoteBranchExists(ctx, "https://github.com/org/repo.git", "feature/abc")` is called and a network error occurs
- **THEN** it returns `false`

### Requirement: Concurrent remote branch query for all modules
The engine SHALL query all configured modules concurrently to determine which ones have a specific remote branch.

#### Scenario: All modules have the branch
- **WHEN** `GetModulesWithRemoteBranch(ctx, "feature/abc")` is called and all modules have the branch
- **THEN** it returns a map where all module names map to `true`

#### Scenario: No modules have the branch
- **WHEN** `GetModulesWithRemoteBranch(ctx, "feature/abc")` is called and no modules have the branch
- **THEN** it returns a map where all module names map to `false`

#### Scenario: Partial modules have the branch
- **WHEN** `GetModulesWithRemoteBranch(ctx, "feature/abc")` is called and only some modules have the branch
- **THEN** it returns a map where only the modules with the branch map to `true`

#### Scenario: Some module URLs are empty
- **WHEN** `GetModulesWithRemoteBranch(ctx, "feature/abc")` is called and some modules have empty URLs
- **THEN** those modules with empty URLs map to `false` in the result

### Requirement: Module selector pre-selects based on remote branch
The module selector SHALL pre-select modules that have the remote branch, in addition to locally existing modules.

#### Scenario: Pre-select when remote has branch
- **WHEN** `SelectModules(modules, existingModules, remoteHasBranch)` is called with a module that has the remote branch but is not locally existing
- **THEN** that module is pre-selected in the UI

#### Scenario: Pre-select when both local and remote have branch
- **WHEN** `SelectModules(modules, existingModules, remoteHasBranch)` is called with a module that is locally existing and has the remote branch
- **THEN** that module is pre-selected in the UI

#### Scenario: Do not pre-select when remote does not have branch
- **WHEN** `SelectModules(modules, existingModules, remoteHasBranch)` is called with a module that is not locally existing and does not have the remote branch
- **THEN** that module is not pre-selected in the UI
