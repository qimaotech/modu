# modu TUI 规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## Purpose

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
## Requirements
### Requirement: 操作菜单显示
用户按 Enter 后，TUI SHALL 显示操作菜单，包含内置操作选项，以及当前机器已安装的配置化 app opener。

#### Scenario: 进入操作菜单
- **WHEN** 用户在列表视图按 Enter
- **THEN** TUI 显示操作菜单，当前选中第一项（打开 VS Code）

#### Scenario: feature 操作菜单显示内置选项
- **WHEN** 用户选中 feature 并进入操作菜单
- **THEN** 菜单显示内置选项 "打开 VS Code"、"打开 Codex"、"复制路径"、"更新代码"、"Modules 管理" 和 "删除"

#### Scenario: 已安装配置化 app opener 显示在菜单中
- **WHEN** `.modu.yaml` 配置了 `app: Zed`、`label: Zed`、`shortcut: z` 的 app opener 且当前机器已安装 Zed
- **THEN** 操作菜单显示 "打开 Zed (z)"，位置在内置打开项之后、非打开操作之前

#### Scenario: 未安装配置化 app opener 不显示
- **WHEN** `.modu.yaml` 配置了 `app: Cursor` 的 app opener 但当前机器未安装 Cursor
- **THEN** 操作菜单不显示 "打开 Cursor"

#### Scenario: 冲突快捷键不显示
- **WHEN** `.modu.yaml` 配置了 `app: Cursor`、`label: Cursor`、`shortcut: c` 且 Cursor 已安装
- **THEN** 操作菜单显示 "打开 Cursor" 且不显示 "(c)"，因为 `c` 已被复制路径占用

### Requirement: 操作菜单导航
用户 MUST 可以在操作菜单中使用上下键选择不同的操作。

#### Scenario: 向上导航
- **WHEN** 用户在操作菜单按向上键且不是第一项
- **THEN** 选中项向上移动一项

#### Scenario: 向下导航
- **WHEN** 用户在操作菜单按向下键且不是最后一项
- **THEN** 选中项向下移动一项

### Requirement: 操作菜单执行操作
用户 MUST 可以选择操作并执行；配置化 app opener 在可见时 MUST 支持 Enter 执行，并且仅在快捷键未冲突时支持直接按键执行。

#### Scenario: Enter 执行选中操作
- **WHEN** 用户在操作菜单按 Enter
- **THEN** TUI 执行当前选中的操作

#### Scenario: 执行删除操作
- **WHEN** 用户在操作菜单选中"删除"项并按 d
- **THEN** TUI 进入删除确认状态

#### Scenario: 执行打开 VS Code
- **WHEN** 用户在操作菜单选中"打开 VS Code"项并按 o
- **THEN** 在 VS Code 中打开主项目，然后返回列表视图

#### Scenario: Enter 执行配置化 app opener
- **WHEN** 用户在操作菜单选中 "打开 Zed (z)" 并按 Enter
- **THEN** TUI 使用配置的 app 名称打开当前选中条目的主项目路径，然后返回列表视图

#### Scenario: 唯一快捷键执行配置化 app opener
- **WHEN** 用户在操作菜单按下配置化 app opener 的唯一快捷键 `z`
- **THEN** TUI 使用该 opener 的 app 名称打开当前选中条目的主项目路径，然后返回列表视图

#### Scenario: 冲突快捷键不执行配置化 app opener
- **WHEN** Cursor opener 配置了冲突快捷键 `c` 且用户在操作菜单按 `c`
- **THEN** TUI 执行内置复制路径操作，而不是打开 Cursor

#### Scenario: 配置化 app opener 缺少可打开路径
- **WHEN** 用户对没有主项目路径的 feature 执行配置化 app opener
- **THEN** TUI 进入错误状态并显示无法打开的中文错误信息

#### Scenario: 退出操作菜单
- **WHEN** 用户在操作菜单按 esc 或 q
- **THEN** TUI 返回列表视图

### Requirement: 列表视图快捷删除
用户 MUST 可以在列表视图直接按 d 触发删除确认。

#### Scenario: 直接删除
- **WHEN** 用户在列表视图按 d
- **THEN** TUI 进入删除确认状态，显示待删除的特征名

### Requirement: TUI confirms unpushed branch deletion
The TUI SHALL require an additional confirmation before deleting a feature when the delete operation would remove local branches that are not fully pushed.

#### Scenario: Delete has no unpushed branches
- **WHEN** the user confirms feature deletion and all branch candidates are fully pushed
- **THEN** the TUI deletes the feature and returns to the refreshed list view

#### Scenario: Delete has unpushed branches
- **WHEN** the user confirms feature deletion and at least one branch candidate is not fully pushed
- **THEN** the TUI displays a second confirmation view listing the affected project or module branches

#### Scenario: User confirms unpushed deletion
- **WHEN** the unpushed-branch confirmation view is visible and the user presses `y` or Enter
- **THEN** the TUI deletes the feature with explicit unpushed-branch permission and refreshes the list

#### Scenario: User cancels unpushed deletion
- **WHEN** the unpushed-branch confirmation view is visible and the user presses `n` or Esc
- **THEN** the TUI cancels deletion and returns to the list without removing worktrees or branches

### Requirement: TUI feature creation entry
The TUI SHALL provide a list-view action for starting feature creation without changing existing list navigation, open, update, copy, module-management, delete, or quit shortcuts.

