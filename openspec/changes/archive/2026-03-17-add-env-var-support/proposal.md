## Why

当前 .modu.yaml 配置文件中的 workspace 和 worktree-root 字段使用硬编码路径，导致在不同开发者的机器上无法共享同一份配置文件。每个人的工作目录不同，需要手动修改配置文件，降低了团队协作效率。

## What Changes

- 在配置加载时解析 `workspace` 和 `worktree-root` 字段中的环境变量（支持 `$VAR` 和 `${VAR}` 语法）
- 当环境变量未定义时，报错并提示具体是哪个字段、哪个环境变量缺失
- 未配置的字段不进行检查
- 更新 README.md 文档说明环境变量支持

## Capabilities

### New Capabilities
- `config-env-var`: 支持在 workspace 和 worktree-root 配置中使用环境变量

## Impact

- 修改 `internal/config/config.go` 添加环境变量解析逻辑
- 修改 `README.md` 添加环境变量配置说明
- 不影响现有功能，未配置环境变量的配置文件保持原样工作
