## Context

当前 `ListWorktrees` 函数在 `internal/engine/engine.go` 中只扫描 `worktreeRoot` 的直接子目录，无法识别带 `/` 的 feature 名（如 `feature/abc`）。

用户希望能够使用 `feature/abc` 这样的分支名，同时保持文件系统的目录结构扁平。

## Goals / Non-Goals

**Goals:**
- 支持带 `/` 的分支名（如 `feature/abc`, `feature/a/b`）
- 文件系统目录扁平化，避免深层嵌套
- 兼容现有的扁平结构 features（如 `feature-abc`）

**Non-Goals:**
- 不修改 git worktree 的底层行为
- 不添加新的配置文件或 API

## Decisions

### 1. 目录名转换策略

**方案**: 使用 `-` 替代 `/` 存储目录名（单向转换）

- 用户输入：`feature/hello`
- 目录名：`feature-hello`
- 显示：直接显示目录名 `feature-hello`

**实现**:
```go
func featureToDirName(feature string) string {
    return strings.ReplaceAll(feature, "/", "-")
}
```

### 2. 修改的函数

需要在以下函数中使用转换逻辑：
- `CreateWorktree`：创建目录时使用 `featureToDirName`
- `ListWorktrees`：列出时使用 `dirNameToFeature`
- `DeleteWorktree`：删除时使用 `featureToDirName`
- `UpdateWorktree`：更新时使用 `featureToDirName`
- `GetWorktreeInfo`：获取详情时使用 `featureToDirName`
- `AddModule`：添加模块时使用 `featureToDirName`
- `RemoveModule`：删除模块时使用 `featureToDirName`

### 3. 兼容现有结构

- 现有目录 `feature-abc` → 显示为 `feature/abc`（自动转换）
- 用户创建新的 `feature/abc` → 创建目录 `feature-abc`

这样无需迁移现有数据，自然兼容。

## Risks / Trade-offs

- **冲突风险**: `feature-abc` 和 `feature_abc` 可能混淆
  - **缓解**: 只转换 `/` 为 `-`，不处理其他字符，用户可自行区分
- **转换不可逆**: `-` 在分支名中罕见，影响较小
