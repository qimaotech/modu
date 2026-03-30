## Context

`modu info <feature>` 当前要求必传 feature 参数。实现位于 `cmd/modu/main.go` 的 `runInfo` 函数（第 433-446 行），通过 `featureToDirName` 将 feature 名转换为目录名后拼接 `worktreeRoot` 路径查询。

工作目录结构：

```
worktreeRoot/
└── feature-abc/        ← feature 目录
      ├── mainproject  ← 主项目 worktree（无子目录）
      └── module1/     ← 模块 worktree
```

用户期望在 `feature-abc/` 或其任意子目录执行 `modu info` 时，自动识别当前所属 feature 并展示其详情。

## Goals / Non-Goals

**Goals:**
- `modu info` 无参数时，从当前工作目录向上回溯，自动推断所属 feature 并展示详情。
- 保持 `modu info <feature>` 显式传参行为不变。
- 合理的错误提示。

**Non-Goals:**
- 不改变 `GetWorktreeInfo` engine 函数的签名或行为。
- 不引入新的公共 API。
- 不处理跨 worktreeRoot 的复杂场景。

## Decisions

### 1. 目录推断逻辑

**方案**：在 `runInfo` 中内联实现，不新增 engine 方法。

```go
func runInfo(cmd *cobra.Command, args []string) {
    var feature string
    if len(args) == 0 {
        // 无参数：从当前目录向上回溯推断 feature
        cwd, err := os.Getwd()
        if err != nil {
            fmt.Fprintln(os.Stderr, "获取当前目录失败")
            os.Exit(1)
        }
        feature = inferCurrentFeature(cwd, eng.Config.WorktreeRoot)
        if feature == "" {
            fmt.Fprintln(os.Stderr, "当前目录不在任何 feature 下")
            os.Exit(1)
        }
    } else {
        feature = args[0]
    }
    // ... 后续逻辑不变
}
```

`inferCurrentFeature` 伪实现：

```go
func inferCurrentFeature(cwd, worktreeRoot string) string {
    // 规范化路径
    worktreeRoot, _ = filepath.EvalSymlinks(worktreeRoot)
    for {
        parent := filepath.Dir(cwd)
        if parent == worktreeRoot || parent == "." {
            // 找到 worktreeRoot，返回当前目录名
            return filepath.Base(cwd)
        }
        if parent == "/" || parent == "." {
            break
        }
        cwd = parent
    }
    return "" // 未找到
}
```

**理由**：
- 内联实现简单直接，避免引入新方法污染 engine 公共接口。
- 循环向上遍历而非字符串比较，兼容 symlink 场景。

### 2. CLI 参数约束调整

```go
// Before
Args: cobra.ExactArgs(1)

// After
Args: cobra.MaximumNArgs(1)
```

**理由**：`cobra.MaximumNArgs(1)` 允许 0 或 1 个参数，与推断逻辑配合实现无参数场景。

### 3. 边界处理

| 场景 | 行为 |
|------|------|
| 当前目录不在 `worktreeRoot` 下 | 报错：当前目录不在任何 feature 下 |
| 当前目录恰好是 `worktreeRoot` 本身 | 报错：请指定 feature |
| 推断出的 feature 不存在 | 复用现有 `GetWorktreeInfo` 的错误处理 |

**理由**：明确的错误提示比静默失败更友好。

## Risks / Trade-offs

[Risk] 用户在 `worktreeRoot` 根目录执行 `modu info` 时无法推断 → [Mitigation] 明确报错提示需指定 feature 参数。

[Risk] 目录名可能与真实 git 分支名不一致 → [Mitigation] 按设计文档说明，modu 系统内统一使用目录名作为 feature 标识，info/list/tui 均如此。

## Open Questions

无。
