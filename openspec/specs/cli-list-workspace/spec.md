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

### Requirement: 按需检查状态

默认文本列表（包括 -a）SHALL 每个源仓库读取一次 worktree 注册信息，不执行文件状态扫描。

#### Scenario: 大量 feature
- **WHEN** 用户执行默认文本 list
- **THEN** 元数据 Git 调用次数随源仓库数量增长，不随 feature 数量成倍增长

#### Scenario: 请求文件状态
- **WHEN** 用户使用 --status、status 命令或 JSON 列表
- **THEN** 每个 worktree 通过一次 porcelain v2 调用读取分支和脏状态，并使用全局有界并发
