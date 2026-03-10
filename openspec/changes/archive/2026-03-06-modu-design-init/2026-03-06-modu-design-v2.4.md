# modu 技术实现文档 (v2.4 - 整合最终版)

## 1. 概述与背景

### 1.1 项目背景

**modu** 是一个基于 Go 语言开发的多模块 Git Worktree 管理工具。它旨在替代复杂的 Shell 脚本，通过 TUI 和强类型 CLI 提供跨平台一致、安全、高效的工作流管理。

**解决的问题：**
- 团队成员多数无 Shell 经验
- Mac/Linux 行为可能不一致
- 分发给团队需安装 task runner

### 1.2 目标用户

- 团队成员（主要）
- 大模型/脚本调用（次要，通过 `-o json`）

### 1.3 环境约束

- **Go Version**: 1.21+ (利用 `errors.Join` 处理并发多错)
- **Git Version**: 2.25+ (支持 `worktree` 核心功能)
- **OS**: 完全兼容 Linux, macOS (Darwin)。Windows 仅支持 WSL2 环境

---

## 2. 架构设计 (Architecture)

### 2.1 分层架构与依赖关系

采用典型的 Go Clean Architecture 简化版，确保核心逻辑（Domain）不依赖于具体实现（Git/TUI）。

```
┌─────────────────────────────────────────────────────────┐
│ cmd/modu                                               │
│ 入口层：Cobra 路由、全局 Flag 解析、环境初始化          │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│ internal/core                                          │
│ 领域对象：Repo, Worktree, WorktreeEnv, ModuleStatus     │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│ internal/engine                                        │
│ 核心控制器：并发调度、事务编排（Create 失败回滚）、脏检查 │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│ internal/gitproxy                                       │
│ Git 原语封装：屏蔽 OS 执行细节，命令调用及 stderr 解析   │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│ internal/ui                                            │
│ 基于 Bubble Tea 的状态机，仅负责 View 渲染               │
└─────────────────────────────────────────────────────────┘
```

### 2.2 项目目录结构

```
.
├── cmd/modu/main.go          # Cobra 路由：处理 CLI 参数与 TUI 切换
├── internal/
│   ├── core/                 # 领域对象定义（Repo, Worktree, Status）
│   ├── config/               # 配置文件加载、校验、查找逻辑
│   ├── gitproxy/             # Git 命令封装 (Worktree, Status, Clone)
│   ├── engine/               # 核心业务：并发调度、脏检查逻辑
│   ├── ui/                   # Bubble Tea TUI 界面实现
│   └── output/               # 结构化输出 (Table/JSON)
├── modu.yaml                 # 示例配置
└── go.mod
```

---

## 3. 核心模型定义 (Domain Models)

### 3.1 配置文件结构

```go
type Config struct {
    Workspace    string          `yaml:"workspace"`      // 裸仓库/主仓库所在目录
    WorktreeRoot string         `yaml:"worktree-root"`  // 特性分支代码存放目录
    DefaultBase  string         `yaml:"default-base"`    // 默认基准分支 (如 develop)
    Concurrency  int            `yaml:"concurrency"`     // 并发数，默认 5
    AutoFetch    bool           `yaml:"auto-fetch"`      // 操作前自动 fetch
    StrictDirty  bool           `yaml:"strict-dirty-check"` // 删除前强制脏检查
    Modules      []Module       `yaml:"modules"`
}

type Module struct {
    Name       string `yaml:"name"`
    URL        string `yaml:"url"`
    BaseBranch string `yaml:"base-branch,omitempty"` // 可选，覆盖全局设置
}
```

### 3.2 核心领域模型

```go
// WorktreeEnv 表示一个 feature 环境，包含多个模块的工作树
type WorktreeEnv struct {
    Name    string           // Feature 名称
    Base    string           // 基准分支
    Modules []ModuleStatus   // 各模块在该环境下的状态
}

// ModuleStatus 记录单个模块的工作树状态
type ModuleStatus struct {
    Name    string           // 模块名
    Path    string           // 物理路径
    IsDirty bool             // 是否存在未提交修改
    Branch  string           // 当前分支
    Error   error            // 记录该模块操作失败的具体原因
}
```

### 3.3 modu.yaml 配置示例

```yaml
version: "2.4"
workspace: /opt/case/commerce/workspace
worktree-root: /opt/case/commerce/worktrees
default-base: develop
concurrency: 8
auto-fetch: true
strict-dirty-check: true

modules:
  source: config
  repos:
    - name: pixiu-ad-server
      url: git@github.com:commerce/pixiu-ad-server.git
```

---

## 4. 核心命令规范

