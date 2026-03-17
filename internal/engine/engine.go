package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/qimaotech/modu/internal/config"
	"github.com/qimaotech/modu/internal/core"
	errs "github.com/qimaotech/modu/internal/errors"
	"github.com/qimaotech/modu/internal/gitproxy"
	"github.com/qimaotech/modu/internal/logger"
)

// Engine 核心业务引擎
type Engine struct {
	Config   *config.Config
	GitProxy gitproxy.GitClient
}

// MainProjectStatus 主项目（workspace 根仓库）状态
type MainProjectStatus struct {
	Name    string
	Path    string
	IsDirty bool
	Branch  string
}

// worktreeResult 创建工作树的结果
type worktreeResult struct {
	module   config.Module
	path     string
	repoPath string
	err      error
	skipped  bool   // 是否跳过
	skipMsg  string // 跳过原因
}

// New 创建引擎
func New(cfg *config.Config) *Engine {
	return &Engine{
		Config:   cfg,
		GitProxy: gitproxy.New(),
	}
}

// NewWithClient 创建引擎（带自定义 GitClient，用于测试）
func NewWithClient(cfg *config.Config, client gitproxy.GitClient) *Engine {
	return &Engine{
		Config:   cfg,
		GitProxy: client,
	}
}

// featureToDirName 将 feature 名转换为目录名
// feature/hello → feature-hello
func featureToDirName(feature string) string {
	return strings.ReplaceAll(feature, "/", "-")
}