#### Scenario: Start creation from list view
- **WHEN** the user is in the list view and presses `n`
- **THEN** the TUI enters the feature-name input state for creating a new feature

#### Scenario: Existing quit shortcut remains in list view
- **WHEN** the user is in the list view and presses `q`
- **THEN** the TUI exits as before

#### Scenario: Create shortcut is shown
- **WHEN** the TUI renders the list view
- **THEN** the shortcut help includes the Chinese create-feature action

### Requirement: Editable feature-name input
The TUI SHALL let users enter and edit the feature name with cursor movement before selecting projects/modules.

#### Scenario: Printable character insertion
- **WHEN** the user types printable characters in the feature-name input state
- **THEN** the characters are inserted at the current cursor position and the cursor moves after the inserted text

#### Scenario: Cursor movement
- **WHEN** the user presses left, right, home, or end in the feature-name input state
- **THEN** the input cursor moves within the feature-name text without leaving the input state

#### Scenario: In-place deletion
- **WHEN** the user presses backspace or delete in the feature-name input state
- **THEN** the TUI removes text adjacent to the cursor without leaving the input state

#### Scenario: q is normal input text
- **WHEN** the user presses `q` in the feature-name input state
- **THEN** the TUI inserts `q` into the feature name and MUST NOT quit or cancel input

#### Scenario: Empty name is rejected
- **WHEN** the user presses Enter while the feature-name input is empty or whitespace-only
- **THEN** the TUI remains in the feature-name input state and shows a Chinese validation message

#### Scenario: Cancel feature-name input
- **WHEN** the user presses Esc or Ctrl+C in the feature-name input state
- **THEN** the TUI cancels creation and returns to the list view without creating a worktree

### Requirement: Base branch selection before creation
The TUI SHALL let users select or confirm the base branch after entering a feature name and before selecting projects/modules.

#### Scenario: Default base is preselected
- **WHEN** the user enters a valid feature name and presses Enter
- **THEN** the TUI displays a base-branch selection list with `.modu.yaml` `default-base` preselected

#### Scenario: Available branches are selectable
- **WHEN** the user presses Up/Down or k/j in the base-branch selection state
- **THEN** the TUI moves the highlighted base branch without leaving the base-branch selection state

#### Scenario: Branch list failure falls back to default
- **WHEN** the TUI cannot read the workspace branch list
- **THEN** the TUI still offers `.modu.yaml` `default-base` as the selectable base branch and shows a Chinese warning

#### Scenario: Base selection advances to modules
- **WHEN** the user confirms the highlighted base branch in the base-branch selection state
- **THEN** the TUI displays the project/module selection step

#### Scenario: Cancel base selection
- **WHEN** the user presses Esc or Ctrl+C in the base-branch selection state
- **THEN** the TUI cancels creation and returns to the list view without creating a worktree

### Requirement: Project selection before creation
The TUI SHALL provide a project/module selection step after the user confirms a feature name and base branch and before any worktree creation starts.

#### Scenario: Open selection after valid name and base
- **WHEN** the user enters a non-empty feature name, confirms a non-empty base branch, and presses Enter
- **THEN** the TUI displays a selection view for configured projects/modules before creating the feature

#### Scenario: Main project is included
- **WHEN** the user reaches the project/module selection step
- **THEN** the main project is treated as always included for feature creation

#### Scenario: Configured modules are selectable
- **WHEN** the project/module selection step is displayed
- **THEN** each configured module is shown as selectable and can be toggled before confirmation

#### Scenario: Default and remote branch preselection
- **WHEN** configured default-selected modules or modules with the target remote branch exist
- **THEN** those modules are preselected in the project/module selection step

#### Scenario: Selection cancellation
- **WHEN** the user cancels the project/module selection step
- **THEN** the TUI returns to the list view without creating a worktree

### Requirement: Create feature from TUI selection
The TUI SHALL create the feature worktree from the confirmed base branch after the user confirms project/module selection.

#### Scenario: Create selected projects
- **WHEN** the user confirms project/module selection for a feature name and base branch
- **THEN** the TUI creates the main project worktree from the confirmed base branch and only the selected module worktrees

#### Scenario: Create with no selected modules
- **WHEN** the user confirms creation with no modules selected
- **THEN** the TUI creates the main project feature worktree from the confirmed base branch without module worktrees

#### Scenario: Module base branch still wins
- **WHEN** the user confirms creation with a selected module that configures `base-branch`
- **THEN** that module is created from its configured `base-branch`

#### Scenario: Creation success feedback
- **WHEN** feature creation succeeds
- **THEN** the TUI reloads the list view and shows a Chinese success message containing the feature name

#### Scenario: Creation failure feedback
- **WHEN** feature creation fails
- **THEN** the TUI enters the error state and shows the creation error

### Requirement: App opener availability detection

TUI SHALL determine configured app opener visibility at startup by checking whether each configured app can be resolved on the current machine.

#### Scenario: Installed app is available
- **WHEN** a configured app opener refers to an installed app
- **THEN** TUI includes that opener in the operation menu

#### Scenario: Missing app is unavailable
- **WHEN** a configured app opener refers to an app that is not installed or cannot be resolved
- **THEN** TUI excludes that opener from the operation menu without entering the error state

#### Scenario: Availability is evaluated before rendering menu
- **WHEN** the TUI first enters the list view after startup
- **THEN** the operation menu model already contains only available configured app openers

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
