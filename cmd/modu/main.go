package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"codeup.aliyun.com/qimao/public/devops/modu/internal/config"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/engine"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/errors"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/output"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/ui"

	"github.com/spf13/cobra"
)

// 版本信息，运行时从 git 自动获取
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func init() {
	ctx := context.Background()
	// 启动时自动从 git 获取版本信息
	if v, err := gitDescribe(ctx); err == nil {
		version = v
	}
	if c, err := gitCommit(ctx); err == nil {
		commit = c
	}
	if d, err := gitDate(ctx); err == nil {
		date = d
	}
}

func gitDescribe(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "describe", "--tags", "--abbrev=0").Output()
	if err != nil {
		return "dev", fmt.Errorf("git describe: %w", err)
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "v"), nil
}

func gitCommit(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD").Output()
	return strings.TrimSpace(string(out)), err
}

func gitDate(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "log", "-1", "--format=%ci").Output()
	return strings.TrimSpace(string(out)), err
}

// isInteractiveTerminal 检查是否是交互式终端
func isInteractiveTerminal() bool {
	// 检查 stdin/stdout/stderr 是否为 TTY
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice != 0
}

var (
	configPath string
	outputFmt  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "modu",
		Short: "modu - 多模块 Git Worktree 管理工具",
		Long:  `modu 是一个基于 Go 语言开发的多模块 Git Worktree 管理工具，用于简化多仓库协作开发流程。`,
		Run: func(cmd *cobra.Command, args []string) {
			// 无子命令时启动 TUI
			if len(args) == 0 {
				if err := ui.StartTUI(configPath); err != nil {
					// 检查是否是配置文件不存在错误
					if config.IsConfigNotFoundError(err) {
						fmt.Println("配置文件不存在，是否现在创建?")
						fmt.Println("运行 'modu init' 或 'modu tui' 启动配置向导")
						fmt.Println()
						fmt.Printf("错误: %v\n", err)
						os.Exit(1)
					}
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				return
			}
			// 显示帮助
			_ = cmd.Help()
		},
	}

	// 全局 flag
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".modu.yaml", "配置文件路径")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "text", "输出格式: text, json")

	// create 命令
	createCmd := &cobra.Command{
		Use:   "create <feature>",
		Short: "创建 feature 工作树",
		Args:  cobra.ExactArgs(1),
		Run:   runCreate,
	}
	createCmd.Flags().String("base", "develop", "基准分支")
	createCmd.Flags().StringSlice("modules", nil, "指定要创建的模块（逗号分隔），默认创建所有模块")

	// delete 命令
	deleteCmd := &cobra.Command{
		Use:   "delete <feature>",
		Short: "删除 feature 工作树",
		Args:  cobra.ExactArgs(1),
		Run:   runDelete,
	}
	deleteCmd.Flags().BoolP("force", "f", false, "强制删除（跳过脏检查）")

	// list 命令
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有 feature 工作树",
		Run:   runList,
	}
	listCmd.Flags().BoolP("verbose", "v", false, "显示详细信息（模块、分支、状态）")

	// info 命令
	infoCmd := &cobra.Command{
		Use:   "info <feature>",
		Short: "查看 feature 详情",
		Args:  cobra.ExactArgs(1),
		Run:   runInfo,
	}

	// config 命令 - 配置管理
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "配置管理",
	}
	configCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "创建配置文件",
		Run:   runConfigCreate,
	}
	configCreateCmd.Flags().String("workspace", ".", "仓库目录")
	configCreateCmd.Flags().String("worktree-root", "../worktrees", "Worktree 存放目录")
	configCreateCmd.Flags().String("default-base", "develop", "默认基准分支")
	configCreateCmd.Flags().StringArray("module", []string{}, "模块 (格式: name=url)")
	configCmd.AddCommand(configCreateCmd)

	// config scan 命令 - 扫描当前目录发现模块并更新配置
	configScanCmd := &cobra.Command{
		Use:   "scan",
		Short: "扫描当前目录自动发现模块",
		Run:   runConfigScan,
	}
	configCmd.AddCommand(configScanCmd)

	// init 命令 - 初始化仓库（克隆所有配置的仓库）
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "初始化仓库（克隆所有配置的仓库）",
		Run:   runInit,
	}
	initCmd.Flags().Bool("scan", false, "自动扫描并添加模块")

	// version 命令
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run:   runVersion,
	}

	// status 命令
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "显示所有模块的脏状态",
		Run:   runStatus,
	}

	// update 命令 - 更新主项目或指定 feature 的 worktree
	updateCmd := &cobra.Command{
		Use:   "update [feature]",
		Short: "更新代码（主项目或指定 feature 的 worktree，fetch + rebase）",
		Long:  "无参数时更新主项目（workspace + 所有模块）；带 feature 时更新该 feature 的 worktree。",
		Args:  cobra.MaximumNArgs(1),
		Run:   runUpdate,
	}

	// tui 命令（显式启动 TUI）
	tuiCmd := &cobra.Command{
		Use:   "tui",
		Short: "启动 TUI 界面",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.StartTUI(configPath); err != nil {
				// 检查是否是配置文件不存在错误
				if config.IsConfigNotFoundError(err) {
					fmt.Println("配置文件不存在，正在启动配置向导...")
					fmt.Println()
					if err := ui.RunConfigWizard(); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(1)
					}
					return
				}
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(createCmd, deleteCmd, listCmd, infoCmd, configCmd, initCmd, statusCmd, updateCmd, tuiCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadConfig() *engine.Engine {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n\n", err)

		// 检查配置文件是否存在
		if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
			fmt.Fprintf(os.Stderr, "配置文件 %s 不存在\n", configPath)
			fmt.Fprintf(os.Stderr, "请运行以下命令创建配置文件:\n")
			fmt.Fprintf(os.Stderr, "  modu config create\n")
			fmt.Fprintf(os.Stderr, "或使用交互式向导:\n")
			fmt.Fprintf(os.Stderr, "  modu init\n")
		} else if config.IsConfigValidationError(err) {
			// 配置验证错误，给出具体提示
			errMsg := err.Error()
			if strings.Contains(errMsg, "at least one module is required") {
				fmt.Fprintf(os.Stderr, "配置文件中没有模块，请添加至少一个模块:\n")
				fmt.Fprintf(os.Stderr, "  modu config create --module \"name=https://github.com/xxx.git\"\n")
			} else if strings.Contains(errMsg, "workspace is required") {
				fmt.Fprintf(os.Stderr, "请在配置文件中设置 workspace 字段\n")
			} else if strings.Contains(errMsg, "worktree-root is required") {
				fmt.Fprintf(os.Stderr, "请在配置文件中设置 worktree-root 字段\n")
			} else if strings.Contains(errMsg, "default-base is required") {
				fmt.Fprintf(os.Stderr, "请在配置文件中设置 default-base 字段\n")
			} else {
				fmt.Fprintf(os.Stderr, "请检查配置文件格式是否正确\n")
			}
		} else {
			fmt.Fprintf(os.Stderr, "请检查配置文件格式是否正确\n")
		}
		os.Exit(1)
	}
	return engine.New(cfg)
}

