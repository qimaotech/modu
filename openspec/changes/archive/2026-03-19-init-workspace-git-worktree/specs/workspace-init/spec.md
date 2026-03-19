# workspace-init 规范

**版本**: 1.0

## ADDED Requirements

### Requirement: 配置保存到 workspace 目录

配置向导保存 `.modu.yaml` 时，应保存到用户输入的 `workspace` 目录，而非当前工作目录。

### Requirement: workspace 目录自动创建

如果 `workspace` 目录不存在，配置向导 SHALL 自动创建该目录。

### Requirement: workspace Git 仓库初始化

如果 `workspace` 目录不是 Git 仓库，配置向导 SHALL 自动执行 `git init` 并 `git checkout -b <default-base>` 创建默认分支。

#### Scenario: workspace 目录不存在且非 git 仓库
- **WHEN** 用户在配置向导完成所有步骤并确认保存
- **THEN** 系统自动创建 workspace 目录
- **AND** 系统自动执行 `git init`
- **AND** 系统自动创建并切换到 default-base 分支（如 `develop`）

#### Scenario: workspace 目录已存在且是 git 仓库
- **WHEN** 用户在配置向导完成所有步骤并确认保存
- **THEN** 系统跳过 git 初始化
- **AND** 直接保存配置文件到 workspace 目录

#### Scenario: workspace 目录已存在但非 git 仓库
- **WHEN** 用户在配置向导完成所有步骤并确认保存
- **THEN** 系统自动执行 `git init`
- **AND** 系统自动创建并切换到 default-base 分支

### Requirement: worktree-root 目录自动创建

如果 `worktree-root` 目录不存在，配置向导 SHALL 自动创建该目录。

#### Scenario: worktree-root 目录不存在
- **WHEN** 用户在配置向导完成所有步骤并确认保存
- **THEN** 系统自动创建 worktree-root 目录

### Requirement: Git 操作失败时阻止保存

如果任何 Git 操作（`git init`、`git checkout -b`）失败，配置向导 SHALL 返回错误并阻止保存配置文件。

#### Scenario: git init 失败
- **WHEN** 执行 `git init` 时发生错误（如权限不足）
- **THEN** 系统显示错误信息并阻止保存配置文件
- **AND** 用户可重试或退出

#### Scenario: git checkout -b 失败
- **WHEN** 执行 `git checkout -b <branch>` 时分支创建失败（如分支已存在）
- **THEN** 系统显示错误信息并阻止保存配置文件
- **AND** 用户可重试或退出