| 命令 | 参数 | 说明 |
|------|------|------|
| `modu` | - | 无子命令时进入 TUI |
| `modu scan` | `--export` | 扫描 workspace，显示模块列表，可导出 YAML |
| `modu init` | `--parallel` | 并发克隆所有配置中的仓库 |
| `modu create` | `<feature> [--base <branch>]` | 并发创建基于基准分支的 worktree |
| `modu list` | - | 列出所有 worktree |
| `modu info` | `<feature>` | 查看 worktree 详情 |
| `modu delete` | `<feature> [-f, --force]` | 删除 worktree（前置脏检查） |
| `modu status` | - | 批量展示所有模块 Dirty 状态 |

---

## 5. 核心引擎流程 (Engine Logic)

### 5.1 事务性并发创建 (Atomic Create)

为避免多仓库环境下出现"半完成"状态，`create` 需具备简单的回滚逻辑。

1. **Pre-check**: 检查 `worktree-root/<feature>` 是否已存在
2. **Execution**: 启动 `errgroup` 并发执行 `git worktree add`
3. **Rollback**: 若任一模块失败，记录错误并并发执行 `rm -rf` 已创建的目录，确保环境洁净

```go
func (e *Engine) CreateWorktree(ctx context.Context, feature, base string) error {
    // 1. Pre-check
    if exists(e.Config.WorktreeRoot + "/" + feature) {
        return ErrFeatureExists
    }

    // 2. Execution
    var created []string
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(e.Config.Concurrency)

    for _, module := range e.Config.Modules {
        module := module
        g.Go(func() error {
            // git fetch + git worktree add
            path := fmt.Sprintf("%s/%s/%s", e.Config.WorktreeRoot, feature, module.Name)
            err := e.GitProxy.CreateWorktree(ctx, module.Name, feature, base, path)
            if err == nil {
                created = append(created, path)
            }
            return err
        })
    }

    // 3. Rollback on failure
    if err := g.Wait(); err != nil {
        e.rollback(created)
        return err
    }
    return nil
}
```

### 5.2 并发引擎 (Concurrent Engine)

使用 `golang.org/x/sync/errgroup` 实现限流并发：

```go
func (e *Engine) RunParallel(ctx context.Context, tasks []Task) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(e.Config.Concurrency)

    for _, task := range tasks {
        task := task
        g.Go(func() error {
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                return task.Execute(ctx)
            }
        })
    }
    return g.Wait()
}
```

### 5.3 脏检查算法 (Dirty Check Algorithm)

```go
// 遍历所有模块，检查是否存在未提交修改
func (e *Engine) CheckDirty(env WorktreeEnv) ([]ModuleStatus, error) {
    var dirty []ModuleStatus

    for _, module := range env.Modules {
        status, err := e.GitProxy.GetStatus(ctx, module.Path)
        if err != nil {
            return nil, err
        }

        if status.IsDirty {
            dirty = append(dirty, status)
        }
    }

    return dirty, nil
}
```

**逻辑伪代码：**
```
for _, module := range environment.Modules:
    # 执行 git status --porcelain
    # 检查结果：
    # 1. 如果有输出 -> Dirty
    # 2. 如果无输出 -> Clean
    # 3. 如果目录不存在 -> Missing
```

### 5.4 命令与 Git 原语映射

| modu 命令 | 内部逻辑 (Git Primitives) |
| --- | --- |
| `modu init` | `git clone <url> <workspace>/<name>` |
| `modu create <f>` | 1. `git -C <ws>/<name> fetch`<br>2. `git -C <ws>/<name> worktree add <wt-root>/<f>/<name> -b <f> <base>` |
| `modu list` | 1. 扫描 `<wt-root>` 目录<br>2. `git worktree list` |
| `modu delete <f>` | 1. **Dirty Check**<br>2. `git worktree remove <path>`<br>3. `rm -rf <wt-root>/<f>` |

---

## 6. 错误处理体系 (Error Handling)

### 6.1 错误分级与代码

引入结构化错误代码，便于 CLI 和 JSON 输出统一解析：

| 错误码 | 说明 |
|--------|------|
| `ERR_CONFIG_INVALID` | `modu.yaml` 格式或路径非法 |
| `ERR_GIT_EXEC` | Git 命令执行失败（需包含 `exit code` 和 `stderr`） |
| `ERR_DIRTY_WORKTREE` | 脏检查拦截 |
| `ERR_PARTIAL_FAILURE` | 并发操作中部分成功，部分失败 |
| `ERR_FEATURE_EXISTS` | Feature 目录已存在 |
| `ERR_FEATURE_NOT_FOUND` | Feature 目录不存在 |

### 6.2 错误链条包装

所有 `gitproxy` 层抛出的错误必须包含上下文：

```go
fmt.Errorf("[%s] git worktree add failed: %w", moduleName, err)
```

### 6.3 错误日志规范

**错误日志要很详细**，便于排查：

- **上下文**：当前命令与参数、使用的配置文件路径、workspace/worktree-root、涉及模块名
- **链式原因**：多步操作（如 init → clone）要保留每一步的失败原因，不截断
- **外部命令**：Git 等调用的完整 stderr/stdout 写入日志
- **路径与状态**：失败时的绝对路径、目标分支、当前分支等

