# modu TUI 规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## 目的

基于 Bubble Tea 的状态机 TUI，提供 worktree 列表查看、创建/删除 worktree，删除前二次确认。

## 入口

- 裸命令 `modu` 且在交互式终端时进入 TUI。
- 命令 `modu tui` 显式启动 TUI；无配置文件时可启动配置向导。

## 状态机

1. **LoadingState**：并发执行 init 或 create 时，显示每个 Module 的当前进度（如 `api-server: Cloning...`）。
2. **ListState**：展示所有 feature 列表；光标选中时展示该环境下各模块的 Branch 与 Status（Clean/Dirty）。
3. **ConfirmState**：删除前的二次确认。
4. **ErrorState**：操作失败时显示错误详情，允许重试。

## UI 表现

- 多行并行进度（Multi-Spinner），实时显示各模块任务状态。
- 键盘：上下选择、回车确认、ESC 取消。

## 能力范围

- **只读**：worktree 列表、分支、模块状态。
- **写**：创建 worktree、删除 worktree（删除前必须确认；脏检查由 engine 执行，TUI 展示结果）。

## 与代码的对应

- 实现：`internal/ui`（Bubble Tea 模型、状态转换）；`internal/ui/config_wizard.go`（配置向导）。

## 操作菜单

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

## 用户可见文案（中文）

TUI 面向用户的所有提示 SHALL 使用中文。删除与错误相关文案如下（来源：docs/plans/2026-03-09-delete-prompts-localization.md）：

| 场景 | 文案 |
|------|------|
| 删除确认标题 | 确认删除 |
| 删除确认说明 | 确定要删除 feature「%s」吗？ |
| 删除确认操作提示 | 按 y 确认，n 取消 |
| 删除成功反馈 | 已删除 feature: &lt;feature&gt; |
| 错误界面标题 | 错误 |
| 错误界面继续提示 | 按任意键继续... |
