# modu list workspace 显示规范

**版本**: 1.0 | **来源**: add-list-workspace-flag change

## 目的

定义 `modu list` 命令的 `-a/--all` flag 行为，用于显示主项目（workspace）及其模块的分支信息。

## 需求

### Requirement: modu list command supports -a flag to show workspace
The `modu list` command SHALL support a `-a` or `--all` flag that, when provided, displays the workspace (main project) information along with all its modules' branch status.

#### Scenario: list without -a flag
- **WHEN** user runs `modu list` without any flags
- **THEN** the command displays only the feature worktrees (current behavior)

#### Scenario: list with -a flag
- **WHEN** user runs `modu list -a`
- **THEN** the command displays workspace information followed by feature worktrees

#### Scenario: list with --all flag
- **WHEN** user runs `modu list --all`
- **THEN** the command displays workspace information followed by feature worktrees (same as -a)

### Requirement: Workspace information displays branch status
When `-a` flag is used, the workspace information SHALL include:
- The workspace name
- The current branch name
- All modules under workspace with their current branch names

#### Scenario: Workspace branch display format
- **WHEN** workspace is on `develop` branch with modules `pixiu-ad-backend`, `pixiu-frontend`
- **THEN** output shows:
  ```
  Workspace [develop]
    - pixiu-ad-backend: develop
    - pixiu-frontend: develop
  ```

### Requirement: Workspace displayed above features
When `-a` flag is used, the workspace information SHALL be displayed before the feature list.

#### Scenario: Output order with -a flag
- **WHEN** user runs `modu list -a`
- **THEN** workspace section appears first, followed by Features section
