## Context

`modu create <feature>` 在交互式终端模式下会调用 `ui.SelectModules` 让用户选择要创建的模块。当前预选逻辑只考虑「本地已存在的模块」，无法感知远端仓库是否有该分支。

当同事需要复现他人的 feature 环境时，必须知道哪些子模块实际包含该分支，然后手动勾选。用户期望只需告知分支名，系统自动感知哪些模块需要创建。

## Goals / Non-Goals

**Goals:**
- 在模块选择阶段自动查询远端分支，有该分支的模块默认预选
- 不改变现有交互流程，用户仍可手动调整选择
- 查询失败时 graceful degradation，不阻塞用户操作

**Non-Goals:**
- 不新增 CLI 命令，复用 `modu create`
- 不修改 `modu create` 的非交互模式行为
- 不实现后台预加载或缓存机制

## Decisions

### 1. 使用 `git ls-remote --heads` 查询远端分支

**选择**: 直接调用 `git ls-remote --heads <repo> refs/heads/<branch>`
**替代方案**:
- `git fetch` 后 `git branch -r --list */<branch>`: 需要 fetch 到本地，更重
- API 调用 (GitHub/GitLab): 需要认证，复杂度高

**理由**: 轻量、免认证、覆盖所有 Git 仓库实现

### 2. Engine 层实现并发查询

**选择**: 在 `Engine.GetModulesWithRemoteBranch` 中并发调用 `GitProxy.RemoteBranchExists`
**替代方案**:
- UI 层直接调用: 破坏分层，UI 不应直接操作 Git
- Engine 缓存结果: 一次查询多次使用，但增加复杂度，YAGNI

**理由**: 保持分层，并发开销小，errgroup 统一错误处理

### 3. 预选逻辑: 本地已有 OR 远端有该分支

**选择**: `existingMap[m.Name] || remoteHasBranch[m.Name]`
**替代方案**:
- 只有远端有才预选: 已有模块反而需要用户重新选中，体验差
- 只有本地有才预选: 失去增强意义

**理由**: 最大化用户体验，已有模块保持选中，新增模块自动感知

### 4. 查询失败时跳过预选，不阻塞用户

**选择**: 查询远端分支失败时，`remoteHasBranch` 返回空 map，走默认预选逻辑
**理由**: 网络问题不应阻塞核心功能，用户仍可正常创建

## Risks / Trade-offs

- **[Risk] 网络延迟**: 每次 `create` 都需要查询远端，网络慢时增加等待时间
  - **Mitigation**: 并发查询 + 显示 "正在查询远端分支..." 提示；用户可接受现有 create 的等待时间

- **[Risk] 部分模块查询失败**: 某些仓库网络不通，导致预选不完整
  - **Mitigation**: graceful degradation，失败的模块视为无该分支，用户可手动选中

- **[Risk] 远端分支存在但已落后**: `git ls-remote` 只检查分支是否存在，不保证最新
  - **Mitigation**: 这是预期行为，用户创建时仍会基于最新代码，落后问题由 `create` 时的 rebase 处理

## Open Questions

无
