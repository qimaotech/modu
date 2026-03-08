package core

// WorktreeEnv 表示一个 feature 环境，包含多个模块的工作树
type WorktreeEnv struct {
	Name        string         `json:"name"`        // Feature 名称
	Base        string         `json:"base"`        // 基准分支
	MainProject *ModuleStatus  `json:"mainProject"` // 主项目状态（可选）
	Modules     []ModuleStatus `json:"modules"`     // 各模块在该环境下的状态
}

// ModuleStatus 记录单个模块的工作树状态
type ModuleStatus struct {
	Name    string `json:"name"`            // 模块名
	Path    string `json:"path"`            // 物理路径
	IsDirty bool   `json:"isDirty"`         // 是否存在未提交修改
	Branch  string `json:"branch"`          // 当前分支
	Error   error  `json:"error,omitempty"` // 记录该模块操作失败的具体原因
}
