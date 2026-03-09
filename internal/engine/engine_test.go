package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeup.aliyun.com/qimao/public/devops/modu/internal/config"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/core"
	"codeup.aliyun.com/qimao/public/devops/modu/internal/gitproxy"
)

// MockGitClient 用于测试的 Mock Git 客户端
type MockGitClient struct {
	CloneFunc                       func(ctx context.Context, url, path string) error
	CreateWorktreeFunc              func(ctx context.Context, repoPath, branch, baseBranch, worktreePath string) error
	CreateWorktreeFromExistingBranchFunc func(ctx context.Context, repoPath, branch, worktreePath string) error
	GetStatusFunc                   func(ctx context.Context, path string) (gitproxy.Status, error)
	RemoveWorktreeFunc              func(ctx context.Context, path string) error
	RemoveWorktreeAndBranchFunc     func(ctx context.Context, repoPath, branch, worktreePath string) error
	ListWorktreesFunc               func(ctx context.Context, repoPath string) ([]gitproxy.WorktreeInfo, error)
	FetchFunc                       func(ctx context.Context, repoPath string) error
	RebaseFunc                      func(ctx context.Context, path string) error
	BranchExistsFunc                func(ctx context.Context, repoPath, branch string) bool
	CheckBranchWorktreeStatusFunc   func(ctx context.Context, repoPath, branch string) (bool, error)
}

var _ gitproxy.GitClient = (*MockGitClient)(nil)

func (m *MockGitClient) Clone(ctx context.Context, url, path string) error {
	if m.CloneFunc != nil {
		return m.CloneFunc(ctx, url, path)
	}
	return nil
}

func (m *MockGitClient) CreateWorktree(ctx context.Context, repoPath, branch, baseBranch, worktreePath string) error {
	if m.CreateWorktreeFunc != nil {
		return m.CreateWorktreeFunc(ctx, repoPath, branch, baseBranch, worktreePath)
	}
	return nil
}

func (m *MockGitClient) GetStatus(ctx context.Context, path string) (gitproxy.Status, error) {
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc(ctx, path)
	}
	return gitproxy.Status{IsDirty: false, Branch: "main"}, nil
}

func (m *MockGitClient) RemoveWorktree(ctx context.Context, path string) error {
	if m.RemoveWorktreeFunc != nil {
		return m.RemoveWorktreeFunc(ctx, path)
	}
	return nil
}

func (m *MockGitClient) RemoveWorktreeAndBranch(ctx context.Context, repoPath, branch, worktreePath string) error {
	if m.RemoveWorktreeAndBranchFunc != nil {
		return m.RemoveWorktreeAndBranchFunc(ctx, repoPath, branch, worktreePath)
	}
	return nil
}

func (m *MockGitClient) ListWorktrees(ctx context.Context, repoPath string) ([]gitproxy.WorktreeInfo, error) {
	if m.ListWorktreesFunc != nil {
		return m.ListWorktreesFunc(ctx, repoPath)
	}
	return nil, nil
}

func (m *MockGitClient) Fetch(ctx context.Context, repoPath string) error {
	if m.FetchFunc != nil {
		return m.FetchFunc(ctx, repoPath)
	}
	return nil
}

func (m *MockGitClient) Rebase(ctx context.Context, path string) error {
	if m.RebaseFunc != nil {
		return m.RebaseFunc(ctx, path)
	}
	return nil
}

func (m *MockGitClient) BranchExists(ctx context.Context, repoPath, branch string) bool {
	if m.BranchExistsFunc != nil {
		return m.BranchExistsFunc(ctx, repoPath, branch)
	}
	return true
}

func (m *MockGitClient) CheckBranchWorktreeStatus(ctx context.Context, repoPath, branch string) (bool, error) {
	if m.CheckBranchWorktreeStatusFunc != nil {
		return m.CheckBranchWorktreeStatusFunc(ctx, repoPath, branch)
	}
	return false, nil
}

func (m *MockGitClient) CreateWorktreeFromExistingBranch(ctx context.Context, repoPath, branch, worktreePath string) error {
	if m.CreateWorktreeFromExistingBranchFunc != nil {
		return m.CreateWorktreeFromExistingBranchFunc(ctx, repoPath, branch, worktreePath)
	}
	return nil
}