func runCreate(cmd *cobra.Command, args []string) {
	feature := args[0]
	base, _ := cmd.Flags().GetString("base")
	modules, _ := cmd.Flags().GetStringSlice("modules")

	eng := loadConfig()

	// 检查 feature 是否已存在
	featurePath := filepath.Join(eng.Config.WorktreeRoot, feature)
	var existingModules []string
	if _, err := os.Stat(featurePath); err == nil {
		// feature 已存在，只把配置中的模块算作已有模块，避免把 .claude、openspec 等非模块目录纳入增删
		configuredNames := make(map[string]bool)
		for _, m := range eng.Config.Modules {
			configuredNames[m.Name] = true
		}
		fmt.Printf("Feature %s 已存在，以下模块已创建:\n", feature)
		entries, _ := os.ReadDir(featurePath)
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() != ".git" && configuredNames[entry.Name()] {
				existingModules = append(existingModules, entry.Name())
				fmt.Printf("  - %s\n", entry.Name())
			}
		}
		fmt.Println()
	}

	// 记录哪些模块需要删除（已存在但用户取消选中的）
	modulesToDelete := make(map[string]bool)
	for _, name := range existingModules {
		modulesToDelete[name] = true // 默认标记为待删除
	}

	// 如果没有指定 modules
	// 标记是否只删除不创建
	deleteOnly := false

	if len(modules) == 0 {
		if isInteractiveTerminal() {
			// 交互式终端：使用 TUI 选择
			selectedModules, isQuit, err := ui.SelectModules(eng.Config.Modules, existingModules)
			if err != nil {
				// TUI 不可用时回退到非交互模式
				fmt.Fprintf(os.Stderr, "TUI 不可用: %v\n", err)
			} else if isQuit {
				// 用户按 q/ctrl+c 退出，保留已存在的模块，不执行任何操作
				fmt.Println("已取消操作，保留所有已存在的模块")
				return
			} else if len(selectedModules) == 0 && len(existingModules) > 0 {
				// 用户按回车但没有选择任何模块，只删除不创建
				fmt.Println("未选择任何模块，将删除所有已存在的模块（如需保留请按 q 退出）")
				deleteOnly = true
			} else {
				eng.Config.Modules = selectedModules
				// 标记用户选中的模块为不需要删除
				for _, m := range selectedModules {
					delete(modulesToDelete, m.Name)
				}
			}
		} else {
			// 非交互式：自动过滤掉已存在的模块（不删除，只创建新的）
			if len(existingModules) > 0 {
				existingMap := make(map[string]bool)
				for _, name := range existingModules {
					existingMap[name] = true
				}
				var filtered []config.Module
				for _, m := range eng.Config.Modules {
					if !existingMap[m.Name] {
						filtered = append(filtered, m)
					}
				}
				eng.Config.Modules = filtered
				fmt.Printf("将创建以下新模块: ")
				for i, m := range eng.Config.Modules {
					if i > 0 {
						fmt.Print(", ")
					}
					fmt.Print(m.Name)
				}
				fmt.Println()
				// 非交互模式下，不删除任何模块
				modulesToDelete = make(map[string]bool)
			}
		}
	} else if len(modules) > 0 {
		// 命令行指定了 modules，过滤配置
		moduleSet := make(map[string]bool)
		for _, m := range modules {
			moduleSet[m] = true
		}
		var filtered []config.Module
		for _, m := range eng.Config.Modules {
			if moduleSet[m.Name] {
				filtered = append(filtered, m)
			}
		}
		eng.Config.Modules = filtered
	}

	formatter := output.New(outputFmt)

	// 先删除需要移除的模块
	if len(modulesToDelete) > 0 && len(existingModules) > 0 {
		fmt.Println("删除模块:")
		for name := range modulesToDelete {
			modulePath := filepath.Join(featurePath, name)
			repoPath := filepath.Join(eng.Config.Workspace, name)
			// 删除 worktree 和分支
			if err := eng.GitProxy.RemoveWorktreeAndBranch(cmd.Context(), repoPath, feature, modulePath); err != nil {
				fmt.Printf("  - %s: 删除失败 %v\n", name, err)
			} else {
				fmt.Printf("  - %s: 已删除（含分支）\n", name)
			}
		}
		fmt.Println()
	}

	// 如果是删除模式（用户未选择任何模块），则不创建新模块
	if deleteOnly {
		fmt.Println("✓ 操作完成")
		return
	}

	err := eng.CreateWorktree(cmd.Context(), feature, base)
	if err != nil {
		fmt.Print(formatter.FormatError(errors.Code(err), err.Error(), nil))
		os.Exit(1)
	}

	fmt.Print(formatter.FormatCreateResponse(feature, []output.Result{}, nil))
}

