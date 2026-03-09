## ADDED Requirements

### Requirement: 操作菜单显示
用户按 Enter 后，TUI SHALL 显示操作菜单，包含可用的操作选项。

#### Scenario: 进入操作菜单
- **WHEN** 用户在列表视图按 Enter
- **THEN** TUI 显示操作菜单，当前选中第一项（打开 VS Code）

#### Scenario: 操作菜单显示正确选项
- **WHEN** 操作菜单显示
- **THEN** 显示"打开 VS Code"和"删除"两个选项（打开在前，删除在后）

### Requirement: 操作菜单导航
用户 MUST 可以在操作菜单中使用上下键选择不同的操作。

#### Scenario: 向上导航
- **WHEN** 用户在操作菜单按向上键且不是第一项
- **THEN** 选中项向上移动一项

#### Scenario: 向下导航
- **WHEN** 用户在操作菜单按向下键且不是最后一项
- **THEN** 选中项向下移动一项

### Requirement: 操作菜单执行操作
用户 MUST 可以选择操作并执行。

#### Scenario: Enter 执行选中操作
- **WHEN** 用户在操作菜单按 Enter
- **THEN** TUI 执行当前选中的操作

#### Scenario: 执行删除操作
- **WHEN** 用户在操作菜单选中"删除"项并按 d
- **THEN** TUI 进入删除确认状态

#### Scenario: 执行打开 VS Code
- **WHEN** 用户在操作菜单选中"打开 VS Code"项并按 o
- **THEN** 在 VS Code 中打开主项目，然后返回列表视图

#### Scenario: 退出操作菜单
- **WHEN** 用户在操作菜单按 esc 或 q
- **THEN** TUI 返回列表视图

### Requirement: 列表视图快捷删除
用户 MUST 可以在列表视图直接按 d 触发删除确认。

#### Scenario: 直接删除
- **WHEN** 用户在列表视图按 d
- **THEN** TUI 进入删除确认状态，显示待删除的特征名
