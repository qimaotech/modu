## Why

当前只支持扁平结构的 feature 分支名（如 `feature-abc`），但 git 分支名可以包含 `/`（如 `feature/abc`）。用户希望能够使用 `feature/abc` 这样的分支名，同时避免在文件系统中创建深层嵌套目录。

## What Changes

- **目录名转换**：在文件系统中使用 `-` 替代 `/`（如 `feature/abc` → 目录名 `feature-abc`）
- **显示**：直接显示目录名（如 `feature-abc`）
- **新增转换函数**：`featureToDirName()`（仅单向转换）
- **兼容现有结构**：现有的扁平 features（如 `feature-abc`）仍然正常工作

## Capabilities

### New Capabilities
<!-- 无新增 spec 文件，主要是现有功能的增强 -->

### Modified Capabilities
- `engine`: CreateWorktree、ListWorktrees、DeleteWorktree、UpdateWorktree、GetWorktreeInfo、AddModule、RemoveModule 需要支持 feature 名的转换

## Impact

- 主要修改 `internal/engine/engine.go`：新增转换函数，修改各操作函数使用转换逻辑
- 其他模块（gitproxy, ui）无需改动