// Init 并发克隆所有配置的仓库
func (e *Engine) Init(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(e.Config.Concurrency)

	for _, module := range e.Config.Modules {
		module := module
		g.Go(func() error {
			path := filepath.Join(e.Config.Workspace, module.Name)
			// 检查是否已存在
			if _, err := os.Stat(path); err == nil {
				return nil // 已存在，跳过
			}
			return e.GitProxy.Clone(ctx, module.URL, path)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to init repositories: %w", fmt.Errorf("%w", err))
	}
	return nil
}

// CreateWorktree 并发创建 feature 工作树
func (e *Engine) CreateWorktree(ctx context.Context, feature, base string) error {
	logger.Info("开始创建 feature: %s, base: %s", feature, base)

	// 将 feature 名转换为目录名（feature/hello → feature-hello）
	dirName := featureToDirName(feature)

	// 1. 前置检查
	featurePath := filepath.Join(e.Config.WorktreeRoot, dirName)
	featureExists := false
	if _, err := os.Stat(featurePath); err == nil {
		featureExists = true
	}

	// 如果 feature 不存在，确保 worktree-root 目录存在
	if !featureExists {
		if err := os.MkdirAll(e.Config.WorktreeRoot, 0755); err != nil {
			return fmt.Errorf("failed to create worktree root: %w", err)
		}
	}

	// 2. 主项目直接放在 feature 目录下，不需要再创建子目录
	mainProjectPath := featurePath

	// 3. 如果 feature 不存在，先创建主项目的 worktree
	if !featureExists {
		err := e.GitProxy.CreateWorktree(ctx, e.Config.Workspace, feature, base, mainProjectPath)
		if err != nil {
			return fmt.Errorf("failed to create worktree for main project: %w", err)
		}
	}

	// 4. 并发创建其他模块的 worktree
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(e.Config.Concurrency)

	results := make(chan worktreeResult, len(e.Config.Modules))

	// 为每个模块创建 worktree（跳过已存在的）
	for _, module := range e.Config.Modules {
		module := module
		g.Go(func() error {
			repoPath := filepath.Join(e.Config.Workspace, module.Name)
			worktreePath := filepath.Join(featurePath, module.Name)

			// 检查模块是否已存在
			if _, err := os.Stat(worktreePath); err == nil {
				// 模块已存在，跳过
				results <- worktreeResult{module: module, path: worktreePath, repoPath: repoPath, err: nil}
				return nil
			}

			// 检查分支是否已存在
			branchExists := e.GitProxy.BranchExists(ctx, repoPath, feature)

			if branchExists {
				// 分支已存在，检查是否被其他 worktree 使用
				isUsed, err := e.GitProxy.CheckBranchWorktreeStatus(ctx, repoPath, feature)
				if err != nil {
					results <- worktreeResult{module: module, path: worktreePath, repoPath: repoPath, err: err}
					return fmt.Errorf("failed to check branch worktree status for %s: %w", module.Name, err)
				}

				if isUsed {
					// 分支已被其他 worktree 使用，跳过
					skipMsg := fmt.Sprintf("分支 %s 已被其他 worktree 使用", feature)
					logger.Info("[SKIP] %s: %s", module.Name, skipMsg)
					results <- worktreeResult{module: module, path: worktreePath, repoPath: repoPath, err: nil, skipped: true, skipMsg: skipMsg}
					return nil
				}

				// 分支存在但未被使用，复用现有分支
				logger.Info("复用现有分支 %s 创建 worktree: module=%s", feature, module.Name)
				err = e.GitProxy.CreateWorktreeFromExistingBranch(ctx, repoPath, feature, worktreePath)
				results <- worktreeResult{module: module, path: worktreePath, repoPath: repoPath, err: err}
				if err != nil {
					return fmt.Errorf("failed to create worktree for %s: %w", module.Name, err)
				}
				return nil
			}

			// 分支不存在，创建新分支
			// 使用模块指定的 baseBranch 或全局 defaultBase
			branch := base
			if module.BaseBranch != "" {
				branch = module.BaseBranch
			}

			err := e.GitProxy.CreateWorktree(ctx, repoPath, feature, branch, worktreePath)
			results <- worktreeResult{module: module, path: worktreePath, repoPath: repoPath, err: err}
			if err != nil {
				return fmt.Errorf("failed to create worktree for %s: %w", module.Name, err)
			}
			return nil
		})
	}

	// 等待所有任务完成
	if err := g.Wait(); err != nil {
		// 3. 失败回滚：先收集所有结果到切片，避免 channel 已关闭导致的问题
		close(results)
		var resultSlice []worktreeResult
		for r := range results {
			resultSlice = append(resultSlice, r)
		}

		// 收集所有错误信息
		var errMsgs []string
		for _, r := range resultSlice {
			if r.err != nil {
				errMsgs = append(errMsgs, fmt.Sprintf("%s: %v", r.module.Name, r.err))
			}
		}

		// 回滚：删除已创建的模块 worktree
		for _, r := range resultSlice {
			if r.err == nil && r.path != "" && r.repoPath != "" {
				_ = e.GitProxy.RemoveWorktreeAndBranch(ctx, r.repoPath, feature, r.path)
				_ = os.RemoveAll(r.path)
			}
		}
		// 删除主项目 worktree（feature 目录）
		_ = e.GitProxy.RemoveWorktreeAndBranch(ctx, e.Config.Workspace, feature, mainProjectPath)
		_ = os.RemoveAll(mainProjectPath)

		if len(errMsgs) > 0 {
			return fmt.Errorf("create worktree failed: %s: %w", strings.Join(errMsgs, "; "), errs.ErrPartialFailure)
		}
		return fmt.Errorf("create worktree failed: %w", errs.ErrPartialFailure)
	}
	close(results)

	// 收集结果统计
	var successCount, skipCount int
	var skipModules []string
	var resultSlice []worktreeResult
	for r := range results {
		resultSlice = append(resultSlice, r)
		if r.skipped {
			skipCount++
			skipModules = append(skipModules, r.module.Name)
		} else if r.err == nil && r.path != "" {
			successCount++
		}
	}

	// 输出 summary
	if skipCount > 0 {
		logger.Info("创建成功: %d 个模块，跳过: %d 个模块 (%s)", successCount, skipCount, strings.Join(skipModules, ", "))
	} else {
		logger.Info("创建成功: %d 个模块", successCount)
	}

	logger.Info("成功创建 feature: %s", feature)

	// 创建 VSCode workspace 文件（使用目录名）
	if err := e.createVSCodeWorkspace(dirName, featurePath); err != nil {
		logger.Warn("创建 VSCode workspace 文件失败: %v", err)
	}

	return nil
}

// vscodeWorkspace VSCode workspace 配置结构
type vscodeWorkspace struct {
	Folder    []folder    `json:"folders"`
	Settings  settings    `json:"settings"`
	Extensions extensions `json:"extensions"`
}

type folder struct {
	Path string `json:"path"`
}

type settings struct {
	GoToolsManagementAutoUpdate bool            `json:"go.toolsManagement.autoUpdate"`
	GoLintTool                 string           `json:"go.lintTool"`
	GoLintOnSave               string           `json:"go.lintOnSave"`
	GoFormatTool               string           `json:"go.formatTool"`
	GoUseLanguageServer        bool             `json:"go.useLanguageServer"`
	GoAlternateTools           map[string]any  `json:"go.alternateTools"`
}

type extensions struct {
	Recommendations []string `json:"recommendations"`
}

// createVSCodeWorkspace 创建 VSCode workspace 文件
func (e *Engine) createVSCodeWorkspace(feature, featurePath string) error {
	workspaceFile := filepath.Join(featurePath, feature+".code-workspace")

	// 构建 folders 数组：只包含 feature 中实际存在的模块
	folders := make([]folder, 0, 8)

	// 获取配置的模块名称集合
	configuredModules := make(map[string]bool)
	for _, m := range e.Config.Modules {
		configuredModules[m.Name] = true
	}

	// 扫描 feature 目录，只添加实际存在的模块
	entries, err := os.ReadDir(featurePath)
	if err != nil {
		return fmt.Errorf("failed to read feature directory: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// 跳过非模块目录（如 .git, .claude 等）
		if !configuredModules[entry.Name()] {
			continue
		}
		folders = append(folders, folder{Path: entry.Name()})
	}

	workspace := vscodeWorkspace{
		Folder: folders,
		Settings: settings{
			GoToolsManagementAutoUpdate: true,
			GoLintTool:                 "golangci-lint",
			GoLintOnSave:               "package",
			GoFormatTool:               "gofmt",
			GoUseLanguageServer:        true,
			GoAlternateTools: map[string]any{
				"go": "/usr/local/go/bin/go",
			},
		},
		Extensions: extensions{
			Recommendations: []string{
				"golang.go",
				"vue.volar",
				"ms-vscode.vscode-typescript-next",
			},
		},
	}

	data, err := json.MarshalIndent(workspace, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workspace: %w", err)
	}

	if err := os.WriteFile(workspaceFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write workspace file: %w", err)
	}

	logger.Info("创建 VSCode workspace 文件: %s", workspaceFile)
	return nil
}

// CheckDirty 检查环境是否存在未提交修改
func (e *Engine) CheckDirty(ctx context.Context, env core.WorktreeEnv) ([]core.ModuleStatus, error) {
	var dirty []core.ModuleStatus

	for _, module := range env.Modules {
		status, err := e.GitProxy.GetStatus(ctx, module.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to get status for %s: %w", module.Path, err)
		}

		if status.IsDirty {
			dirty = append(dirty, core.ModuleStatus{
				Name:    module.Name,
				Path:    module.Path,
				IsDirty: true,
				Branch:  status.Branch,
			})
		}
	}

	return dirty, nil
}

// DeleteWorktree 删除 feature 工作树
func (e *Engine) DeleteWorktree(ctx context.Context, feature string, force bool) error {
	logger.Info("开始删除 feature: %s, force: %v", feature, force)

	// 将 feature 名转换为目录名（feature/hello → feature-hello）
	dirName := featureToDirName(feature)
	featurePath := filepath.Join(e.Config.WorktreeRoot, dirName)

	// 检查是否存在
	if _, err := os.Stat(featurePath); os.IsNotExist(err) {
		return fmt.Errorf("feature %s not found: %w", feature, errs.ErrFeatureNotFound)
	}

	// 脏检查
	if !force && e.Config.StrictDirty {
		env := core.WorktreeEnv{
			Name: feature,
		}
		// 遍历 feature 目录下的所有模块
		entries, err := os.ReadDir(featurePath)
		if err != nil {
			return fmt.Errorf("failed to read feature directory: %w", err)
		}
		configuredNames := e.configuredModuleNames()
		for _, entry := range entries {
			if !entry.IsDir() || !configuredNames[entry.Name()] {
				continue
			}
			modulePath := filepath.Join(featurePath, entry.Name())
			env.Modules = append(env.Modules, core.ModuleStatus{
				Name: entry.Name(),
				Path: modulePath,
			})
		}

		dirty, err := e.CheckDirty(ctx, env)
		if err != nil {
			return err
		}
		if len(dirty) > 0 {
			return fmt.Errorf("cannot delete: uncommitted changes detected in %v: %w", dirty, errs.ErrDirtyWorktree)
		}
	}

	// 获取 feature 目录下实际存在的模块
	entries, err := os.ReadDir(featurePath)
	if err != nil {
		return fmt.Errorf("failed to read feature directory: %w", err)
	}

	// 删除所有存在的模块的 worktree
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".git" {
			continue
		}
		moduleName := entry.Name()
		modulePath := filepath.Join(featurePath, moduleName)

		// 检查是否在配置的 modules 中
		isConfiguredModule := false
		for _, m := range e.Config.Modules {
			if m.Name == moduleName {
				isConfiguredModule = true
				break
			}
		}

		if !isConfiguredModule {
			// 不在配置中，跳过不删除
			logger.Info("目录 %s 不在配置中，跳过", moduleName)
			continue
		}

		repoPath := filepath.Join(e.Config.Workspace, moduleName)

		// 检查仓库是否存在
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			logger.Warn("模块仓库不存在，跳过: module=%s, repo=%s", moduleName, repoPath)
			// 仓库不存在，只删除目录
			if err := os.RemoveAll(modulePath); err != nil {
				logger.Error("删除模块目录失败: %s, error=%v", moduleName, err)
			}
			continue
		}

		logger.Info("删除模块 worktree: module=%s, repo=%s, branch=%s, path=%s", moduleName, repoPath, feature, modulePath)
		if err := e.GitProxy.RemoveWorktreeAndBranch(ctx, repoPath, feature, modulePath); err != nil {
			logger.Error("删除模块 worktree 失败: module=%s, error=%v", moduleName, err)
			continue
		}
		logger.Info("删除模块 worktree 成功: %s", moduleName)
	}

	// 删除主项目的 worktree（主项目直接放在 feature 目录下）
	mainProjectPath := featurePath

	// 检查主项目仓库是否存在
	if _, err := os.Stat(e.Config.Workspace); os.IsNotExist(err) {
		logger.Warn("主项目仓库不存在，跳过: repo=%s", e.Config.Workspace)
	} else {
		logger.Info("删除主项目 worktree: repo=%s, branch=%s, path=%s", e.Config.Workspace, feature, mainProjectPath)
		if err := e.GitProxy.RemoveWorktreeAndBranch(ctx, e.Config.Workspace, feature, mainProjectPath); err != nil {
			logger.Error("删除主项目 worktree 失败: error=%v", err)
			fmt.Printf("Warning: failed to remove main project worktree: %v\n", err)
		} else {
			logger.Info("删除主项目 worktree 成功")
		}
	}

	// 删除 feature 目录
	if err := os.RemoveAll(featurePath); err != nil {
		logger.Error("删除 feature 目录失败: %s, error: %v", feature, err)
		return fmt.Errorf("failed to remove feature directory: %w", err)
	}

	logger.Info("成功删除 feature: %s", feature)
	return nil
}

// GetMainProject 获取主项目（workspace）状态，若路径无效或非 git 仓库则返回 nil 与 nil error
func (e *Engine) GetMainProject(ctx context.Context) (*MainProjectStatus, error) {
	if e.Config.Workspace == "" {
		return nil, nil
	}
	if _, err := os.Stat(e.Config.Workspace); err != nil {
		return nil, nil
	}
	status, err := e.GitProxy.GetStatus(ctx, e.Config.Workspace)
	if err != nil {
		return nil, nil
	}
	name := filepath.Base(e.Config.Workspace)
	return &MainProjectStatus{
		Name:    name,
		Path:    e.Config.Workspace,
		IsDirty: status.IsDirty,
		Branch:  status.Branch,
	}, nil
}

// GetMainProjectModules 获取主项目及其所有模块的分支状态
func (e *Engine) GetMainProjectModules(ctx context.Context) (*MainProjectStatus, []core.ModuleStatus, error) {
	main, err := e.GetMainProject(ctx)
	if err != nil || main == nil {
		return nil, nil, err
	}

	modules := make([]core.ModuleStatus, 0, len(e.Config.Modules))
	for _, module := range e.Config.Modules {
		repoPath := filepath.Join(e.Config.Workspace, module.Name)
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			continue
		}
		status, err := e.GitProxy.GetStatus(ctx, repoPath)
		if err != nil {
			continue
		}
		modules = append(modules, core.ModuleStatus{
			Name:    module.Name,
			Path:    repoPath,
			IsDirty: status.IsDirty,
			Branch:  status.Branch,
		})
	}

	return main, modules, nil
}

