## Why

TUI 操作菜单目前把可打开的工具写死为 VS Code 和 Codex，用户无法按项目配置扩展到 Zed、Cursor 等常用编辑器。将打开工具改为配置驱动，可以让团队按自己的开发工具偏好扩展菜单，同时避免未安装工具或冲突快捷键破坏交互。

## What Changes

- 在现有 `.modu.yaml` 中增加可选的自定义 app opener 配置，描述展示名称、macOS app 名称、可选快捷键等信息。
- TUI 启动时读取配置中的 app opener 列表，并检测对应 app 是否已安装。
- 操作菜单只展示已安装的自定义 app opener，未安装的工具不展示。
- 当自定义 app opener 的快捷键与内置菜单项或其他可见 opener 冲突时，菜单仍展示该 opener，但不展示也不绑定冲突快捷键。
- 内置打开能力保持可用，支持以 Zed / `z` 作为配置示例。

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `config`: 配置文件结构新增可选 app opener 字段，并定义加载、校验与兼容规则。
- `tui`: 操作菜单根据配置、安装检测和快捷键冲突结果动态展示打开工具项。

## Impact

- `internal/config`: 配置结构体、YAML 解析、校验与测试。
- `internal/ui`: 操作菜单构建、渲染、快捷键处理、打开 app 命令与测试。
- `README.md`: 配置示例和 TUI 快捷键说明需要补充。
- 无新增依赖；安装检测优先使用系统已有能力。
