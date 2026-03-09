## Context

当前 TUI 应用（modu）的操作菜单只支持「打开 VS Code」和「删除」两个功能。用户需要在 TUI 中动态管理 feature 下的模块（添加/删除模块的 worktree），而无需使用 CLI 命令。

**当前状态：**
- 列表视图：显示所有 feature，按 Enter 进入操作菜单，按 `d` 删除，按 `o` 打开 VS Code
- 操作菜单：两个选项「打开 VS Code」「删除」
- 现有 `ModuleSelector` 组件已支持模块选择 UI（用于 `modu create` 命令）

**约束：**
- 使用 Bubble Tea 框架（TUI）
- 保持与现有 `ModuleSelector` 组件的交互一致性
- 复用 Engine 层的 worktree 操作能力

## Goals / Non-Goals

**Goals:**
- 在列表视图按 `m` 直接进入 Modules 管理（快速入口）
- 在操作菜单中增加「Modules 管理」选项（第二位，删除移到最后）
- Modules 管理视图显示所有配置模块，已创建的标记 `[x]`，未创建的标记 `[ ]`
- 空格键切换选中状态，回车确认执行增删

**Non-Goals:**
- 修改模块配置本身（增删改配置中的模块定义）
- 批量创建/删除多个 feature 的模块

## Decisions

### 1. 单层设计 vs 子菜单设计

**决策：单层设计**

Modules 管理视图直接显示所有模块，用户通过空格切换选择。与 `modu create` 的交互一致，上手成本低。

**备选：子菜单设计**（先显示「添加/删除」再选模块）—— 交互更繁琐，排除。

### 2. 操作菜单顺序

**决策：打开 VS Code → Modules 管理 → 删除**

- 打开 VS Code：最高频
- Modules 管理：次高频
- 删除：低频且危险，放最后

### 3. Engine 层单模块操作复用

**决策：复用现有 `CreateWorktree` 和 `DeleteWorktree` 的逻辑**

- `AddModule`：复用 `CreateWorktree` 中单模块创建逻辑，但不创建主项目 worktree
- `RemoveModule`：复用 `DeleteWorktree` 中单模块删除逻辑，但不删除整个 feature 目录

**备选：完全新建独立方法** —— 代码重复，排除。

## Risks / Trade-offs

| 风险 |  Mitigation |
|------|-------------|
| 模块增删过程中用户中断 | 记录已成功的操作，部分成功时提示用户 |
| 并发创建多个模块失败 | 复用 CreateWorktree 的错误处理和回滚逻辑 |
| 脏检查阻止删除 | 删除模块前检查脏状态，提示用户 |

## Migration Plan

1. 部署新版本 TUI，用户可通过 `m` 键或操作菜单进入 Modules 管理
2. 无数据迁移需求
3. 回滚：部署旧版本即可，旧版本不支持 Modules 管理但不影响其他功能