// UpdateMainProject 并发对主项目和所有模块执行 fetch + rebase，返回成功数量和失败 map[name]error
func (e *Engine) UpdateMainProject(ctx context.Context) (success int, failed map[string]error) {
	failed = make(map[string]error)
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(e.Config.Concurrency)

	mainName := filepath.Base(e.Config.Workspace)
	g.Go(func() error {
		// 主项目切换到 default-base 分支
		if err := e.GitProxy.FetchAndSwitchBranch(ctx, e.Config.Workspace, e.Config.DefaultBase); err != nil {
			failed[mainName] = err
		}
		return nil
	})

	for _, module := range e.Config.Modules {
		module := module
		repoPath := filepath.Join(e.Config.Workspace, module.Name)
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			continue
		}
		// 使用模块的 base-branch（如果有）或全局 default-base
		baseBranch := module.BaseBranch
		if baseBranch == "" {
			baseBranch = e.Config.DefaultBase
		}
		g.Go(func() error {
			if err := e.GitProxy.FetchAndSwitchBranch(ctx, repoPath, baseBranch); err != nil {
				failed[module.Name] = err
			}
			return nil
		})
	}

	_ = g.Wait()

	if _, ok := failed[mainName]; !ok {
		success++
	}
	for _, module := range e.Config.Modules {
		repoPath := filepath.Join(e.Config.Workspace, module.Name)
		if _, err := os.Stat(repoPath); err == nil {
			if _, hasErr := failed[module.Name]; !hasErr {
				success++
			}
		}
	}
	return success, failed
}

