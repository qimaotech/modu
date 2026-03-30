# CLI Info Autodetect

**版本**: 1.0 | **来源**: change: info-auto-detect-feature

## Purpose

TBD

## ADDED Requirements

### Requirement: modu info 无参数时自动推断 feature

`modu info` 命令在未传入 `<feature>` 参数时 SHALL 自动从当前工作目录向上回溯，推断所属 feature 并展示其详情。

#### Scenario: 在 feature 根目录执行无参数 info

- **WHEN** 用户在 `worktreeRoot/<feature>/` 目录下执行 `modu info`（无参数）
- **THEN** 系统 SHALL 识别当前目录所属的 feature 并显示其详情，等同于 `modu info <feature>`

#### Scenario: 在 feature 子目录执行无参数 info

- **WHEN** 用户在 `worktreeRoot/<feature>/<module>/` 目录下执行 `modu info`（无参数）
- **THEN** 系统 SHALL 向上回溯找到 `worktreeRoot` 的直接子目录作为 feature 并显示其详情

#### Scenario: 在 feature 目录内执行带参数 info

- **WHEN** 用户在 feature 目录下执行 `modu info other-feature`
- **THEN** 系统 SHALL 显示 `other-feature` 的详情（参数优先，不进行自动推断）

#### Scenario: 当前目录不在任何 feature 下

- **WHEN** 用户在 `worktreeRoot` 目录本身或之外执行 `modu info`（无参数）
- **THEN** 系统 SHALL 显示错误信息：当前目录不在任何 feature 下

#### Scenario: 当前目录恰好是 worktreeRoot

- **WHEN** 用户在 `worktreeRoot` 目录下执行 `modu info`（无参数）
- **THEN** 系统 SHALL 显示错误信息并提示需要指定 feature 参数

#### Scenario: 指定的 feature 不存在

- **WHEN** 用户执行 `modu info nonexistent` 且 `nonexistent` 目录不存在于 `worktreeRoot` 下
- **THEN** 系统 SHALL 复用现有的 `GetWorktreeInfo` 错误处理，显示 feature 不存在的错误