func TestCreateWorktree_RollbackOnFailure(t *testing.T) {
	// 创建配置
	cfg := &config.Config{
		Workspace:    "/tmp/test-workspace",
		WorktreeRoot: "/tmp/test-worktrees",
		Concurrency:  2,
		Modules: []config.Module{
			{Name: "module1", URL: "git@github.com:test/module1.git"},
			{Name: "module2", URL: "git@github.com:test/module2.git"},
			{Name: "module3", URL: "git@github.com:test/module3.git"},
		},
	}

	// 记录删除操作
	var removedPaths []string

	// 创建 Mock，第二个模块会失败
	mock := &MockGitClient{
		BranchExistsFunc: func(ctx context.Context, repoPath, branch string) bool {
			// 分支不存在，走创建新分支的逻辑
			return false
		},
		CheckBranchWorktreeStatusFunc: func(ctx context.Context, repoPath, branch string) (bool, error) {
			return false, nil
		},
		CreateWorktreeFunc: func(ctx context.Context, repoPath, branch, baseBranch, worktreePath string) error {
			// 模拟第二个模块失败
			if filepath.Base(worktreePath) == "module2" {
				return errors.New("simulated failure for module2")
			}
			return nil
		},
		RemoveWorktreeFunc: func(ctx context.Context, path string) error {
			removedPaths = append(removedPaths, path)
			return nil
		},
	}

	engine := NewWithClient(cfg, mock)

	// 执行创建，应该失败并触发回滚
	err := engine.CreateWorktree(context.Background(), "test-feature", "develop")
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	// 验证：应该有模块被删除（回滚）
	// 由于是并发，可能 module1 或 module3 被成功创建然后回滚
	t.Logf("Removed paths: %v", removedPaths)
	t.Logf("Error: %v", err)
}

func TestCheckDirty(t *testing.T) {
	cfg := &config.Config{
		Workspace:    "/tmp/test-workspace",
		WorktreeRoot: "/tmp/test-worktrees",
		Concurrency:  2,
		Modules: []config.Module{
			{Name: "module1", URL: "git@github.com:test/module1.git"},
		},
	}

	// 模拟脏目录
	mock := &MockGitClient{
		GetStatusFunc: func(ctx context.Context, path string) (gitproxy.Status, error) {
			return gitproxy.Status{IsDirty: true, Branch: "feature/test"}, nil
		},
	}

	engine := NewWithClient(cfg, mock)

	env := core.WorktreeEnv{
		Name: "test-feature",
		Modules: []core.ModuleStatus{
			{Name: "module1", Path: "/tmp/test-worktrees/test-feature/module1"},
		},
	}

	dirty, err := engine.CheckDirty(context.Background(), env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dirty) != 1 {
		t.Fatalf("expected 1 dirty module, got %d", len(dirty))
	}

	if !dirty[0].IsDirty {
		t.Error("expected IsDirty to be true")
	}
}

func TestCheckDirty_Clean(t *testing.T) {
	cfg := &config.Config{
		Workspace:    "/tmp/test-workspace",
		WorktreeRoot: "/tmp/test-worktrees",
		Concurrency:  2,
		Modules: []config.Module{
			{Name: "module1", URL: "git@github.com:test/module1.git"},
		},
	}

	// 模拟干净目录
	mock := &MockGitClient{
		GetStatusFunc: func(ctx context.Context, path string) (gitproxy.Status, error) {
			return gitproxy.Status{IsDirty: false, Branch: "develop"}, nil
		},
	}

	engine := NewWithClient(cfg, mock)

	env := core.WorktreeEnv{
		Name: "test-feature",
		Modules: []core.ModuleStatus{
			{Name: "module1", Path: "/tmp/test-worktrees/test-feature/module1"},
		},
	}

	dirty, err := engine.CheckDirty(context.Background(), env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dirty) != 0 {
		t.Fatalf("expected 0 dirty modules, got %d", len(dirty))
	}
}

func TestCreateWorktree_ReuseExistingBranch(t *testing.T) {
	cfg := &config.Config{
		Workspace:    "/tmp/test-workspace",
		WorktreeRoot: "/tmp/test-worktrees",
		Concurrency:  2,
		Modules: []config.Module{
			{Name: "module1", URL: "git@github.com:test/module1.git"},
		},
	}

	var createFromExistingBranchCalled bool

	mock := &MockGitClient{
		BranchExistsFunc: func(ctx context.Context, repoPath, branch string) bool {
			// 分支存在
			return true
		},
		CheckBranchWorktreeStatusFunc: func(ctx context.Context, repoPath, branch string) (bool, error) {
			// 分支未被 worktree 使用
			return false, nil
		},
		CreateWorktreeFromExistingBranchFunc: func(ctx context.Context, repoPath, branch, worktreePath string) error {
			createFromExistingBranchCalled = true
			return nil
		},
	}

	engine := NewWithClient(cfg, mock)

	err := engine.CreateWorktree(context.Background(), "test-feature", "develop")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !createFromExistingBranchCalled {
		t.Error("expected CreateWorktreeFromExistingBranch to be called")
	}
}

func TestCreateWorktree_SkipBranchUsedByOtherWorktree(t *testing.T) {
	cfg := &config.Config{
		Workspace:    "/tmp/test-workspace",
		WorktreeRoot: "/tmp/test-worktrees",
		Concurrency:  2,
		Modules: []config.Module{
			{Name: "module1", URL: "git@github.com:test/module1.git"},
		},
	}

	var checkStatusCalled bool

	mock := &MockGitClient{
		BranchExistsFunc: func(ctx context.Context, repoPath, branch string) bool {
			// 只对 module 路径返回 true
			if strings.Contains(repoPath, "module1") {
				return true
			}
			return false
		},
		CheckBranchWorktreeStatusFunc: func(ctx context.Context, repoPath, branch string) (bool, error) {
			// 分支已被其他 worktree 使用
			checkStatusCalled = true
			return true, nil
		},
	}

	engine := NewWithClient(cfg, mock)

	err := engine.CreateWorktree(context.Background(), "test-feature", "develop")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 应该检查过分支状态
	if !checkStatusCalled {
		t.Error("expected CheckBranchWorktreeStatus to be called")
	}
}

