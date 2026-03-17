## MODIFIED Requirements

### Requirement: feature 名与目录名的转换
feature 名 SHALL 支持包含 `/` 的分支名，系统 SHALL 自动将 `/` 转换为 `-` 作为文件系统目录名。

#### 原始描述
扫描 `worktree-root` 下子目录，每个子目录名视为 feature 名。

#### 新描述
feature 名在创建时进行转换，列表时直接显示目录名：
- **创建时**：将 feature 名中的 `/` 替换为 `-` 作为目录名（如 `feature/abc` → 目录 `feature-abc`）
- **列出时**：直接使用目录名作为 feature 名显示（如目录 `feature-abc` → 显示为 `feature-abc`）

feature 名必须满足：仅包含字母，数字，`-`，`_`。

#### 场景

##### Scenario: 创建带 / 的 feature
- **GIVEN** 用户执行 `modu create feature/abc develop`
- **WHEN** 系统创建 worktree
- **AND** 在文件系统中创建目录 `worktree-root/feature-abc`
- **AND** git 分支名为 `feature/abc`

##### Scenario: 列出带 / 的 feature
- **GIVEN** worktree-root 下存在目录 feature-abc
- **AND** feature-abc 包含主项目 worktree
- **WHEN** 执行 ListWorktrees
- **THEN** 返回 feature 名称为 `feature-abc`

##### Scenario: 删除带 / 的 feature
- **GIVEN** 存在 feature 名称 `feature/abc`（对应目录 feature-abc）
- **WHEN** 执行 `modu delete feature/abc`
- **AND** 删除目录 `worktree-root/feature-abc`

##### Scenario: 更新带 / 的 feature
- **GIVEN** 存在 feature 名称 `feature/abc`（对应目录 feature-abc）
- **WHEN** 执行 `modu update feature/abc`
- **AND** 更新目录 `worktree-root/feature-abc` 下的 worktree
