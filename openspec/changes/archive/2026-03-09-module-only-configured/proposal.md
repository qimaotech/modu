# 变更：仅将配置内目录视为模块

## 问题

- `modu list` 会把 feature 下所有子目录（含 `.claude`、`openspec`）当模块展示。
- create 增删模块时，会把「已存在但未在配置中的目录」也纳入删除逻辑，导致 `.claude`、`openspec` 等被误删。

## 方案

凡涉及「模块」的 list/create/delete 逻辑，仅处理 `Config.Modules` 中的目录；其他子目录一律忽略（不展示、不参与增删、不参与脏检查）。实现后已同步到 `openspec/specs/engine/spec.md` 与 `openspec/specs/domain/spec.md`。
