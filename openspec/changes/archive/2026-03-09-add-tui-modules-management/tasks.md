## 1. Engine 层 - 单模块操作

- [x] 1.1 在 Engine 结构体添加 Config 字段引用（支持访问 modules 配置）
- [x] 1.2 实现 AddModule 方法：为 feature 添加单个模块的 worktree
- [x] 1.3 实现 RemoveModule 方法：为 feature 删除单个模块的 worktree
- [x] 1.4 添加模块增删的脏检查逻辑

## 2. UI 层 - 数据结构变更

- [x] 2.1 在 App 结构体添加 `moduleSelector` 字段（存储模块选择状态）
- [x] 2.2 在 App 结构体添加 `moduleCursor` 字段（模块列表光标位置）
- [x] 2.3 在 App 结构体添加 `modulesFeature` 字段（当前操作的 feature 名称）
- [x] 2.4 在 state 字段支持 "modules" 状态

## 3. UI 层 - 状态处理

- [x] 3.1 在 Update 方法中添加 "modules" 状态的 case 处理
- [x] 3.2 创建 `handleModulesKey` 方法处理模块管理的键盘事件
- [x] 3.3 在 `handleListKey` 中添加 "m" 键处理：进入模块管理视图

## 4. UI 层 - 键盘事件处理

- [x] 4.1 在 `handleModulesKey` 中实现：上下键导航模块列表
- [x] 4.2 在 `handleModulesKey` 中实现：空格键切换选中状态
- [x] 4.3 在 `handleModulesKey` 中实现：回车键确认执行模块增删
- [x] 4.4 在 `handleModulesKey` 中实现：q/esc 键返回操作菜单
- [x] 4.5 在 `handleMenuKey` 中添加 "m" 键处理：进入模块管理视图
- [x] 4.6 修改操作菜单项顺序：打开 VS Code → Modules 管理 → 删除
- [x] 4.7 修改 menuSelected 范围从 1 改为 2

## 5. UI 层 - 视图渲染

- [x] 5.1 在 View 方法中添加 "modules" 状态的渲染分支
- [x] 5.2 创建 `renderModules` 方法渲染模块管理界面
- [x] 5.3 renderModules 显示所有配置模块，已创建的标记 `[x]`，未创建的标记 `[ ]`
- [x] 5.4 renderModules 显示当前操作的 feature 名称

## 6. 测试与验证

- [x] 6.1 运行 `go build` 确保代码编译通过
- [x] 6.2 手动测试：列表视图按 m 进入模块管理
- [x] 6.3 手动测试：操作菜单按 m 进入模块管理
- [x] 6.4 手动测试：模块列表上下键导航
- [x] 6.5 手动测试：空格键切换选中状态
- [x] 6.6 手动测试：回车键确认添加模块
- [x] 6.7 手动测试：回车键确认删除模块
- [x] 6.8 手动测试：q/esc 返回操作菜单
- [x] 6.9 手动测试：操作菜单顺序正确（打开 VS Code → Modules 管理 → 删除）
