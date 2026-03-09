## 1. 数据结构变更

- [x] 1.1 在 App 结构体添加 `menuSelected` 字段（int 类型），跟踪操作菜单选中项

## 2. 状态处理

- [x] 2.1 在 Update 方法中添加 "menu" 状态的 case 处理
- [x] 2.2 创建 `handleMenuKey` 方法处理操作菜单的键盘事件

## 3. 键盘事件处理

- [x] 3.1 修改 `handleListKey` 方法：Enter 键从进入确认改为进入操作菜单
- [x] 3.2 在 `handleListKey` 中添加 "d" 键处理：直接触发删除确认
- [x] 3.3 在 `handleMenuKey` 中实现：上下键导航菜单项
- [x] 3.4 在 `handleMenuKey` 中实现：d 键触发删除确认
- [x] 3.5 在 `handleMenuKey` 中实现：o 键打开 VS Code
- [x] 3.6 在 `handleMenuKey` 中实现：esc/q 键返回列表视图
- [x] 3.7 在 `handleMenuKey` 中实现：Enter 键执行当前选中操作
- [x] 3.8 打开 VS Code 后自动返回列表视图（按 o 或 Enter）

## 4. 视图渲染

- [x] 4.1 在 View 方法中添加 "menu" 状态的渲染分支
- [x] 4.2 创建 `renderMenu` 方法渲染操作菜单界面
- [x] 4.3 操作菜单显示"打开 VS Code"和"删除"两个选项，带选中高亮（打开在前，删除在后）
- [x] 4.4 更新列表视图帮助文案，增加 d 删除描述

## 5. 测试与验证

- [x] 5.1 运行 `go build` 确保代码编译通过
- [x] 5.2 手动测试：Enter 进入操作菜单
- [x] 5.3 手动测试：操作菜单内上下键导航
- [x] 5.4 手动测试：操作菜单内按 d 删除
- [x] 5.5 手动测试：操作菜单内按 o 打开 VS Code
- [x] 5.6 手动测试：列表视图直接按 d 删除
- [x] 5.7 手动测试：操作菜单内按 Enter 执行选中操作
- [x] 5.8 手动测试：打开 VS Code 后自动返回列表视图
