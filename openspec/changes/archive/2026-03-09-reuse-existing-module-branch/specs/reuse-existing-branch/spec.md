## ADDED Requirements

### Requirement: Module branch exists and not in use
当创建 feature 时，如果 module 的远程仓库已存在同名分支，且该分支未被任何 worktree 使用，系统 SHALL 直接使用该分支创建 worktree。

#### Scenario: Reuse existing branch for module
- **WHEN** 用户执行 `modu create feature feature-xxx`，module-a 的远程仓库已存在 `feature-xxx` 分支
- **AND** 该分支未被任何 worktree 使用
- **THEN** 系统直接 checkout 该分支创建 worktree，无需创建新分支

### Requirement: Module branch exists and in use by other worktree
当创建 feature 时，如果 module 的远程仓库已存在同名分支，但该分支已被其他 worktree 使用，系统 SHALL 跳过该 module 并输出明确提示。

#### Scenario: Skip module when branch is used by other worktree
- **WHEN** 用户执行 `modu create feature feature-xxx`，module-a 的远程仓库已存在 `feature-xxx` 分支
- **AND** 该分支已被 worktree `/path/to/other-worktree` 使用
- **THEN** 系统跳过 module-a 的 worktree 创建
- **AND** 输出提示 "[SKIP] module-a: 分支 feature-xxx 已被其他 worktree 使用"

### Requirement: Module branch does not exist
当创建 feature 时，如果 module 的远程仓库不存在同名分支，系统 SHALL 创建新分支（现有逻辑不变）。

#### Scenario: Create new branch for module
- **WHEN** 用户执行 `modu create feature feature-xxx`，module-a 的远程仓库不存在 `feature-xxx` 分支
- **THEN** 系统从 base 分支创建新分支并创建 worktree

### Requirement: Main project always creates new branch
主项目（workspace）保持现有逻辑，始终创建新分支，不复用已存在的分支。

#### Scenario: Main project creates new branch even if exists
- **WHEN** 用户执行 `modu create feature feature-xxx`，主项目的远程仓库已存在 `feature-xxx` 分支
- **THEN** 系统仍然创建新分支（如果已存在则失败，或覆盖，取决于现有实现）

### Requirement: Summary output after creation
创建完成后，系统 SHALL 输出 summary 显示创建成功和跳过的 module 数量。

#### Scenario: Display summary after feature creation
- **WHEN** 创建 feature 完成
- **THEN** 输出类似 "创建成功: 3 个模块，跳过: 1 个模块" 的 summary
