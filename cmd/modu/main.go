package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"codeup.aliyun.com/qimao/public/devops/modu/internal/config"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/engine"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/errors"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/output"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/ui"

	"github.com/spf13/cobra"
)

// isInteractiveTerminal 检查是否是交互式终端
func isInteractiveTerminal() bool {
	// 检查 stdin 是否为 TTY
	cmd := exec.CommandContext(context.Background(), "test", "-t", "0")
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
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
	configCreateCmd.Flags().String("workspace", "./workspace", "仓库目录")
	configCreateCmd.Flags().String("worktree-root", "./worktrees", "Worktree 存放目录")
	configCreateCmd.Flags().String("default-base", "develop", "默认基准分支")
	configCreateCmd.Flags().StringArray("module", []string{}, "模块 (格式: name=url)")
	configCreateCmd.Flags().BoolP("scan", "s", false, "扫描当前目录自动发现模块")
	configCmd.AddCommand(configCreateCmd)

	// init 命令 - 初始化仓库（克隆所有配置的仓库）
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "初始化仓库（克隆所有配置的仓库）",
		Run:   runInit,
	}

	// status 命令
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "显示所有模块的脏状态",
		Run:   runStatus,
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

	rootCmd.AddCommand(createCmd, deleteCmd, listCmd, infoCmd, configCmd, initCmd, statusCmd, tuiCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadConfig() *engine.Engine {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	return engine.New(cfg)
}

func runCreate(cmd *cobra.Command, args []string) {
	feature := args[0]
	base, _ := cmd.Flags().GetString("base")

	eng := loadConfig()
	formatter := output.New(outputFmt)

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
	eng := loadConfig()
	formatter := output.New(outputFmt)

	err := eng.Init(cmd.Context())
	if err != nil {
		fmt.Print(formatter.FormatError("ERR_INIT_FAILED", err.Error(), nil))
		os.Exit(1)
	}

	fmt.Println("✓ Initialized all repositories")
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

// runConfigCreate 运行配置创建命令
func runConfigCreate(cmd *cobra.Command, args []string) {
	scan, _ := cmd.Flags().GetBool("scan")

	// 检查是否可以使用 TTY（只有在非 scan 模式时）
	if isInteractiveTerminal() && !scan {
		if err := ui.RunConfigWizard(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	// scan 模式下检查配置文件是否已存在
	if scan {
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("配置文件 %s 已存在，是否覆盖? (y/N): ", configPath)
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				fmt.Println("已取消")
				os.Exit(0)
			}
		}
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

	// 如果是 scan 模式，扫描当前目录
	if scan {
		runConfigCreateWithScan(cmd, cfg)
		return
	}

	if err := config.SaveConfig(cfg, configPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("配置文件已创建: %s\n", configPath)
}

// runConfigCreateWithScan 创建配置后扫描当前目录自动发现模块
func runConfigCreateWithScan(cmd *cobra.Command, cfg *config.Config) {
	// 先保存初始配置
	if err := config.SaveConfig(cfg, configPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("配置文件已创建: %s\n", configPath)

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

	fmt.Printf("✓ 已扫描并添加 %d 个模块\n", addedCount)
}
