# modu 架构规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## 目的

采用 Go Clean Architecture 简化版，确保核心逻辑（Domain）不依赖具体实现（Git/TUI）。

## 分层与依赖

```
cmd/modu          → 入口：Cobra 路由、全局 Flag、环境初始化
     ↓
internal/core    → 领域对象：Repo, Worktree, WorktreeEnv, ModuleStatus
     ↓
internal/engine   → 核心控制器：并发调度、事务编排、脏检查
     ↓
internal/gitproxy → Git 原语封装：命令调用及 stderr 解析
     ↓
internal/ui      → Bubble Tea 状态机，仅负责 View 渲染
```

同级或独立：`internal/config`（配置加载与校验）、`internal/output`（Table/JSON 输出）。

## 目录结构

```
.
├── cmd/modu/main.go       # Cobra 路由、CLI 参数与 TUI 切换
├── internal/
│   ├── core/              # 领域对象（WorktreeEnv, ModuleStatus）
│   ├── config/            # 配置加载、校验、查找
│   ├── gitproxy/          # Git 命令封装（Worktree, Status, Clone）
│   ├── engine/            # 并发调度、脏检查、Create 回滚
│   ├── ui/                # Bubble Tea TUI
│   └── output/            # 结构化输出（Table/JSON）
├── modu.yaml              # 示例配置
└── go.mod
```

## 依赖规则

- **core** 不依赖 config/gitproxy/engine/ui/output
- **engine** 依赖 core、config、gitproxy；不依赖 ui、output
- **cmd** 可依赖所有 internal 包，负责组装与路由
