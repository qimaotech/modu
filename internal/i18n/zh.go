package i18n

// 中文翻译
var zhMessages = Messages{
	// 通用
	"error_prefix":          "错误",
	"config_not_found":      "配置文件不存在",
	"config_not_found_tip":  "运行 'modu init' 或 'modu tui' 启动配置向导",
	"failed_to_load_config": "加载配置文件失败",
	"config_file_not_found": "配置文件不存在",
	"config_file_at":        "配置文件",
	"please_run_command":    "请运行以下命令创建配置文件:",
	"or_use_wizard":         "或使用交互式向导:",
	"modu_config_create":    "modu config create",
	"modu_init":             "modu init",
	"config_create_failed":  "创建配置文件失败",
	"config_invalid":        "配置无效",
	"please_check_config":   "请检查配置文件格式是否正确",

	// loadConfig 特定
	"workspace_required":     "请在配置文件中设置 workspace 字段",
	"worktree_root_required": "请在配置文件中设置 worktree-root 字段",
	"default_base_required":  "请在配置文件中设置 default-base 字段",
	"no_module_configured":   "配置文件中没有模块，请添加至少一个模块",
	"add_module_hint":        "modu config create --module \"name=https://github.com/xxx.git\"",
	"scan_for_new_modules":   "你也可以复制已有的 Git 文件夹到当前目录，再执行 modu config scan",

	// TUI
	"starting_config_wizard": "配置文件不存在，正在启动配置向导...",
	"config_saved":           "配置已保存",
	"workspace":              "Workspace",
	"worktree":               "Worktree",
	"default_branch":         "默认分支",
	"manual_steps":           "手动执行",
	"move_git_projects":      "将您已有的 Git 项目移动到 Workspace 目录",
	"run_config_scan":        "执行 modu config scan",

	// init 命令
	"scanning_for_modules":    "正在扫描目录自动发现模块...",
	"created_default_config":  "已创建默认配置文件",
	"please_edit_config":      "请编辑配置文件添加模块后再次运行",
	"no_module_found":         "未发现任何模块",
	"scan_for_modules_hint":   "提示: 当前没有配置模块，如需自动发现模块请运行",
	"initialized_repos":       "Initialized all repositories",
	"update_gitignore_failed": "警告: 更新 .gitignore 失败",

	// runConfigScan
	"config_not_exist":      "配置文件 %s 不存在，请先运行 modu config create",
	"load_config_failed":    "加载配置失败",
	"scan_dir_failed":       "扫描目录失败",
	"no_new_git_repo":       "未发现新的 git 仓库",
	"scanned_added_modules": "已扫描并添加 %d 个模块",
	"scan_success_hint":     "准备好创建 feature 了吗？使用 modu create feature/<name>",
	"module_branch_missing": "⚠️ 模块 %s 缺少分支 %s，请先在仓库中创建该分支",

	// runCreate
	"feature_already_exists":    "Feature %s 已存在，以下模块已创建",
	"tui_not_available":         "TUI 不可用",
	"operation_cancelled":       "已取消操作，保留所有已存在的模块",
	"no_module_selected":        "未选择任何模块，将删除所有已存在的模块（如需保留请按 q 退出）",
	"will_create_modules":       "将创建以下新模块",
	"deleting_modules":          "删除模块",
	"delete_failed":             "删除失败",
	"delete_success":            "已删除（含分支）",
	"operation_complete":        "操作完成",
	"created_successfully":      "Successfully created feature",
	"failed_to_create":          "Failed to create feature",
	"failed_to_create_worktree": "创建 worktree 失败",
	"worktree_invalid_ref":      "分支不存在或没有提交，请先在仓库中创建至少一个提交",

	// runDelete
	"deleted_successfully": "已删除 feature",
	"failed_to_delete":     "删除 feature 失败",

	// runList
	"failed_to_list":       "Failed to list worktrees",
	"failed_to_get_status": "Failed to get status",
	"features":             "Features",
	"workspace_header":     "Workspace",

	// runInfo
	"feature": "Feature",
	"modules": "Modules",
	"branch":  "Branch",
	"status":  "Status",
	"path":    "Path",
	"clean":   "clean",
	"dirty":   "dirty",

	// runUpdate
	"update_success_main":         "更新成功: 主项目",
	"update_success_with_modules": "更新成功: 主项目 + %d 个模块",
	"update_success_feature":      "更新成功: feature %s（主项目 + %d 个模块）",
	"update_partial":              "更新成功: %d 个，失败: %d 个 (%s)",
	"feature_not_exist":           "feature %s 不存在",

	// runVersion
	"version_info": "modu version %s",
	"commit":       "commit",
	"date":         "date",
	"go_version":   "go",

	// config create
	"config_created": "配置文件已创建",
}