func runDelete(cmd *cobra.Command, args []string) {
	feature := args[0]
	force, _ := cmd.Flags().GetBool("force")

	eng := loadConfig()
	formatter := output.New(outputFmt)

	err := eng.DeleteWorktree(cmd.Context(), feature, force)
	if err != nil {
		fmt.Print(formatter.FormatError(errors.Code(err), err.Error(), nil))
		os.Exit(1)
	}

	fmt.Print(formatter.FormatDeleteResponse(feature, nil))
}

func runList(cmd *cobra.Command, args []string) {
	eng := loadConfig()
	formatter := output.New(outputFmt)

	envs, err := eng.ListWorktrees(cmd.Context())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list worktrees: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(formatter.FormatListResponse(envs))
}

func runInfo(cmd *cobra.Command, args []string) {
	feature := args[0]

	eng := loadConfig()
	formatter := output.New(outputFmt)

	env, err := eng.GetWorktreeInfo(cmd.Context(), feature)
	if err != nil {
		fmt.Print(formatter.FormatError(errors.Code(err), err.Error(), nil))
		os.Exit(1)
	}

	fmt.Print(formatter.FormatInfoResponse(env))
}

func runInit(cmd *cobra.Command, args []string) {
	shouldScan, _ := cmd.Flags().GetBool("scan")

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，先尝试扫描
		if shouldScan {
			fmt.Println("配置文件不存在，正在扫描目录自动发现模块...")
			fmt.Println()
			// 创建默认配置用于扫描
			cfg := config.DefaultConfig()
			if err := config.SaveConfig(cfg, configPath); err != nil {
				fmt.Fprintf(os.Stderr, "创建配置文件失败: %v\n", err)
				os.Exit(1)
			}
			runConfigScan(cmd, []string{})
		}

		// 再次检查配置文件是否存在（扫描后可能已创建）
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// 尝试使用交互式向导
			if isInteractiveTerminal() {
				fmt.Println("配置文件不存在，正在启动配置向导...")
				fmt.Println()
				if err := ui.RunConfigWizard(); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				// 配置向导完成后，重新加载配置
				fmt.Println()
				fmt.Println("配置文件已创建，正在初始化仓库...")
			} else {
				// 非交互式环境：创建默认配置并提示用户
				cfg := config.DefaultConfig()
				if err := config.SaveConfig(cfg, configPath); err != nil {
					fmt.Fprintf(os.Stderr, "创建默认配置文件失败: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("已创建默认配置文件: %s\n", configPath)
				fmt.Println()
				fmt.Println("请编辑配置文件添加模块后再次运行:")
				fmt.Printf("  modu config create --module \"name=https://github.com/xxx.git\"\n")
				fmt.Println("或使用交互式向导:")
				fmt.Printf("  modu init\n")
				os.Exit(0)
			}
		}
	}

	// 先用不验证模块的方式加载配置
	cfg, err := config.LoadConfigForScan(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 如果没有模块且用户指定了 --scan，执行扫描
	if !shouldScan && len(cfg.Modules) == 0 {
		// 没有模块，检查是否在交互式环境中
		if isInteractiveTerminal() {
			fmt.Println()
			fmt.Print("未发现任何模块，是否扫描当前目录自动发现模块? [Y/n]: ")
			var input string
			if _, err := fmt.Scanln(&input); err != nil {
				// 读取失败，使用默认值
				input = ""
			}
			input = strings.ToLower(strings.TrimSpace(input))
			if input == "" || input == "y" || input == "yes" {
				shouldScan = true
			}
		} else {
			fmt.Println()
			fmt.Println("提示: 当前没有配置模块，如需自动发现模块请运行:")
			fmt.Printf("  modu init --scan\n")
		}
	}

	if shouldScan && len(cfg.Modules) == 0 {
		fmt.Println()
		fmt.Println("正在扫描目录自动发现模块...")
		runConfigScan(cmd, []string{})
		// 重新加载配置
		cfg, err = config.LoadConfigForScan(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 使用完整配置创建 Engine
	eng := engine.New(cfg)
	formatter := output.New(outputFmt)

	err = eng.Init(cmd.Context())
	if err != nil {
		fmt.Print(formatter.FormatError("ERR_INIT_FAILED", err.Error(), nil))
		os.Exit(1)
	}

	fmt.Println("✓ Initialized all repositories")

	// 更新主项目的 .gitignore，添加模块目录
	if cfg.Workspace != "" && len(cfg.Modules) > 0 {
		// 将 workspace 转换为绝对路径（相对于配置文件）
		workspacePath := cfg.Workspace
		if !filepath.IsAbs(workspacePath) {
			configDir := filepath.Dir(configPath)
			workspacePath = filepath.Join(configDir, workspacePath)
		}
		if err := config.UpdateGitignore(workspacePath, cfg.Modules); err != nil {
			fmt.Fprintf(os.Stderr, "警告: 更新 .gitignore 失败: %v\n", err)
		}
	}
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("modu version %s\n", version)
	fmt.Printf("  commit: %s\n", commit)
	fmt.Printf("  date: %s\n", date)
	fmt.Printf("  go: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func runStatus(cmd *cobra.Command, args []string) {
	eng := loadConfig()
	formatter := output.New(outputFmt)

	envs, err := eng.ListWorktrees(cmd.Context())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get status: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(formatter.FormatListResponse(envs))
}

func runUpdate(cmd *cobra.Command, args []string) {
	eng := loadConfig()

	if len(args) == 0 {
		success, failed := eng.UpdateMainProject(cmd.Context())
		printUpdateResult("", success, failed)
		if len(failed) > 0 {
			os.Exit(1)
		}
		return
	}

	feature := args[0]
	featurePath := filepath.Join(eng.Config.WorktreeRoot, feature)
	if _, err := os.Stat(featurePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "feature %s 不存在: %s\n", feature, featurePath)
		os.Exit(1)
	}
	success, failed := eng.UpdateWorktree(cmd.Context(), feature)
	printUpdateResult(feature, success, failed)
	if len(failed) > 0 {
		os.Exit(1)
	}
}

func printUpdateResult(feature string, success int, failed map[string]error) {
	if len(failed) == 0 {
		if feature == "" {
			if success == 1 {
				fmt.Println("更新成功: 主项目")
			} else {
				fmt.Printf("更新成功: 主项目 + %d 个模块\n", success-1)
			}
		} else {
			fmt.Printf("更新成功: feature %s（主项目 + %d 个模块）\n", feature, success-1)
		}
		return
	}
	names := make([]string, 0, len(failed))
	for name := range failed {
		names = append(names, name)
	}
	fmt.Printf("更新成功: %d 个，失败: %d 个 (%s)\n", success, len(failed), strings.Join(names, ", "))
}

// runConfigCreate 运行配置创建命令
func runConfigCreate(cmd *cobra.Command, args []string) {
	// 检查是否可以使用 TTY
	if isInteractiveTerminal() {
		if err := ui.RunConfigWizard(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	// 非交互式模式：使用命令行参数或默认值创建配置
	workspace, _ := cmd.Flags().GetString("workspace")
	worktreeRoot, _ := cmd.Flags().GetString("worktree-root")
	defaultBase, _ := cmd.Flags().GetString("default-base")
	modules, _ := cmd.Flags().GetStringArray("module")

	cfg := config.DefaultConfig()
	if workspace != "" {
		cfg.Workspace = workspace
	}
	if worktreeRoot != "" {
		cfg.WorktreeRoot = worktreeRoot
	}
	if defaultBase != "" {
		cfg.DefaultBase = defaultBase
	}

	// 解析模块
	for _, m := range modules {
		parts := strings.Split(m, "=")
		if len(parts) == 2 {
			cfg.Modules = append(cfg.Modules, config.Module{
				Name: parts[0],
				URL:  parts[1],
			})
		}
	}

	if err := config.SaveConfig(cfg, configPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("配置文件已创建: %s\n", configPath)
}

// runConfigScan 扫描当前目录自动发现模块并更新配置
func runConfigScan(cmd *cobra.Command, args []string) {
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("配置文件 %s 不存在，请先运行 modu config create\n", configPath)
		os.Exit(1)
	}

	// 加载现有配置
	cfg, err := config.LoadConfigForScan(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 扫描当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取当前目录失败: %v\n", err)
		os.Exit(1)
	}

	newModules, err := config.ScanWorkspace(cmd.Context(), currentDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "扫描目录失败: %v\n", err)
		os.Exit(1)
	}

	if len(newModules) == 0 {
		fmt.Println("未发现新的 git 仓库")
		return
	}

	// 合并模块（按 URL 去重）
	existingURLs := make(map[string]bool)
	for _, m := range cfg.Modules {
		existingURLs[m.URL] = true
	}

	addedCount := 0
	for _, m := range newModules {
		if !existingURLs[m.URL] {
			cfg.Modules = append(cfg.Modules, m)
			existingURLs[m.URL] = true
			addedCount++
		}
	}

	// 保存配置
	if err := config.SaveConfig(cfg, configPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 更新主项目的 .gitignore，添加模块目录
	if cfg.Workspace != "" && addedCount > 0 {
		// 将 workspace 转换为绝对路径（相对于配置文件）
		workspacePath := cfg.Workspace
		if !filepath.IsAbs(workspacePath) {
			configDir := filepath.Dir(configPath)
			workspacePath = filepath.Join(configDir, workspacePath)
		}
		if err := config.UpdateGitignore(workspacePath, cfg.Modules); err != nil {
			fmt.Fprintf(os.Stderr, "警告: 更新 .gitignore 失败: %v\n", err)
		}
	}

	fmt.Printf("✓ 已扫描并添加 %d 个模块\n", addedCount)
}
