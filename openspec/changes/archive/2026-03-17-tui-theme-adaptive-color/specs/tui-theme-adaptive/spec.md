## ADDED Requirements

### Requirement: TUI 自动适配系统终端背景主题
TUI 界面 SHALL 根据系统终端的背景颜色自动选择深色或浅色配色方案，确保在任意背景下都能清晰可读。

#### Scenario: 深色终端背景
- **WHEN** TUI 运行在深色背景的终端（如 iTerm2 深色主题）
- **THEN** 使用深色配色方案：深灰背景 + 浅灰/青色文字

#### Scenario: 浅色终端背景
- **WHEN** TUI 运行在浅色背景的终端（如 Terminal.app 亮色主题）
- **THEN** 使用浅色配色方案：浅灰背景 + 深灰/蓝色文字

#### Scenario: 配色一致性
- **WHEN** 用户启动 TUI 程序
- **THEN** 所有 UI 组件（列表项、选中项、提示文字）使用统一的配色方案

### Requirement: 关键颜色可读性保证
错误信息和成功信息的颜色 SHALL 在深色和浅色模式下均保持良好可读性。

#### Scenario: 错误提示可见
- **WHEN** TUI 显示错误信息
- **THEN** 错误文字使用红色，在深色和浅色背景下均清晰可见

#### Scenario: 成功提示可见
- **WHEN** TUI 显示成功信息
- **THEN** 成功文字使用绿色，在深色和浅色背景下均清晰可见
