package i18n

// 英文翻译
var enMessages = Messages{
	// 通用
	"error_prefix":          "Error",
	"config_not_found":      "Configuration file not found",
	"config_not_found_tip":  "Run 'modu init' or 'modu tui' to start the configuration wizard",
	"failed_to_load_config": "Failed to load config",
	"config_file_not_found": "Config file not found",
	"config_file_at":        "Config file",
	"please_run_command":    "Run the following command to create a config file:",
	"or_use_wizard":         "Or use the interactive wizard:",
	"modu_config_create":    "modu config create",
	"modu_init":             "modu init",
	"config_create_failed":  "Failed to create config file",
	"config_invalid":        "Invalid configuration",
	"please_check_config":   "Please check if the config file format is correct",

	// loadConfig 特定
	"workspace_required":     "Please set the workspace field in the config file",
	"worktree_root_required": "Please set the worktree-root field in the config file",
	"default_base_required":  "Please set the default-base field in the config file",
	"no_module_configured":   "No modules configured, please add at least one module",
	"add_module_hint":        "modu config create --module \"name=https://github.com/xxx.git\"",
	"scan_for_new_modules":   "You can also copy existing Git folders to the current directory and run modu config scan",

	// TUI
	"starting_config_wizard": "Config file not found, starting configuration wizard...",
	"config_saved":           "Configuration saved",
	"workspace":              "Workspace",
	"worktree":               "Worktree",
	"default_branch":         "Default branch",
	"manual_steps":           "Manual steps",
	"move_git_projects":      "Move your existing Git projects to the Workspace directory",
	"run_config_scan":        "Run modu config scan",

	// init 命令
	"scanning_for_modules":    "Scanning directory to discover modules...",
	"created_default_config":  "Created default config file",
	"please_edit_config":      "Please edit the config file to add modules and run again",
	"no_module_found":         "No modules found",
	"scan_for_modules_hint":   "Hint: No modules configured. To auto-discover modules, run:",
	"initialized_repos":       "Initialized all repositories",
	"update_gitignore_failed": "Warning: Failed to update .gitignore",

	// runConfigScan
	"config_not_exist":      "Config file %s does not exist, please run modu config create first",
	"load_config_failed":    "Failed to load config",
	"scan_dir_failed":       "Failed to scan directory",
	"no_new_git_repo":       "No new git repositories found",
	"scanned_added_modules": "Scanned and added %d modules",
	"scan_success_hint":     "Ready to create a feature? Use modu create feature/<name>",
	"module_branch_missing": "⚠️ Module %s is missing branch %s, please create the branch in the repository first",

	// runCreate
	"feature_already_exists":    "Feature %s already exists, following modules have been created",
	"tui_not_available":         "TUI not available",
	"operation_cancelled":       "Operation cancelled, all existing modules preserved",
	"no_module_selected":        "No module selected, will delete all existing modules (press q to quit and preserve)",
	"will_create_modules":       "Will create the following modules",
	"deleting_modules":          "Deleting modules",
	"delete_failed":             "Delete failed",
	"delete_success":            "Deleted (including branch)",
	"operation_complete":        "Operation complete",
	"created_successfully":      "Successfully created feature",
	"failed_to_create":          "Failed to create feature",
	"failed_to_create_worktree": "Failed to create worktree",
	"worktree_invalid_ref":      "Branch does not exist or has no commits. Please create at least one commit in the repository first",

	// runDelete
	"deleted_successfully": "Successfully deleted feature",
	"failed_to_delete":     "Failed to delete feature",

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
	"update_success_main":         "Update success: main project",
	"update_success_with_modules": "Update success: main project + %d modules",
	"update_success_feature":      "Update success: feature %s (main project + %d modules)",
	"update_partial":              "Update success: %d, failed: %d (%s)",
	"feature_not_exist":           "feature %s does not exist",

	// runVersion
	"version_info": "modu version %s",
	"commit":       "commit",
	"date":         "date",
	"go_version":   "go",

	// config create
	"config_created": "Config file created",
}
