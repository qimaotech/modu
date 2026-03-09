# modu 领域模型规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## 目的

定义与 Git worktree 工作流相关的核心领域对象，不依赖 Git 或 UI 实现。

## 领域对象

### WorktreeEnv

表示一个 feature 环境，包含该环境下各模块的工作树状态。**Modules 仅包含配置内模块**（`Config.Modules` 中列出的目录），feature 下的其他子目录（如 `.claude`、`openspec`）不在此列表。

| 字段 | 类型 | 说明 |
|------|------|------|
| Name | string | Feature 名称 |
| Base | string | 基准分支 |
| MainProject | *ModuleStatus | 主项目状态（可选） |
| Modules | []ModuleStatus | 各配置内模块在该环境下的状态 |

### ModuleStatus

记录单个模块在工作树下的状态。

| 字段 | 类型 | 说明 |
|------|------|------|
| Name | string | 模块名 |
| Path | string | 物理路径 |
| IsDirty | bool | 是否存在未提交修改 |
| Branch | string | 当前分支 |
| Error | error | 该模块操作失败原因（可选） |

## 使用约定

- Engine 通过 gitproxy 获取状态后组装为 `WorktreeEnv`/`ModuleStatus`。
- 输出层（Table/JSON）仅依赖这些结构，不解析 Git 原始输出。

## 与代码的对应

- 实现：`internal/core/domain.go`（WorktreeEnv、ModuleStatus）。
