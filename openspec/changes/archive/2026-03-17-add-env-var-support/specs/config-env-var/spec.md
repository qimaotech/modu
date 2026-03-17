## ADDED Requirements

### Requirement: 支持 $VAR 格式环境变量
配置文件中的 workspace 或 worktree-root 字段如果包含 $VAR 格式的环境变量，该环境变量 SHALL 正确展开为其对应的值。

#### Scenario: 环境变量已定义
- **GIVEN** 配置文件中 workspace 字段为 "$MY_WORKSPACE"
- **AND** 环境变量 MY_WORKSPACE 已设置为 "/opt/workspace"
- **WHEN** modu 加载配置文件
- **THEN** workspace 字段的值应为 "/opt/workspace"

#### Scenario: 环境变量未定义
- **GIVEN** 配置文件中 worktree-root 字段为 "$UNDEFINED_VAR"
- **AND** 环境变量 UNDEFINED_VAR 未设置
- **WHEN** modu 加载配置文件
- **THEN** 加载应失败并报错，提示 worktree-root 字段的未定义环境变量

### Requirement: 支持 ${VAR} 格式环境变量
配置文件中的 workspace 或 worktree-root 字段如果包含 ${VAR} 格式的环境变量，该环境变量 SHALL 正确展开为其对应的值。

#### Scenario: ${VAR} 格式环境变量已定义
- **GIVEN** 配置文件中 workspace 字段为 "${MY_WORKSPACE}"
- **AND** 环境变量 MY_WORKSPACE 已设置为 "/opt/workspace"
- **WHEN** modu 加载配置文件
- **THEN** workspace 字段的值应为 "/opt/workspace"

#### Scenario: ${VAR} 格式环境变量未定义
- **GIVEN** 配置文件中 worktree-root 字段为 "${UNDEFINED_VAR}"
- **AND** 环境变量 UNDEFINED_VAR 未设置
- **WHEN** modu 加载配置文件
- **THEN** 加载应失败并报错，提示 worktree-root 字段的未定义环境变量

#### Scenario: ${VAR} 格式带默认值语法不应展开
- **GIVEN** 配置文件中 workspace 字段为 "${MY_WORKSPACE:-default}"
- **AND** 环境变量 MY_WORKSPACE 未设置
- **WHEN** modu 加载配置文件
- **THEN** 加载应失败并报错（不支持默认值语法）

### Requirement: 路径包含环境变量
环境变量可以出现在路径的任意位置。

#### Scenario: 环境变量在路径中间
- **GIVEN** 配置文件中 workspace 字段为 "/home/$USER/workspace"
- **AND** 环境变量 USER 已设置为 "john"
- **WHEN** modu 加载配置文件
- **THEN** workspace 字段的值应为 "/home/john/workspace"

### Requirement: 未配置字段不检查环境变量
如果 workspace 或 worktree-root 字段在配置文件中未设置，则不进行环境变量检查。

#### Scenario: workspace 未配置
- **GIVEN** 配置文件中未设置 workspace 字段
- **AND** worktree-root 字段设置为 "/opt/worktrees"
- **WHEN** modu 加载配置文件
- **THEN** 加载成功（验证会检查 worktree-root，不检查 workspace）

### Requirement: 无环境变量的配置正常工作
配置文件中不包含环境变量的现有配置 SHALL 继续正常工作。

#### Scenario: 无环境变量配置
- **GIVEN** 配置文件中 workspace 字段为 "/opt/workspace"
- **AND** worktree-root 字段为 "/opt/worktrees"
- **WHEN** modu 加载配置文件
- **THEN** 加载成功，字段值保持原样