// UpdateWorktree 对指定 feature 的 worktree（主项目 + 该目录下存在的模块）并发执行 fetch + rebase
func (e *Engine) UpdateWorktree(ctx context.Context, feature string) (success int, failed map[string]error) {
	failed = make(map[string]error)
	// 将 feature 名转换为目录名（feature/hello → feature-hello）
	dirName := featureToDirName(feature)
	featurePath := filepath.Join(e.Config.WorktreeRoot, dirName)
	if _, err := os.Stat(featurePath); os.IsNotExist(err) {
		return 0, failed
	}
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(e.Config.Concurrency)

	mainName := filepath.Base(e.Config.Workspace)
	g.Go(func() error {
		if err := e.GitProxy.Rebase(ctx, featurePath); err != nil {
			failed[mainName] = err
		}
		return nil
	})

	for _, module := range e.Config.Modules {
		module := module
		modulePath := filepath.Join(featurePath, module.Name)
		if _, err := os.Stat(modulePath); os.IsNotExist(err) {
			continue
		}
		g.Go(func() error {
			if err := e.GitProxy.Rebase(ctx, modulePath); err != nil {
				failed[module.Name] = err
			}
			return nil
		})
	}

	_ = g.Wait()

	if _, ok := failed[mainName]; !ok {
		success++
	}
	for _, module := range e.Config.Modules {
		modulePath := filepath.Join(featurePath, module.Name)
		if _, err := os.Stat(modulePath); err == nil {
			if _, hasErr := failed[module.Name]; !hasErr {
				success++
			}
		}
	}
	return success, failed
}