### 6.4 并发错误聚合

当并发执行失败时，`modu` 不应立即崩溃，而是：

1. **收集错误**：使用 `errors.Join` 等待当前批次任务结束
2. **汇总报告**：在终端打印"成功 X 个，失败 Y 个"，并列出失败模块的具体 Git 报错
3. **JSON 响应**：`-o json` 模式下，`errors` 数组将包含所有失败协程返回的上下文

### 6.5 机器输出协议 (`-o json`)

**成功响应：**

```json
{
  "success": true,
  "action": "create",
  "feature": "feature-login",
  "results": [
    { "module": "auth-svc", "status": "success", "path": "/worktrees/feature-login/auth-svc" }
  ],
  "errors": []
}
```

**失败响应：**

```json
{
  "code": "ERR_DIRTY_WORKTREE",
  "message": "cannot delete: uncommitted changes detected",
  "data": {
    "feature": "feat-login",
    "dirty_modules": [
      { "name": "auth-api", "files": ["main.go", "config.yaml"] }
    ]
  }
}
```

---

## 7. TUI 状态机设计 (Bubble Tea)

### 7.1 入口与状态

- **入口**：裸命令 `modu` 进入 TUI；带子命令则走 CLI
- **能力**：只读（worktree 列表、分支、模块）+ 创建/删除 worktree；删除前必须确认

### 7.2 状态机定义

1. **LoadingState**: 并发执行 `init` 或 `create` 时，显示每一个 Module 的当前进度（如：`api-server: Cloning...`）
2. **ListState**: 展示所有 `feature` 列表，光标选中时展示该环境下各模块的 `Branch` 和 `Status (Clean/Dirty)`
3. **ConfirmState**: 删除前的二次确认
4. **ErrorState**: 操作失败时显示错误详情，允许重试

### 7.3 UI 表现

- 多行并行进度条（Multi-Spinner），实时显示每个模块的任务状态
- 支持键盘快捷键：上下选择、回车确认、ESC 取消

---

## 8. 测试覆盖 (Testing Strategy)

### 8.1 覆盖率目标

- **核心引擎 (internal/engine)**: 100% 逻辑覆盖（使用 Mock Git）
- **整体项目**: 强制 **> 85%**
- 新增代码原则上需同步补测，覆盖率不降低

### 8.2 关键测试用例矩阵

| 模块 | 测试场景 | 预期行为 |
| --- | --- | --- |
| **Config** | 缺失 `workspace` 字段 | 报错 `ERR_CONFIG_INVALID` |
| **Engine** | 并发创建时某一个模块失败 | 触发回滚，删除其他已创建的 worktree 目录 |
| **Engine** | 在 Dirty 目录下执行 delete | 拦截操作，返回该模块名 |
| **GitProxy** | 解析 `git status` 输出 | 正确识别 `M`, `??`, `D` 状态为 Dirty |
| **E2E** | 模拟完整 `init` -> `create` -> `delete` | 物理目录结构与 `git worktree list` 一致 |

### 8.3 脏检查测试

- 在临时目录修改文件但不提交，断言 `modu delete` 返回 `ERR_DIRTY_WORKTREE`
- 添加新文件（Untracked），验证脏检查识别
- 使用 `--force` 参数，验证跳过检查

### 8.4 并发竞争测试

- 模拟多个模块同时向同一个 `worktree-root` 写入数据，验证路径隔离性
- 测试并发数为 1 时的行为与串行一致

### 8.5 CLI 端到端测试

每个子命令至少一条 E2E：
- `modu scan`、`modu scan --export`
- `modu init`（可用 fixture 或 mock 远端）
- `modu create` / `modu list` / `modu info` / `modu delete`
- 断言退出码、stdout/stderr 关键内容

---

## 9. 分发与升级

- 使用 **GoReleaser** 打包，支持 `brew install` 接入
- 内置 `modu upgrade` 命令，通过 GitHub API 检查并替换二进制文件
- 支持平台：Darwin/Linux arm64/amd64
- 一键安装：`curl -sSL https://get.modu.sh | sh`

---

## 10. 附录：错误处理行为矩阵

| 场景 | CLI 行为 | TUI 行为 |
|------|----------|----------|
| 配置/路径错误 | 直接退出并打印错误 | 在界面提示，不退出 |
| Git 失败 | 带出完整错误信息 | 界面提示并允许重试 |
| 并发部分失败 | 打印成功/失败汇总 | 显示失败模块列表 |
| 脏检查失败 | 退出码非 0 | 弹窗阻止删除 |

---

## 版本变更记录

| 版本 | 日期 | 变更内容 |
|------|------|----------|
| v2.4 | 2026-03-06 | 整合 v2.2 完整性 + v2.3 架构规范 |
