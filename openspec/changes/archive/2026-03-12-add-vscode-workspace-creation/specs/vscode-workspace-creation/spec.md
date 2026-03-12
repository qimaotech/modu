## ADDED Requirements

### Requirement: VSCode workspace 文件自动创建
当 `modu create` 命令成功创建 feature worktree 时，系统 SHALL 自动生成一个 `.code-workspace` 文件。

#### Scenario: create 成功时生成 workspace 文件
- **WHEN** 用户执行 `modu create <feature>` 且 worktree 创建成功
- **THEN** 系统在 feature 目录下生成 `<feature>.code-workspace` 文件

#### Scenario: workspace 只包含实际存在的模块
- **WHEN** workspace 文件被生成
- **THEN** workspace 的 folders 数组只包含 feature 目录下实际存在的模块（配置的 modules 与实际目录的交集），跳过非模块目录如 .git, .claude 等

#### Scenario: workspace 包含标准 Go 开发配置
- **WHEN** workspace 文件被生成
- **THEN** settings 包含 go.toolsManagement.autoUpdate、go.lintTool、go.formatTool 等标准配置

#### Scenario: create 失败时不生成 workspace 文件
- **WHEN** 用户执行 `modu create <feature>` 且 worktree 创建失败
- **THEN** 系统不生成 workspace 文件

#### Scenario: 添加模块后更新 workspace
- **WHEN** 用户在 TUI 中通过 "m" 键添加模块
- **THEN** 系统更新 workspace 文件，添加新模块到 folders

#### Scenario: 删除模块后更新 workspace
- **WHEN** 用户在 TUI 中通过 "m" 键删除模块
- **THEN** 系统更新 workspace 文件，从 folders 移除已删除的模块