// ListWorktrees 列出所有 worktree
func (e *Engine) ListWorktrees(ctx context.Context) ([]core.WorktreeEnv, error) {
	// 扫描 worktree-root 目录，如果不存在则创建
	if err := os.MkdirAll(e.Config.WorktreeRoot, 0755); err != nil {
		return nil, fmt.Errorf("failed to create worktree root: %w", err)
	}

	entries, err := os.ReadDir(e.Config.WorktreeRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read worktree root: %w", err)
	}

	// 获取主项目名称
	mainProjectName := filepath.Base(e.Config.Workspace)

	envs := make([]core.WorktreeEnv, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 跳过以 . 开头的隐藏目录
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// 直接使用目录名作为 feature 名
		feature := entry.Name()
		featurePath := filepath.Join(e.Config.WorktreeRoot, entry.Name())

		env := core.WorktreeEnv{
			Name:        feature,
			Base:        "", // TODO: 需要从 git 获取
			MainProject: nil,
			Modules:     []core.ModuleStatus{},
		}

		// 列出该 feature 下的所有子目录
		moduleEntries, err := os.ReadDir(featurePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read feature directory %s: %w", featurePath, err)
		}

		// 1. 检查 feature 目录本身是否是主项目（新结构）
		// feature 目录 = 主项目，feature 下的子目录 = 模块
		status, err := e.GitProxy.GetStatus(ctx, featurePath)
		if err == nil {
			env.MainProject = &core.ModuleStatus{
				Name:    mainProjectName,
				Path:    featurePath,
				IsDirty: status.IsDirty,
				Branch:  status.Branch,
			}
		} else {
			// 2. 检查是否有 workspace 子目录（兼容旧结构）
			for _, me := range moduleEntries {
				if me.Name() == mainProjectName {
					modulePath := filepath.Join(featurePath, me.Name())
					status, err := e.GitProxy.GetStatus(ctx, modulePath)
					if err == nil {
						env.MainProject = &core.ModuleStatus{
							Name:    mainProjectName,
							Path:    modulePath,
							IsDirty: status.IsDirty,
							Branch:  status.Branch,
						}
					}
					break
				}
			}
		}

		// 只把配置中的模块加入列表，避免 .claude、openspec 等非模块目录被当作模块展示
		configuredNames := e.configuredModuleNames()
		for _, me := range moduleEntries {
			if !me.IsDir() {
				continue
			}
			modulePath := filepath.Join(featurePath, me.Name())

			// 跳过主项目目录（已处理）
			if me.Name() == mainProjectName {
				continue
			}
			if !configuredNames[me.Name()] {
				continue
			}

			// 普通模块
			status, err := e.GitProxy.GetStatus(ctx, modulePath)
			if err != nil {
				env.Modules = append(env.Modules, core.ModuleStatus{
					Name:  me.Name(),
					Path:  modulePath,
					Error: err,
				})
				continue
			}
			env.Modules = append(env.Modules, core.ModuleStatus{
				Name:    me.Name(),
				Path:    modulePath,
				IsDirty: status.IsDirty,
				Branch:  status.Branch,
			})
		}

		// 跳过没有主项目的目录（无效的 feature）
		if env.MainProject == nil {
			continue
		}

		envs = append(envs, env)
	}

	return envs, nil
}

