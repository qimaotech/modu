package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"codeup.aliyun.com/qimao/public/devops/modu/internal/config"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/core"
	errs "codeup.aliyun.com/qimao/public/devops/modu/internal/errors"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/gitproxy"
)

// Engine 核心业务引擎
type Engine struct {
	Config   *config.Config
	GitProxy gitproxy.GitClient
}

// worktreeResult 创建工作树的结果
type worktreeResult struct {
	module config.Module
	path   string
	err    error
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
	// 1. 前置检查
	featurePath := filepath.Join(e.Config.WorktreeRoot, feature)
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
				results <- worktreeResult{module: module, path: worktreePath, err: nil}
				return nil
			}

			// 使用模块指定的 baseBranch 或全局 defaultBase
			branch := base
			if module.BaseBranch != "" {
				branch = module.BaseBranch
			}

			err := e.GitProxy.CreateWorktree(ctx, repoPath, feature, branch, worktreePath)
			results <- worktreeResult{module: module, path: worktreePath, err: err}
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
			if r.err == nil && r.path != "" {
				_ = e.GitProxy.RemoveWorktree(ctx, r.path)
				_ = os.RemoveAll(r.path)
			}
		}
		// 删除主项目 worktree（feature 目录）
		_ = e.GitProxy.RemoveWorktree(ctx, mainProjectPath)
		_ = os.RemoveAll(mainProjectPath)

		if len(errMsgs) > 0 {
			return fmt.Errorf("create worktree failed: %s: %w", strings.Join(errMsgs, "; "), errs.ErrPartialFailure)
		}
		return fmt.Errorf("create worktree failed: %w", errs.ErrPartialFailure)
	}
	close(results)

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
	featurePath := filepath.Join(e.Config.WorktreeRoot, feature)

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
		for _, entry := range entries {
			if !entry.IsDir() {
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

	// 删除所有模块的 worktree
	for _, module := range e.Config.Modules {
		modulePath := filepath.Join(featurePath, module.Name)
		if err := e.GitProxy.RemoveWorktree(ctx, modulePath); err != nil {
			// 即使失败也继续删除其他模块
			continue
		}
	}

	// 删除 feature 目录
	if err := os.RemoveAll(featurePath); err != nil {
		return fmt.Errorf("failed to remove feature directory: %w", err)
	}
	return nil
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
		feature := entry.Name()
		featurePath := filepath.Join(e.Config.WorktreeRoot, feature)

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

		// 处理所有子目录（模块）
		for _, me := range moduleEntries {
			if !me.IsDir() {
				continue
			}
			modulePath := filepath.Join(featurePath, me.Name())

			// 跳过主项目目录（已处理）
			if me.Name() == mainProjectName {
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

		envs = append(envs, env)
	}

	return envs, nil
}

// GetWorktreeInfo 获取单个 feature 的详情
func (e *Engine) GetWorktreeInfo(ctx context.Context, feature string) (*core.WorktreeEnv, error) {
	featurePath := filepath.Join(e.Config.WorktreeRoot, feature)

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
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// 跳过主项目目录（兼容旧结构）
		if entry.Name() == mainProjectName {
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
