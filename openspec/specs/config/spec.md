# modu 配置规范

**版本**: 2.4 | **来源**: docs/plans/2026-03-06-modu-design-v2.4.md

## Purpose

定义 modu 配置文件结构、加载与校验规则，以及 config 相关命令行为。本规范约束 CLI 和 TUI 共享的配置字段、兼容性要求、路径解析方式和用户可扩展入口。

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
| `app-openers` | []AppOpener | 否 | TUI 操作菜单中额外可用的 GUI 应用打开器 |
| `modules` | []Module | 是 | 模块列表，至少一个 |

### Module

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 模块名称 |
| `url` | string | 是 | 仓库 URL |
| `base-branch` | string | 否 | 覆盖全局 default-base |

### AppOpener

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 配置项标识 |
| `app` | string | 是 | macOS“应用程序”中的应用名称，例如 VS Code 应填 `Visual Studio Code`，而不是 `vscode` 或 CLI 命令 `code` |
| `label` | string | 否 | 操作菜单展示名称；为空时使用 `app` |
| `shortcut` | string | 否 | 操作菜单直接执行快捷键，必须是单个可打印字符 |

## 加载与校验

- 支持通过 `-c`/`--config` 指定配置文件路径，默认 `.modu.yaml`。
- **必填校验**：缺失 `workspace`、`worktree-root`、`default-base` 或 `modules` 为空时，返回 `ERR_CONFIG_INVALID`。
- **路径**：`workspace`、`worktree-root` 若为相对路径，则相对于配置文件所在目录解析为绝对路径。
- **环境变量**：`workspace`、`worktree-root` 支持 `$VAR` 和 `${VAR}` 语法，环境变量必须已定义，否则加载失败并报错。详见 [config-env-var](../config-env-var/spec.md)。
- **App opener 校验**：`app-openers` 为可选字段；每个条目必须提供 `name` 和 `app`，`shortcut` 若存在则必须是单个可打印字符。快捷键冲突不使配置加载失败，由 TUI 操作菜单在展示与绑定时处理。
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

app-openers:
  - name: zed
    app: Zed
    label: Zed
    shortcut: z
```

## Requirements

### Requirement: Configurable app openers

配置文件 SHALL support an optional `app-openers` list for declaring extra GUI applications that can open the currently selected project path from the TUI operation menu. Each opener's `app` field SHALL use the macOS application name as shown in Applications, not a short alias or CLI command.

#### Scenario: Load Zed opener example
- **WHEN** `.modu.yaml` contains an `app-openers` entry with `name: zed`, `app: Zed`, `label: Zed`, and `shortcut: z`
- **THEN** config loading succeeds and exposes an app opener whose display text can be rendered as "打开 Zed" with shortcut `z`

#### Scenario: Missing app openers remains compatible
- **WHEN** `.modu.yaml` does not contain `app-openers`
- **THEN** config loading succeeds and the TUI uses only built-in operation menu items

#### Scenario: App opener without shortcut
- **WHEN** an `app-openers` entry omits `shortcut`
- **THEN** config loading succeeds and the opener is available for menu selection without a direct shortcut key

### Requirement: App opener config validation

Each configured app opener MUST include enough information to identify the app and MUST use a valid shortcut shape when a shortcut is provided.

#### Scenario: Missing opener name is invalid
- **WHEN** an `app-openers` entry has an empty `name`
- **THEN** config loading fails with `ERR_CONFIG_INVALID`

#### Scenario: Missing app name is invalid
- **WHEN** an `app-openers` entry has an empty `app`
- **THEN** config loading fails with `ERR_CONFIG_INVALID`

#### Scenario: Shortcut must be a single printable key
- **WHEN** an `app-openers` entry sets `shortcut` to more than one character or to a non-printable key
- **THEN** config loading fails with `ERR_CONFIG_INVALID`

#### Scenario: Shortcut conflict does not invalidate config
- **WHEN** an `app-openers` entry sets `shortcut: c` and `c` is already used by a built-in TUI menu action
- **THEN** config loading still succeeds because shortcut conflict resolution is handled by the TUI menu

### Requirement: Auto-fetch controls feature creation fetch

The configuration field `auto-fetch` SHALL control whether feature creation performs pre-create fetch operations before adding worktrees.

#### Scenario: Auto-fetch enabled keeps fetch-first creation

- **WHEN** `.modu.yaml` omits `auto-fetch` or sets `auto-fetch: true`
- **THEN** feature creation attempts to fetch the main project and selected module repositories before creating their worktrees

#### Scenario: Auto-fetch disabled uses local branches

- **WHEN** `.modu.yaml` sets `auto-fetch: false`
- **THEN** feature creation skips pre-create fetch and creates worktrees from local branch state

#### Scenario: Auto-fetch disabled still uses configured base branches

- **WHEN** feature creation runs with `auto-fetch: false`
- **THEN** the main project uses the requested base branch and each module uses its configured `base-branch` or the global `default-base`

## 与代码的对应

- 实现：`internal/config`（Config、Module、AppOpener、LoadConfig、LoadConfigForScan、validate、validateBasic）。
