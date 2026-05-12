## MODIFIED Requirements

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

## ADDED Requirements

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
