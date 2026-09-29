package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/qimaotech/modu/internal/core"
	errs "github.com/qimaotech/modu/internal/errors"
	"golang.org/x/sync/errgroup"
)

func canonicalPath(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return filepath.Clean(path)
}

type WorkspaceSummary struct {
	Envs        []core.WorktreeEnv
	MainProject *MainProjectStatus
	Modules     []core.ModuleStatus
}

// ListWorkspaceSummaries 每个源仓库只读取一次 worktree 注册信息，不扫描工作区文件。
func (e *Engine) ListWorkspaceSummaries(ctx context.Context) (WorkspaceSummary, error) {
	summary := WorkspaceSummary{Envs: []core.WorktreeEnv{}, Modules: []core.ModuleStatus{}}
	if err := os.MkdirAll(e.Config.WorktreeRoot, 0755); err != nil {
		return summary, err
	}
	entries, err := os.ReadDir(e.Config.WorktreeRoot)
	if err != nil {
		return summary, err
	}
	repositories := []string{e.Config.Workspace}
	for _, module := range e.Config.Modules {
		repositories = append(repositories, filepath.Join(e.Config.Workspace, module.Name))
	}
	branches := make(map[string]string)
	var mu sync.Mutex
	var group errgroup.Group
	group.SetLimit(e.concurrency())
	for _, repoPath := range repositories {
		group.Go(func() error {
			if ctx.Err() != nil {
				return nil
			}
			worktrees, err := e.GitProxy.ListWorktrees(ctx, repoPath)
			if err != nil {
				return nil
			} // 配置允许尚未初始化的模块。
			mu.Lock()
			defer mu.Unlock()
			for _, worktree := range worktrees {
				branches[canonicalPath(worktree.Path)] = worktree.Branch
			}
			return nil
		})
	}
	_ = group.Wait()
	if err := ctx.Err(); err != nil {
		return summary, err
	}
	mainName := filepath.Base(e.Config.Workspace)
	if branch, ok := branches[canonicalPath(e.Config.Workspace)]; ok {
		summary.MainProject = &MainProjectStatus{Name: mainName, Path: e.Config.Workspace, Branch: branch}
	}
	for _, module := range e.Config.Modules {
		path := filepath.Join(e.Config.Workspace, module.Name)
		if branch, ok := branches[canonicalPath(path)]; ok {
			summary.Modules = append(summary.Modules, core.ModuleStatus{Name: module.Name, Path: path, Branch: branch})
		}
	}
	envs := make([]core.WorktreeEnv, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		featurePath := filepath.Join(e.Config.WorktreeRoot, entry.Name())
		mainPath := featurePath
		branch, exists := branches[canonicalPath(mainPath)]
		if !exists {
			mainPath = filepath.Join(featurePath, mainName)
			branch, exists = branches[canonicalPath(mainPath)]
		}
		if !exists {
			continue
		}
		env := core.WorktreeEnv{Name: entry.Name(), DirName: entry.Name(), Branch: branch, Modules: []core.ModuleStatus{},
			MainProject: &core.ModuleStatus{Name: mainName, Path: mainPath, Branch: branch}}
		for _, module := range e.Config.Modules {
			modulePath := filepath.Join(featurePath, module.Name)
			if branch, ok := branches[canonicalPath(modulePath)]; ok {
				if info, err := os.Stat(modulePath); err == nil && info.IsDir() {
					env.Modules = append(env.Modules, core.ModuleStatus{Name: module.Name, Path: modulePath, Branch: branch})
				}
			}
		}
		sort.Slice(env.Modules, func(i, j int) bool { return env.Modules[i].Name < env.Modules[j].Name })
		envs = append(envs, env)
	}
	summary.Envs = envs
	return summary, nil
}

func (e *Engine) ListWorktreeSummaries(ctx context.Context) ([]core.WorktreeEnv, error) {
	summary, err := e.ListWorkspaceSummaries(ctx)
	return summary.Envs, err
}

func (e *Engine) ListWorktrees(ctx context.Context) ([]core.WorktreeEnv, error) {
	envs, err := e.ListWorktreeSummaries(ctx)
	if err != nil {
		return nil, err
	}
	return e.RefreshWorktreeStatuses(ctx, envs)
}

// RefreshWorktreeStatuses 拷贝输入快照，避免后台查询与 TUI 渲染共享可变切片。
func (e *Engine) RefreshWorktreeStatuses(ctx context.Context, envs []core.WorktreeEnv) ([]core.WorktreeEnv, error) {
	result := make([]core.WorktreeEnv, len(envs))
	var targets []*core.ModuleStatus
	for i, env := range envs {
		result[i] = env
		result[i].Modules = append([]core.ModuleStatus{}, env.Modules...)
		if env.MainProject != nil {
			main := *env.MainProject
			result[i].MainProject = &main
			targets = append(targets, &main)
		}
		for j := range result[i].Modules {
			targets = append(targets, &result[i].Modules[j])
		}
	}
	e.inspectStatuses(ctx, targets)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for i := range result {
		if result[i].MainProject != nil {
			result[i].Branch = result[i].MainProject.Branch
		}
	}
	return result, nil
}

func (e *Engine) inspectStatuses(ctx context.Context, targets []*core.ModuleStatus) {
	var group errgroup.Group
	group.SetLimit(e.concurrency())
	for _, target := range targets {
		group.Go(func() error {
			if err := ctx.Err(); err != nil {
				target.Error = err
				return nil
			}
			status, err := e.GitProxy.GetStatus(ctx, target.Path)
			target.Error = err
			if err == nil {
				target.Branch, target.IsDirty = status.Branch, status.IsDirty
			}
			return nil
		})
	}
	_ = group.Wait()
}

// GetWorktreeInfo 只检查指定环境；模块状态使用与列表相同的并发上限。
func (e *Engine) GetWorktreeInfo(ctx context.Context, feature string) (*core.WorktreeEnv, error) {
	dirName := featureToDirName(feature)
	featurePath := filepath.Join(e.Config.WorktreeRoot, dirName)
	if _, err := os.Stat(featurePath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("feature %s: %w", feature, errs.ErrFeatureNotFound)
		}
		return nil, fmt.Errorf("feature %s: %w", feature, err)
	}
	mainPath := featurePath
	if _, err := os.Stat(filepath.Join(mainPath, ".git")); os.IsNotExist(err) {
		legacyPath := filepath.Join(featurePath, filepath.Base(e.Config.Workspace))
		if _, err := os.Stat(filepath.Join(legacyPath, ".git")); err == nil {
			mainPath = legacyPath
		}
	}
	env := core.WorktreeEnv{Name: feature, DirName: dirName, Modules: []core.ModuleStatus{},
		MainProject: &core.ModuleStatus{Name: filepath.Base(e.Config.Workspace), Path: mainPath}}
	for _, module := range e.Config.Modules {
		path := filepath.Join(featurePath, module.Name)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			env.Modules = append(env.Modules, core.ModuleStatus{Name: module.Name, Path: path})
		}
	}
	envs, err := e.RefreshWorktreeStatuses(ctx, []core.WorktreeEnv{env})
	if err != nil {
		return nil, err
	}
	return &envs[0], nil
}
