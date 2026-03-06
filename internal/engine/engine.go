package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
	if _, err := os.Stat(featurePath); err == nil {
		return fmt.Errorf("feature %s already exists at %s: %w", feature, featurePath, errs.ErrFeatureExists)
	}

	// 创建 feature 目录
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		return fmt.Errorf("failed to create feature directory: %w", err)
	}

	// 2. 并发执行
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(e.Config.Concurrency)

	results := make(chan worktreeResult, len(e.Config.Modules))

	for _, module := range e.Config.Modules {
		module := module
		g.Go(func() error {
			repoPath := filepath.Join(e.Config.Workspace, module.Name)
			worktreePath := filepath.Join(featurePath, module.Name)

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
		e.rollback(ctx, resultSlice, featurePath)
		return fmt.Errorf("create worktree failed: %w", errs.ErrPartialFailure)
	}
	close(results)

	return nil
}

// rollback 回滚已创建的工作树
func (e *Engine) rollback(ctx context.Context, results []worktreeResult, featurePath string) {
	for _, r := range results {
		if r.err == nil && r.path != "" {
			// 尝试删除已创建的 worktree
			_ = e.GitProxy.RemoveWorktree(ctx, r.path)
			// 也删除物理目录
			_ = os.RemoveAll(r.path)
		}
	}
	// 删除 feature 目录
	_ = os.RemoveAll(featurePath)
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
	// 扫描 worktree-root 目录
	entries, err := os.ReadDir(e.Config.WorktreeRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read worktree root: %w", err)
	}

	envs := make([]core.WorktreeEnv, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		feature := entry.Name()
		featurePath := filepath.Join(e.Config.WorktreeRoot, feature)

		env := core.WorktreeEnv{
			Name:    feature,
			Base:    "", // TODO: 需要从 git 获取
			Modules: []core.ModuleStatus{},
		}

		// 列出该 feature 下的所有模块
		moduleEntries, err := os.ReadDir(featurePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read feature directory %s: %w", featurePath, err)
		}
		for _, me := range moduleEntries {
			if !me.IsDir() {
				continue
			}
			modulePath := filepath.Join(featurePath, me.Name())
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
		Name:    feature,
		Modules: []core.ModuleStatus{},
	}

	// 列出所有模块
	entries, err := os.ReadDir(featurePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read feature directory %s: %w", featurePath, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
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
