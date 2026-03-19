# tui-copy-path

## ADDED Requirements

### Requirement: Copy path in list view

用户可在列表视图（list state）按 `c` 键复制当前选中项的主项目绝对路径到剪贴板。

#### Scenario: Copy main project path in list view

- **WHEN** 用户在列表视图选中主项目并按 `c`
- **THEN** 系统复制主项目的绝对路径到剪贴板，并显示临时消息 "路径已复制: <path>"

#### Scenario: Copy feature path in list view

- **WHEN** 用户在列表视图选中 feature 并按 `c`
- **THEN** 系统复制该 feature 对应主项目的绝对路径到剪贴板，并显示临时消息 "路径已复制: <path>"

### Requirement: Copy path in menu view

用户可在操作菜单（menu state）按 `c` 键或选择菜单项复制当前选中项的主项目绝对路径。

#### Scenario: Copy path via menu item

- **WHEN** 用户在操作菜单中选中 "复制路径" 菜单项并按 Enter
- **THEN** 系统复制对应主项目的绝对路径到剪贴板，关闭菜单，显示临时消息 "路径已复制: <path>"

#### Scenario: Copy path via shortcut key in menu

- **WHEN** 用户在操作菜单中按 `c` 键
- **THEN** 系统复制对应主项目的绝对路径到剪贴板，关闭菜单，显示临时消息 "路径已复制: <path>"

### Requirement: Error handling for missing main project

当 feature 没有关联主项目时，复制操作应显示错误消息。

#### Scenario: Copy path when main project is missing

- **WHEN** 用户尝试复制一个无主项目的 feature 路径
- **THEN** 系统显示错误消息 "该 feature 无主项目，无法复制路径"

### Requirement: Clipboard write

系统应使用跨平台剪贴板库将路径写入系统剪贴板。

#### Scenario: Write to clipboard on macOS

- **WHEN** 系统在 macOS 环境下执行复制路径
- **THEN** 系统调用 `pbcopy` 将路径写入剪贴板

#### Scenario: Write to clipboard on Linux

- **WHEN** 系统在 Linux 环境下执行复制路径
- **THEN** 系统调用 `xclip` 或 `xsel` 将路径写入剪贴板
