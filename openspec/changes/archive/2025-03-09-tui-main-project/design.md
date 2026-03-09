## Context

当前 modu TUI (`internal/ui/ui.go`) 的列表视图只显示 feature 分支，每个 feature 展示其 dirty 状态。用户无法：
1. 在 TUI 中直接看到主项目（workspace 目录下的主仓库）的状态
2. 通过 TUI 打开主项目的 VS Code
3. 一键更新主项目和所有模块的代码

## Goals / Non-Goals

**Goals:**
- 在 TUI 列表顶部固定显示主项目（显示名称、dirty 状态、分支）
- 主项目使用简化菜单（仅打开 VS Code 和更新代码）
- Feature 使用现有菜单（打开 VS Code、模块管理、删除）
- 实现 `u` 快捷键：一键更新主项目 + 所有 modules（git fetch + rebase）

**Non-Goals:**
- 不修改 CLI 命令的行为
- 不添加其他 IDE 支持（仅 VS Code）
- 不支持选择性更新（要么全更，要么不更）

## Decisions

**1. 数据结构：新增 `MainProjectEntry` 类型**

```go
type ListEntry interface {
    IsMainProject() bool
    GetName() string
    GetDirtyCount() int
}
```

在 `App` 结构体中增加 `mainProject *MainProjectStatus` 字段。

**2. 列表渲染顺序**

```
1. 主项目（固定第一行，标记 "main"）
2. Feature 列表（按现有顺序）
```

**3. 菜单逻辑分支**

根据当前选中项的类型（主项目 vs feature）渲染不同菜单：
- 主项目菜单：`["打开 VS Code (o)", "更新代码 (u)"]`
- Feature 菜单：`["打开 VS Code (o)", "Modules 管理 (m)", "删除 (d)"]`

**4. Engine 层新增方法**

```go
// UpdateMainProject 更新主项目和所有模块的代码
func (e *Engine) UpdateMainProject(ctx context.Context) error
```

实现逻辑：
1. 并发 fetch + rebase 主项目仓库
2. 并发 fetch + rebase 每个 module 仓库（使用当前分支）
3. 汇总成功/失败结果

**5. 快捷键映射**

| 快捷键 | 主项目 | Feature |
|--------|--------|---------|
| `o` | 打开 VS Code | 打开 VS Code |
| `u` | 更新代码 | 无效 |
| `m` | 无效 | 模块管理 |
| `d` | 无效 | 删除确认 |
| `q` | 退出 | 退出/返回 |

## Risks / Trade-offs

- **[风险]** 如果某个模块 rebase 失败，整个更新是否回滚？
  - **[ mitigation ]** 不回滚，已完成的 rebase 保持成功状态，报告失败的模块

- **[风险]** 更新过程中网络超时
  - **[ mitigation ]** 使用 context 超时，每个模块独立超时

- **[风险]** 主项目不存在或不是 git 仓库
  - **[ mitigation ]** 启动时检查配置，如果 workspace 无效则不显示主项目条目
