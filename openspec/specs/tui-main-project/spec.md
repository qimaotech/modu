# TUI 主项目支持规范

**版本**: 1.0

## 目的

在 TUI 列表中展示主项目（workspace 主仓库），支持打开 VS Code 和一键更新代码。

## 列表展示

### Requirement: 主项目条目显示

TUI 列表 SHALL 在顶部固定显示主项目条目。

#### Scenario: 主项目显示在列表顶部
- **WHEN** TUI 加载完成并显示列表
- **THEN** 第一行显示主项目，格式为 "→ <主项目名> [主项目] (<dirty状态>) [<分支名>]"

#### Scenario: 主项目无 dirty
- **WHEN** 主项目没有未提交修改
- **THEN** 显示 "clean" 状态，颜色为绿色

#### Scenario: 主项目有 dirty
- **WHEN** 主项目存在未提交修改
- **THEN** 显示 "dirty" 状态，颜色为红色

### Requirement: Feature 列表显示

TUI SHALL 在主项目下方显示所有 feature 分支列表。

#### Scenario: Feature 列表顺序
- **WHEN** 列表显示
- **THEN** Feature 按现有顺序显示在主项目下方

## 主项目菜单

### Requirement: 主项目操作菜单

用户在主项目条目上按 Enter 后，TUI SHALL 显示主项目专用操作菜单。

#### Scenario: 进入主项目菜单
- **WHEN** 用户在主项目条目上按 Enter
- **THEN** TUI 显示操作菜单，包含 "打开 VS Code" 和 "更新代码" 两个选项

#### Scenario: 主项目菜单打开 VS Code
- **WHEN** 用户在主项目菜单选中 "打开 VS Code" 并按 Enter 或 o
- **THEN** 在 VS Code 中打开主项目路径，然后返回列表视图

#### Scenario: 主项目菜单更新代码
- **WHEN** 用户在主项目菜单选中 "更新代码" 并按 Enter 或 u
- **THEN** TUI 进入 loading 状态，执行 git fetch + rebase 更新主项目和所有 modules，完成后返回列表视图并显示结果消息

### Requirement: 主项目菜单快捷键

用户 MUST 可以使用快捷键直接执行主项目操作。

#### Scenario: 主项目快捷键 o
- **WHEN** 用户在主项目条目上按 o
- **THEN** 在 VS Code 中打开主项目，然后返回列表视图

#### Scenario: 主项目快捷键 u
- **WHEN** 用户在主项目条目上按 u
- **THEN** 执行 git fetch + rebase 更新主项目和所有 modules，然后返回列表视图

### Requirement: 主项目菜单导航

用户 MUST 可以在主项目菜单中使用上下键选择不同的操作。

#### Scenario: 主项目菜单向上导航
- **WHEN** 用户在主项目菜单按向上键且不是第一项
- **THEN** 选中项向上移动一项

#### Scenario: 主项目菜单向下导航
- **WHEN** 用户在主项目菜单按向下键且不是最后一项
- **THEN** 选中项向下移动一项

#### Scenario: 退出主项目菜单
- **WHEN** 用户在主项目菜单按 esc 或 q
- **THEN** TUI 返回列表视图

## 更新代码

### Requirement: 更新代码执行

Engine SHALL 实现 UpdateMainProject 方法，并发执行主项目和所有模块的 git fetch + rebase。

#### Scenario: 更新代码成功
- **WHEN** 用户触发更新代码操作且所有仓库更新成功
- **THEN** 显示成功消息 "更新成功: 主项目 + X 个模块"

#### Scenario: 更新代码部分失败
- **WHEN** 用户触发更新代码操作但部分仓库更新失败
- **THEN** 显示部分成功消息 "更新成功: X 个，失败: Y 个 (模块名列表)"

#### Scenario: 更新代码无配置模块
- **WHEN** 用户触发更新代码操作但没有配置任何模块
- **THEN** 仅更新主项目，显示 "更新成功: 主项目"

## 交互流程

### Requirement: 主项目与 Feature 菜单差异

TUI SHALL 根据选中项类型（主项目或 Feature）显示不同的菜单。

#### Scenario: 主项目显示简化菜单
- **WHEN** 选中主项目时显示菜单
- **THEN** 菜单只包含 "打开 VS Code" 和 "更新代码"

#### Scenario: Feature 显示完整菜单
- **WHEN** 选中 Feature 时显示菜单
- **THEN** 菜单包含 "打开 VS Code"、"Modules 管理"、"删除"

### Requirement: 快捷键上下文

TUI SHALL 根据选中项类型处理快捷键，无效快捷键不触发任何操作。

#### Scenario: 主项目无效快捷键
- **WHEN** 选中主项目时按 m 或 d
- **THEN** 不触发任何操作，保持在当前视图

#### Scenario: Feature 无效快捷键
- **WHEN** 选中 Feature 时按 u
- **THEN** 不触发任何操作，保持在当前视图