// GetWorktreeInfo 获取单个 feature 的详情
func (e *Engine) GetWorktreeInfo(ctx context.Context, feature string) (*core.WorktreeEnv, error) {
	// 将 feature 名转换为目录名（feature/hello → feature-hello）
	dirName := featureToDirName(feature)
	featurePath := filepath.Join(e.Config.WorktreeRoot, dirName)

	if _, err := os.Stat(featurePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("feature %s not found: %w", feature, errs.ErrFeatureNotFound)
	}

	env := &core.WorktreeEnv{
		Name:        feature,
		MainProject: nil,
		Modules:     []core.ModuleStatus{},
	}

	// 获取主项目名称
	mainProjectName := filepath.Base(e.Config.Workspace)

	// 检查 feature 目录本身是否是主项目
	status, err := e.GitProxy.GetStatus(ctx, featurePath)
	if err == nil {
		env.MainProject = &core.ModuleStatus{
			Name:    mainProjectName,
			Path:    featurePath,
			IsDirty: status.IsDirty,
			Branch:  status.Branch,
		}
	}

	// 列出所有模块（子目录）
	entries, err := os.ReadDir(featurePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read feature directory %s: %w", featurePath, err)
	}
	configuredNames := e.configuredModuleNames()
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// 跳过主项目目录（兼容旧结构）
		if entry.Name() == mainProjectName {
			continue
		}
		if !configuredNames[entry.Name()] {
			continue
		}
		modulePath := filepath.Join(featurePath, entry.Name())
		status, err := e.GitProxy.GetStatus(ctx, modulePath)
		if err != nil {
			env.Modules = append(env.Modules, core.ModuleStatus{
				Name:  entry.Name(),
				Path:  modulePath,
				Error: err,
			})
			continue
		}
		env.Modules = append(env.Modules, core.ModuleStatus{
			Name:    entry.Name(),
			Path:    modulePath,
			IsDirty: status.IsDirty,
			Branch:  status.Branch,
		})
	}

	return env, nil
}

