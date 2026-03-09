## Context

feature 目录下除配置中的模块外，还有 `.claude`、`openspec` 等非模块目录（多为 A：主项目内的普通目录；少数为 B：独立 git 仓库但未写入 modu 配置）。原实现用 `ReadDir` 把所有子目录都当模块处理，导致 list 展示噪音、create 时误删非模块目录。

## Goals / Non-Goals

**Goals:**
- `modu list` 只展示配置内模块。
- create 增删模块时只增删配置内目录，不动 `.claude`、`openspec` 等。
- DeleteFeature 的脏检查与删除仅针对配置内模块。

**Non-Goals:**
- 不改变配置格式或 CLI 参数。
- 不自动识别「未配置但为 worktree」的目录并展示。

## Decisions

1. **Engine**：新增 `configuredModuleNames() map[string]bool`，在 ListEnvs、GetWorktreeInfo、DeleteFeature 的脏检查与删除循环中，仅处理名称在此集合内的子目录。
2. **CLI create**：统计 `existingModules` 时只把「在 `Config.Modules` 中且存在的子目录」加入，避免非配置目录进入 `modulesToDelete`。
3. **Spec**：在 `openspec/specs/engine/spec.md` 与 `openspec/specs/domain/spec.md` 中明确「模块认定：仅配置内目录」。

## Risks / Trade-offs

- 未写入配置的独立 worktree 目录不会在 list 中显示，也不会被 create 的增删逻辑操作（保持不动），符合预期。
