## Context

`modu` 是一个多仓库管理工作 CLI，当前 `modu create` 命令在创建 feature worktree 后只输出成功信息。用户需要在 VSCode 中打开多个模块时，需要手动逐个添加或使用命令行打开，效率较低。

## Goals / Non-Goals

**Goals:**
- 在 `modu create` 成功后自动生成 `.code-workspace` 文件
- workspace 文件包含主项目（worktree 根目录）和所有配置的模块
- 生成的 workspace 文件可以直接双击在 VSCode 中打开

**Non-Goals:**
- 不修改 VSCode 以外的编辑器支持
- 不自动打开 VSCode（只生成文件）
- 不支持自定义 workspace 配置（用户可通过后续手动编辑扩展）

## Decisions

1. **在 engine 层实现 vs UI 层**
   - 理由：workspace 文件生成是 worktree 创建的固有部分，放在 engine 层更符合项目架构
   - 替代方案：在 CLI 层实现，需要传递更多参数，增加复杂性

2. **workspace 文件命名为 `{feature}.code-workspace`**
   - 理由：符合用户习惯，文件位于 feature 目录内

3. **使用 Go 标准库 encoding/json 生成 JSON**
   - 理由：无新增依赖，标准库已足够

## Risks / Trade-offs

- [Risk] 如果 feature 目录下存在非模块目录（如 .claude），workspace 会包含非必要目录
  - → Mitigation：只将配置中的 modules 添加到 workspace folders
- [Risk] workspace 文件可能与用户手动编辑的冲突
  - → Mitigation：每次 create 都会覆盖，用户可重新 create 生成最新配置
