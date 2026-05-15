# modu 输出层规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md + 代码

## Purpose

定义 CLI 文本输出与 JSON 结构化输出规则，确保人类可读提示和机器可解析响应在成功、失败及删除备份场景中保持稳定。

## 输出格式

- **text**（默认）：人类可读表格或列表，如 list/status/info 的表格展示。
- **json**：符合 [errors 规范](../errors/spec.md) 中的机器输出协议；成功时含 `success`、`action`、`results`，失败时含 `code`、`message`、`data`。

## 职责

- 接收 engine/core 的数据结构（如 `[]WorktreeEnv`、`ModuleStatus`），转换为表格行或 JSON 字段。
- 不执行 Git 或业务逻辑，仅做展示与序列化。

## 用户可见文案（中文）

text 格式下，面向用户的成功/失败提示 SHALL 使用中文。删除响应文案如下（来源：docs/plans/2026-03-09-delete-prompts-localization.md）：

| 场景 | 文案 |
|------|------|
| 删除成功 | ✓ 已删除 feature: %s |
| 删除失败 | ✗ 删除 feature 失败: %s |
| 错误明细行 | 错误: %s |

JSON 格式中 `action` 等字段保持英文（如 `"delete"`），仅 text 输出本地化。
## Requirements
### Requirement: Delete response includes backup path
Delete responses SHALL include the backup archive path produced by the engine when deletion succeeds. Text output SHALL display the backup path in Chinese user-facing copy, and JSON output SHALL include a `backupPath` field.

#### Scenario: Text delete response
- **WHEN** formatting a successful delete response with backup path `/worktrees/.modu/backups/20260515-153012_my-feature.tar.gz`
- **THEN** text output includes `备份文件: /worktrees/.modu/backups/20260515-153012_my-feature.tar.gz`

#### Scenario: JSON delete response
- **WHEN** formatting a successful delete response with backup path `/worktrees/.modu/backups/20260515-153012_my-feature.tar.gz`
- **THEN** JSON output includes `"backupPath": "/worktrees/.modu/backups/20260515-153012_my-feature.tar.gz"`

## 与代码的对应

- 实现：`internal/output`（Table 渲染、JSON 序列化）。
