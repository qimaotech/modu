## Why

当前 `modu init` 执行时，`.modu.yaml` 配置文件保存到当前工作目录。但用户期望配置文件统一存放在 `workspace` 目录下，便于集中管理。此外，`modu init` 不会自动初始化 workspace 目录为 git 仓库，也不创建默认分支，导致首次使用需要手动执行 `git init` 和创建分支。

## What Changes

1. **配置文件保存位置变更**：`modu init` 和配置向导创建的 `.modu.yaml` 保存到 `workspace` 目录，而非当前工作目录。

2. **workspace 目录初始化**：
   - 如果 `workspace` 目录不存在，自动创建
   - 如果 `workspace` 不是 git 仓库，自动执行 `git init` 并 `git checkout -b <default-base>`（如 `develop`）

3. **worktree 目录初始化**：
   - 如果 `worktree-root` 目录不存在，自动创建

## Capabilities

### New Capabilities

- `workspace-init`: 扩展配置向导和 init 命令，在保存配置文件前检查并初始化 workspace 目录为 git 仓库，同时创建默认分支

## Impact

- **代码改动**：`internal/ui/config_wizard.go` 的 `doSaveConfig()` 方法
- **新函数**：新增 `ensureGitRepo()` 辅助函数，检查并初始化 git 仓库
- **行为变更**：配置文件保存位置从当前目录改为 workspace 目录