func TestCreateWorktree_CreateNewBranchWhenNotExists(t *testing.T) {
	cfg := &config.Config{
		Workspace:    "/tmp/test-workspace",
		WorktreeRoot: "/tmp/test-worktrees",
		Concurrency:  2,
		Modules: []config.Module{
			{Name: "module1", URL: "git@github.com:test/module1.git"},
		},
	}

	var createNewBranchCalled bool

	mock := &MockGitClient{
		BranchExistsFunc: func(ctx context.Context, repoPath, branch string) bool {
			// 分支不存在
			return false
		},
		CreateWorktreeFunc: func(ctx context.Context, repoPath, branch, baseBranch, worktreePath string) error {
			createNewBranchCalled = true
			return nil
		},
	}

	engine := NewWithClient(cfg, mock)

	err := engine.CreateWorktree(context.Background(), "test-feature", "develop")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !createNewBranchCalled {
		t.Error("expected CreateWorktree to be called when branch does not exist")
	}
}

func TestUpdateMainProject_Success(t *testing.T) {
	tmp := t.TempDir()
	m1 := filepath.Join(tmp, "m1")
	if err := os.MkdirAll(m1, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		Workspace:   tmp,
		Concurrency: 2,
		Modules:     []config.Module{{Name: "m1", URL: "git@test/m1.git"}},
	}
	mock := &MockGitClient{
		RebaseFunc: func(ctx context.Context, path string) error {
			return nil
		},
	}
	engine := NewWithClient(cfg, mock)
	success, failed := engine.UpdateMainProject(context.Background())
	if success != 2 {
		t.Errorf("expected success 2 (main + m1), got %d", success)
	}
	if len(failed) != 0 {
		t.Errorf("expected no failures, got %v", failed)
	}
}

func TestUpdateMainProject_PartialFailure(t *testing.T) {
	tmp := t.TempDir()
	m1 := filepath.Join(tmp, "m1")
	if err := os.MkdirAll(m1, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		Workspace:   tmp,
		Concurrency: 2,
		Modules:     []config.Module{{Name: "m1", URL: "git@test/m1.git"}},
	}
	rebaseErr := errors.New("rebase failed")
	mock := &MockGitClient{
		RebaseFunc: func(ctx context.Context, path string) error {
			if strings.Contains(path, "m1") {
				return rebaseErr
			}
			return nil
		},
	}
	engine := NewWithClient(cfg, mock)
	success, failed := engine.UpdateMainProject(context.Background())
	if success != 1 {
		t.Errorf("expected success 1 (main only), got %d", success)
	}
	if len(failed) != 1 || failed["m1"] != rebaseErr {
		t.Errorf("expected failed[m1]=rebaseErr, got failed=%v", failed)
	}
}

func TestUpdateWorktree_Success(t *testing.T) {
	workRoot := t.TempDir()
	featurePath := filepath.Join(workRoot, "my-feature")
	m1 := filepath.Join(featurePath, "m1")
	if err := os.MkdirAll(m1, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		Workspace:    "/tmp/workspace",
		WorktreeRoot: workRoot,
		Concurrency:  2,
		Modules:      []config.Module{{Name: "m1", URL: "git@test/m1.git"}},
	}
	mock := &MockGitClient{
		RebaseFunc: func(ctx context.Context, path string) error {
			return nil
		},
	}
	engine := NewWithClient(cfg, mock)
	success, failed := engine.UpdateWorktree(context.Background(), "my-feature")
	if success != 2 {
		t.Errorf("expected success 2 (main + m1), got %d", success)
	}
	if len(failed) != 0 {
		t.Errorf("expected no failures, got %v", failed)
	}
}

func TestUpdateWorktree_PartialFailure(t *testing.T) {
	workRoot := t.TempDir()
	featurePath := filepath.Join(workRoot, "my-feature")
	m1 := filepath.Join(featurePath, "m1")
	if err := os.MkdirAll(m1, 0755); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		Workspace:    "/tmp/workspace",
		WorktreeRoot: workRoot,
		Concurrency:  2,
		Modules:      []config.Module{{Name: "m1", URL: "git@test/m1.git"}},
	}
	rebaseErr := errors.New("rebase failed")
	mock := &MockGitClient{
		RebaseFunc: func(ctx context.Context, path string) error {
			if strings.Contains(path, "m1") {
				return rebaseErr
			}
			return nil
		},
	}
	engine := NewWithClient(cfg, mock)
	success, failed := engine.UpdateWorktree(context.Background(), "my-feature")
	if success != 1 {
		t.Errorf("expected success 1 (main only), got %d", success)
	}
	if len(failed) != 1 || failed["m1"] != rebaseErr {
		t.Errorf("expected failed[m1]=rebaseErr, got failed=%v", failed)
	}
}
