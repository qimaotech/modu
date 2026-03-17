## Context

当前 `modu list` 命令只显示 feature worktrees 信息，用户无法查看主项目（workspace）及其模块的分支状态。这是一个简单的 CLI 功能增强。

## Goals / Non-Goals

**Goals:**
- 给 `modu list` 添加 `-a` flag 显示主项目及其模块分支
- Workspace 信息显示在 Features 列表上方

**Non-Goals:**
- 不修改 TUI 模式（TUI 已有此功能）
- 不添加新的 API 或服务

## Decisions

1. **Engine 层**: 复用 `GetMainProject` 方法获取主项目状态，新增 `GetMainProjectModules` 方法获取主项目下各模块的状态

2. **Output 层**: 在 `Formatter` 中新增 `FormatMainProjectResponse` 方法，文本格式输出，JSON 格式返回结构化数据

3. **CLI 层**: 在 `list` 命令添加 `-a` flag，调用 engine 获取主项目模块信息并传递给 output

## Risks / Trade-offs

- 风险：主项目路径不存在时应该优雅处理 → 方案：`-a` flag 下如果主项目不存在则不显示，不报错
