## 1. 代码实现

- [x] 1.1 在 internal/config/config.go 中添加 resolveEnvVars 函数
- [x] 1.2 在 loadConfigImpl 中调用 resolveEnvVars 解析 workspace 和 worktree-root
- [x] 1.3 添加 extractUndefinedVars 辅助函数

## 2. 测试

- [x] 2.1 添加单元测试：$VAR 格式环境变量解析
- [x] 2.2 添加单元测试：${VAR} 格式环境变量解析
- [x] 2.3 添加单元测试：环境变量未定义时报错
- [x] 2.4 添加单元测试：环境变量正确定义时正常展开
- [x] 2.5 添加单元测试：未配置的字段不进行检查
- [x] 2.6 添加单元测试：无环境变量的配置正常工作

## 3. 文档

- [x] 3.1 更新 README.md 添加环境变量配置说明

## 4. 验证

- [x] 4.1 运行 go build 确保编译通过
- [x] 4.2 运行 go test ./... 确保测试通过
- [x] 4.3 运行 pre-commit run --all-files 确保代码规范
