## Context

当前 `modu init` 和配置向导 `config_wizard.go` 在保存 `.modu.yaml` 时，直接保存到当前工作目录。但根据用户需求，配置文件应保存到 `workspace` 目录，便于统一管理多项目配置。

同时，`modu init` 需要确保 workspace 和 worktree 目录存在且 workspace 是 git 仓库，首次使用时能自动完成初始化。

## Goals / Non-Goals

**Goals:**
- `.modu.yaml` 保存到 `workspace` 目录
- 自动创建 `workspace` 目录（如果不存在）
- 自动创建 `worktree-root` 目录（如果不存在）
- 如果 `workspace` 不是 git 仓库，自动执行 `git init` 并创建 `default-base` 分支

**Non-Goals:**
- 不修改其他命令的配置文件查找逻辑
- 不修改 `LoadConfig` 的解析逻辑
- 不自动执行 `git clone`（由 `modu init` 后续逻辑处理）

## Decisions

### Decision 1: 配置文件保存位置

**选择**：将 `.modu.yaml` 保存到用户输入的 `workspace` 目录，而非当前工作目录。

**理由**：
- 符合多项目集中管理的需求
- workspace 是用户已配置的目录路径，保存配置文件在其中是最自然的选择

**备选**：
- 保存到当前目录（当前行为）- 不符合用户期望
- 保存到单独配置目录 - 增加复杂度

### Decision 2: Git 仓库检查与初始化时机

**选择**：在 `doSaveConfig()` 中，检查 workspace 是否为 git 仓库，如果否则执行 `git init` + `git checkout -b <base>`。

**理由**：
- 配置向导流程统一，用户在一个流程中完成所有初始化
- 错误处理集中，便于向用户展示错误信息

**备选**：
- 在 `Engine.Init()` 中处理 - 职责分离更清晰，但增加复杂度

### Decision 3: Worktree 目录创建时机

**选择**：在 `doSaveConfig()` 中创建 worktree-root 目录。

**理由**：
- 与 workspace 初始化保持一致，都在保存配置前完成
- `SaveConfig` 函数已有创建 worktree-root 的逻辑，可复用

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|---------|
| `git init` 失败（如权限问题） | 返回明确错误信息，阻止保存配置 |
| `git checkout -b` 失败（如分支已存在） | 返回错误，提示用户分支可能已存在 |
| workspace 路径包含特殊字符 | `os.MkdirAll` 和 `git` 命令都能处理，暂无风险 |

## Open Questions

无
