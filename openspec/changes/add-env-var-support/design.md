## 概述

支持在 .modu.yaml 配置文件的 workspace 和 worktree-root 字段中使用环境变量，允许不同开发者在不同机器上使用同一份配置文件。

## 设计

### 核心函数

在 `internal/config/config.go` 的 `loadConfigImpl` 函数中添加环境变量解析逻辑：

```go
// 环境变量解析（只检查已配置的字段）
if cfg.Workspace != "" {
    if err := resolveEnvVars(&cfg.Workspace, "workspace"); err != nil {
        return nil, err
    }
}
if cfg.WorktreeRoot != "" {
    if err := resolveEnvVars(&cfg.WorktreeRoot, "worktree-root"); err != nil {
        return nil, err
    }
}
```

### resolveEnvVars 函数

```go
// resolveEnvVars 解析并验证环境变量
// 如果值包含 $VAR 或 ${VAR}，则验证该环境变量是否存在
func resolveEnvVars(value *string, fieldName string) error {
    expanded := os.ExpandEnv(*value)

    // 检查是否有未展开的环境变量（环境变量未定义）
    // os.ExpandEnv 会保留未定义的环境变量原样
    if strings.Contains(expanded, "$") {
        undefinedVars := extractUndefinedVars(*value)
        return fmt.Errorf("%w: field '%s' contains undefined environment variable(s): %s",
            errs.ErrConfigInvalid, fieldName, strings.Join(undefinedVars, ", "))
    }

    *value = expanded
    return nil
}

// extractUndefinedVars 从值中提取未定义的环境变量名
func extractUndefinedVars(value string) []string {
    re := regexp.MustCompile(`\$\{?([A-Za-z_][A-Za-z0-9_]*)\}?`)
    matches := re.FindAllStringSubmatch(value, -1)

    var undefined []string
    seen := make(map[string]bool)
    for _, m := range matches {
        varName := m[1]
        if _, ok := os.LookupEnv(varName); !ok && !seen[varName] {
            undefined = append(undefined, m[0])
            seen[varName] = true
        }
    }
    return undefined
}
```

### 错误示例

输入配置：
```yaml
workspace: $MY_WORKSPACE
worktree-root: /opt/worktrees
```

报错：
```
Error: invalid config: field 'workspace' contains undefined environment variable(s): $MY_WORKSPACE
```

## 数据流

1. 用户提供 .modu.yaml 配置文件
2. `LoadConfig()` 调用 `loadConfigImpl()` 读取并解析 YAML
3. 在路径转换前，调用 `resolveEnvVars()` 解析环境变量
4. 如果字段已配置且包含未定义的环境变量，报错停止
5. 继续现有的相对路径转绝对路径逻辑

## 测试

- 测试 `$VAR` 格式的环境变量解析
- 测试 `${VAR}` 格式的环境变量解析
- 测试环境变量未定义的报错
- 测试环境变量正确定义时正常展开
- 测试未配置的字段不进行检查

## 文档更新

在 README.md 添加环境变量配置说明。