// configuredModuleNames 返回配置中模块名称集合，用于区分「配置内模块」与普通目录（如 .claude、openspec）
func (e *Engine) configuredModuleNames() map[string]bool {
	out := make(map[string]bool, len(e.Config.Modules))
	for _, m := range e.Config.Modules {
		out[m.Name] = true
	}
	return out
}

// AddModule 为 feature 添加单个模块的 worktree
func (e *Engine) AddModule(ctx context.Context, feature, moduleName string) error {
	logger.Info("为 feature %s 添加模块: %s", feature, moduleName)

	// 查找模块配置
	var module config.Module
	found := false
	for _, m := range e.Config.Modules {
		if m.Name == moduleName {
			module = m
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("module %s not found in config", moduleName)
	}

	// 将 feature 名转换为目录名（feature/hello → feature-hello）
	dirName := featureToDirName(feature)
	featurePath := filepath.Join(e.Config.WorktreeRoot, dirName)

	// 检查 feature 是否存在
	if _, err := os.Stat(featurePath); os.IsNotExist(err) {
		return fmt.Errorf("feature %s not found: %w", feature, errs.ErrFeatureNotFound)
	}

	// 检查模块是否已存在
	modulePath := filepath.Join(featurePath, moduleName)
	if _, err := os.Stat(modulePath); err == nil {
		return fmt.Errorf("module %s already exists in feature %s", moduleName, feature)
	}

	// 创建模块的 worktree
	repoPath := filepath.Join(e.Config.Workspace, moduleName)

	// 检查仓库是否存在
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repository for module %s not found at %s", moduleName, repoPath)
	}

	// 使用模块指定的 baseBranch 或全局 defaultBase
	branch := e.Config.DefaultBase
	if module.BaseBranch != "" {
		branch = module.BaseBranch
	}

	// 检查分支是否已存在
	branchExists := e.GitProxy.BranchExists(ctx, repoPath, feature)

	if branchExists {
		// 分支已存在，检查是否被其他 worktree 使用
		isUsed, err := e.GitProxy.CheckBranchWorktreeStatus(ctx, repoPath, feature)
		if err != nil {
			return fmt.Errorf("failed to check branch worktree status: %w", err)
		}

		if isUsed {
			// 分支已被其他 worktree 使用，跳过
			logger.Info("[SKIP] %s: 分支 %s 已被其他 worktree 使用", moduleName, feature)
			return nil
		}

		// 分支存在但未被使用，复用现有分支
		logger.Info("复用现有分支 %s 添加模块: module=%s", feature, moduleName)
		if err := e.GitProxy.CreateWorktreeFromExistingBranch(ctx, repoPath, feature, modulePath); err != nil {
			return fmt.Errorf("failed to create worktree for %s: %w", moduleName, err)
		}
		logger.Info("成功为 feature %s 添加模块: %s", feature, moduleName)

		// 更新 VSCode workspace 文件（使用目录名）
		if err := e.createVSCodeWorkspace(dirName, featurePath); err != nil {
			logger.Warn("更新 VSCode workspace 文件失败: %v", err)
		}
		return nil
	}

	// 分支不存在，创建新分支
	if err := e.GitProxy.CreateWorktree(ctx, repoPath, feature, branch, modulePath); err != nil {
		return fmt.Errorf("failed to create worktree for %s: %w", moduleName, err)
	}

	logger.Info("成功为 feature %s 添加模块: %s", feature, moduleName)

	// 更新 VSCode workspace 文件（使用目录名）
	if err := e.createVSCodeWorkspace(dirName, featurePath); err != nil {
		logger.Warn("更新 VSCode workspace 文件失败: %v", err)
	}
	return nil
}

