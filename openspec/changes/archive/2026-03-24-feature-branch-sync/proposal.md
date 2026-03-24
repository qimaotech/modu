## Why

当同事需要接手或 review 别人的 feature 时，只需知道分支名即可快速复现环境。目前 `modu create` 无法感知哪些子模块实际拥有该分支，用户需要手动选择，效率低且容易遗漏。

## What Changes

- 新增 `GitProxy.RemoteBranchExists` 方法，通过 `git ls-remote --heads` 查询远端分支是否存在
- 新增 `Engine.GetModulesWithRemoteBranch` 方法，并发查询所有子模块的远端分支状态
- `ui.SelectModules` 增加 `remoteHasBranch` 参数，预选逻辑加入远端分支判断
- `modu create` 在模块选择阶段自动预选远端已有该分支的模块

## Capabilities

### New Capabilities
- `remote-branch-query`: 查询子模块远端是否包含指定分支，返回布尔值供预选逻辑使用

### Modified Capabilities
- `cli`: `modu create` 的模块选择交互增加远端分支感知预选行为（需求层面无变化，实现层面增强）

## Impact

- **新增代码**: `internal/gitproxy` (RemoteBranchExists), `internal/engine` (GetModulesWithRemoteBranch)
- **修改代码**: `internal/ui` (SelectModules 签名和预选逻辑), `cmd/modu/main.go` (create 命令调用逻辑)
- **性能影响**: create 时需额外并发查询远端（网络 IO），限制并发数同 Config.Concurrency
- **向后兼容**: 纯增强，不改变已有行为，查询失败时 graceful degradation
