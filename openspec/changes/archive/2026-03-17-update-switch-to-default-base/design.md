## Context

当前 `modu update` 命令更新 workspace 主项目时，只是执行 `git fetch` + `git rebase` 到当前分支跟踪的远程分支，没有切换到 `default-base` 配置的分支（如 develop）。这导致 workspace 主项目无法保持最新状态，不利于需求分析。

## Goals / Non-Goals

**Goals:**
- 更新主项目时，自动切换到 `default-base` 配置的分支（如 develop）
- 保持模块更新行为不变（仍只 fetch + rebase 到当前分支跟踪的远程分支）

**Non-Goals:**
- 不修改其他命令（如 create、delete）的行为
- 不修改 TUI 界面的行为（仅 CLI 子命令）

## Decisions

1. **在 GitProxy 新增 `FetchAndSwitchBranch` 方法**
   - 原因：封装 fetch + checkout + rebase 的通用逻辑，供 Engine 调用
   - 备选方案：直接在 Engine 中调用现有方法组合，但会导致重复逻辑

2. **Engine.UpdateMainProject 使用新方法**
   - 原因：主项目和模块更新时都需要切换分支
   - 模块使用各自的 base-branch（如果有配置）或全局 default-base
   - 备选方案：模块保持原有 Rebase 行为，但用户期望模块也切换分支

## Risks / Trade-offs

- **风险**：切换分支可能失败（如远程分支不存在）
  -  mitigation：方法内部处理错误，返回明确的错误信息
- **风险**：切换分支可能导致本地修改丢失
  -  mitigation：方法内部先检查 dirty 状态，必要时拒绝切换
