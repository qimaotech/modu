# modu 配置规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## 目的

定义 modu 配置文件结构、加载与校验规则，以及 config 相关命令行为。

## 配置结构

### Config

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `workspace` | string | 是 | 裸仓库/主仓库所在目录 |
| `worktree-root` | string | 是 | 特性分支代码存放目录 |
| `default-base` | string | 是 | 默认基准分支（如 develop） |
| `concurrency` | int | 否 | 并发数，默认 5 |
| `auto-fetch` | bool | 否 | 操作前自动 fetch |
| `strict-dirty-check` | bool | 否 | 删除前强制脏检查 |
| `modules` | []Module | 是 | 模块列表，至少一个 |

### Module

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 模块名称 |
| `url` | string | 是 | 仓库 URL |
| `base-branch` | string | 否 | 覆盖全局 default-base |

## 加载与校验

- 支持通过 `-c`/`--config` 指定配置文件路径，默认 `.modu.yaml`。
- **必填校验**：缺失 `workspace`、`worktree-root`、`default-base` 或 `modules` 为空时，返回 `ERR_CONFIG_INVALID`。
- **路径**：`workspace`、`worktree-root` 若为相对路径，则相对于配置文件所在目录解析为绝对路径。
- **LoadConfigForScan**：scan 场景可仅校验基础字段，不强制校验 modules（用于先扫后写配置）。

## 配置命令

- **config create**：创建配置文件；支持 `--workspace`、`--worktree-root`、`--default-base`、`--module name=url`（可多次）。
- **config scan**：扫描当前目录（或 workspace）发现模块，可更新或导出 YAML；存在配置文件时可确认是否覆盖。

## 配置示例（modu.yaml）

```yaml
version: "2.4"
workspace: ./workspace
worktree-root: ./worktrees
default-base: develop
concurrency: 8
auto-fetch: true
strict-dirty-check: true

modules:
  - name: pixiu-ad-server
    url: git@github.com:commerce/server.git
```

## 与代码的对应

- 实现：`internal/config`（Config、Module、LoadConfig、LoadConfigForScan、validate、validateBasic）。