// RemoveModule 为 feature 删除单个模块的 worktree
func (e *Engine) RemoveModule(ctx context.Context, feature, moduleName string) error {
	logger.Info("为 feature %s 删除模块: %s", feature, moduleName)

	// 将 feature 名转换为目录名（feature/hello → feature-hello）
	dirName := featureToDirName(feature)
	featurePath := filepath.Join(e.Config.WorktreeRoot, dirName)

	// 检查 feature 是否存在
	if _, err := os.Stat(featurePath); os.IsNotExist(err) {
		return fmt.Errorf("feature %s not found: %w", feature, errs.ErrFeatureNotFound)
	}

	modulePath := filepath.Join(featurePath, moduleName)

	// 检查模块是否存在
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		return fmt.Errorf("module %s not found in feature %s", moduleName, feature)
	}

	// 脏检查
	if e.Config.StrictDirty {
		status, err := e.GitProxy.GetStatus(ctx, modulePath)
		if err == nil && status.IsDirty {
			return fmt.Errorf("cannot remove module %s: uncommitted changes detected: %w", moduleName, errs.ErrDirtyWorktree)
		}
	}

	// 删除模块的 worktree
	repoPath := filepath.Join(e.Config.Workspace, moduleName)

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		// 仓库不存在，只删除目录
		logger.Warn("模块仓库不存在，只删除目录: module=%s", moduleName)
		if err := os.RemoveAll(modulePath); err != nil {
			return fmt.Errorf("failed to remove module directory: %w", err)
		}
		logger.Info("成功删除模块目录: feature=%s, module=%s", feature, moduleName)
		return nil
	}

	if err := e.GitProxy.RemoveWorktreeAndBranch(ctx, repoPath, feature, modulePath); err != nil {
		return fmt.Errorf("failed to remove worktree for %s: %w", moduleName, err)
	}

	// 删除模块目录
	if err := os.RemoveAll(modulePath); err != nil {
		logger.Warn("删除模块目录失败: %s, error=%v", moduleName, err)
	}

	logger.Info("成功为 feature %s 删除模块: %s", feature, moduleName)

	// 更新 VSCode workspace 文件（使用目录名）
	if err := e.createVSCodeWorkspace(dirName, featurePath); err != nil {
		logger.Warn("更新 VSCode workspace 文件失败: %v", err)
	}
	return nil
}
