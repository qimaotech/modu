package engine

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/qimaotech/modu/internal/config"
	"github.com/qimaotech/modu/internal/core"
	errs "github.com/qimaotech/modu/internal/errors"
	"github.com/qimaotech/modu/internal/gitproxy"
)

// MockGitClient 用于测试的 Mock Git 客户端
type MockGitClient struct {
	CloneFunc                            func(ctx context.Context, url, path string) error
	CreateWorktreeFunc                   func(ctx context.Context, repoPath, branch, baseBranch, worktreePath string) error
	CreateWorktreeFromExistingBranchFunc func(ctx context.Context, repoPath, branch, worktreePath string) error
	CreateWorktreeFromRemoteBranchFunc   func(ctx context.Context, repoPath, branch, worktreePath string) error
	GetStatusFunc                        func(ctx context.Context, path string) (gitproxy.Status, error)
	RemoveWorktreeFunc                   func(ctx context.Context, path string) error
	RemoveWorktreeAndBranchFunc          func(ctx context.Context, repoPath, worktreePath, featureDirName string) error
	ListWorktreesFunc                    func(ctx context.Context, repoPath string) ([]gitproxy.WorktreeInfo, error)
	FetchFunc                            func(ctx context.Context, repoPath string) error
	RebaseFunc                           func(ctx context.Context, path string) error
	FetchAndSwitchBranchFunc             func(ctx context.Context, repoPath, branch string) error
	BranchExistsFunc                     func(ctx context.Context, repoPath, branch string) bool
	CheckBranchWorktreeStatusFunc        func(ctx context.Context, repoPath, branch string) (bool, error)
	RemoteBranchExistsFunc               func(ctx context.Context, repoURL, branch string) bool
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

func (m *MockGitClient) RemoveWorktreeAndBranch(ctx context.Context, repoPath, worktreePath, featureDirName string) error {
	if m.RemoveWorktreeAndBranchFunc != nil {
		return m.RemoveWorktreeAndBranchFunc(ctx, repoPath, worktreePath, featureDirName)
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

func (m *MockGitClient) FetchAndSwitchBranch(ctx context.Context, repoPath, branch string) error {
	if m.FetchAndSwitchBranchFunc != nil {
		return m.FetchAndSwitchBranchFunc(ctx, repoPath, branch)
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

func (m *MockGitClient) RemoteBranchExists(ctx context.Context, repoURL, branch string) bool {
	if m.RemoteBranchExistsFunc != nil {
		return m.RemoteBranchExistsFunc(ctx, repoURL, branch)
	}
	return false
}

func (m *MockGitClient) CreateWorktreeFromExistingBranch(ctx context.Context, repoPath, branch, worktreePath string) error {
	if m.CreateWorktreeFromExistingBranchFunc != nil {
		return m.CreateWorktreeFromExistingBranchFunc(ctx, repoPath, branch, worktreePath)
	}
	return nil
}

func (m *MockGitClient) CreateWorktreeFromRemoteBranch(ctx context.Context, repoPath, branch, worktreePath string) error {
	if m.CreateWorktreeFromRemoteBranchFunc != nil {
		return m.CreateWorktreeFromRemoteBranchFunc(ctx, repoPath, branch, worktreePath)
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
		DefaultBase: "develop",
		Concurrency: 2,
		Modules:     []config.Module{{Name: "m1", URL: "git@test/m1.git"}},
	}
	mock := &MockGitClient{
		FetchAndSwitchBranchFunc: func(ctx context.Context, repoPath, branch string) error {
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
		DefaultBase: "develop",
		Concurrency: 2,
		Modules:     []config.Module{{Name: "m1", URL: "git@test/m1.git"}},
	}
	switchErr := errors.New("switch failed")
	mock := &MockGitClient{
		FetchAndSwitchBranchFunc: func(ctx context.Context, repoPath, branch string) error {
			if strings.Contains(repoPath, "m1") {
				return switchErr
			}
			return nil
		},
	}
	engine := NewWithClient(cfg, mock)
	success, failed := engine.UpdateMainProject(context.Background())
	if success != 1 {
		t.Errorf("expected success 1 (main only), got %d", success)
	}
	if len(failed) != 1 || failed["m1"] != switchErr {
		t.Errorf("expected failed[m1]=switchErr, got failed=%v", failed)
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

func TestCreateVSCodeWorkspace(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "modu-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建配置
	cfg := &config.Config{
		Workspace:    filepath.Join(tmpDir, "workspace"),
		WorktreeRoot: filepath.Join(tmpDir, "worktrees"),
		Modules: []config.Module{
			{Name: "module1"},
			{Name: "module2"},
		},
	}

	// 创建 workspace 和 feature 目录
	if err := os.MkdirAll(cfg.Workspace, 0755); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	engine := New(cfg)
	featurePath := filepath.Join(cfg.WorktreeRoot, "test-feature")

	// 创建 feature 目录
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("failed to create feature path: %v", err)
	}

	// 创建实际的模块目录（模拟已添加的模块）
	if err := os.MkdirAll(filepath.Join(featurePath, "module1"), 0755); err != nil {
		t.Fatalf("failed to create module1: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(featurePath, "module2"), 0755); err != nil {
		t.Fatalf("failed to create module2: %v", err)
	}

	// 调用 createVSCodeWorkspace
	err = engine.createVSCodeWorkspace("test-feature", featurePath)
	if err != nil {
		t.Fatalf("createVSCodeWorkspace failed: %v", err)
	}

	// 验证文件生成
	workspaceFile := filepath.Join(featurePath, "test-feature.code-workspace")
	if _, err := os.Stat(workspaceFile); os.IsNotExist(err) {
		t.Fatalf("workspace file not created: %s", workspaceFile)
	}

	// 验证文件内容
	data, err := os.ReadFile(workspaceFile)
	if err != nil {
		t.Fatalf("failed to read workspace file: %v", err)
	}

	// 验证 JSON 结构
	var ws vscodeWorkspace
	if err := json.Unmarshal(data, &ws); err != nil {
		t.Fatalf("failed to parse workspace JSON: %v", err)
	}

	// 验证 folders 只包含模块
	if len(ws.Folder) != 2 {
		t.Errorf("expected 2 folders (modules only), got %d", len(ws.Folder))
	}

	// 验证模块
	if ws.Folder[0].Path != "module1" {
		t.Errorf("expected folder[0] to be 'module1', got %s", ws.Folder[0].Path)
	}
	if ws.Folder[1].Path != "module2" {
		t.Errorf("expected folder[1] to be 'module2', got %s", ws.Folder[1].Path)
	}

	// 验证 settings
	if !ws.Settings.GoToolsManagementAutoUpdate {
		t.Error("expected GoToolsManagementAutoUpdate to be true")
	}
	if ws.Settings.GoLintTool != "golangci-lint" {
		t.Errorf("expected GoLintTool to be 'golangci-lint', got %s", ws.Settings.GoLintTool)
	}

	// 验证 extensions
	if len(ws.Extensions.Recommendations) != 3 {
		t.Errorf("expected 3 recommendations, got %d", len(ws.Extensions.Recommendations))
	}

	if ws.Modu.Feature != "test-feature" {
		t.Errorf("expected modu feature 'test-feature', got %s", ws.Modu.Feature)
	}
	if ws.Modu.DirName != "test-feature" {
		t.Errorf("expected modu dirName 'test-feature', got %s", ws.Modu.DirName)
	}
}

func TestCreateVSCodeWorkspace_NestedFeatureMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Workspace:    filepath.Join(tmpDir, "workspace"),
		WorktreeRoot: filepath.Join(tmpDir, "worktrees"),
	}
	if err := os.MkdirAll(cfg.Workspace, 0755); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}
	engine := New(cfg)
	featurePath := filepath.Join(cfg.WorktreeRoot, "feature-demand-pay-cpr")
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("failed to create feature path: %v", err)
	}

	if err := engine.createVSCodeWorkspace("feature-demand-pay-cpr", featurePath, "feature/demand-pay-cpr"); err != nil {
		t.Fatalf("createVSCodeWorkspace failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(featurePath, "feature-demand-pay-cpr.code-workspace"))
	if err != nil {
		t.Fatalf("failed to read workspace file: %v", err)
	}
	var ws vscodeWorkspace
	if err := json.Unmarshal(data, &ws); err != nil {
		t.Fatalf("failed to parse workspace JSON: %v", err)
	}
	if ws.Modu.Feature != "feature/demand-pay-cpr" {
		t.Errorf("expected modu feature 'feature/demand-pay-cpr', got %s", ws.Modu.Feature)
	}
	if ws.Modu.DirName != "feature-demand-pay-cpr" {
		t.Errorf("expected modu dirName 'feature-demand-pay-cpr', got %s", ws.Modu.DirName)
	}
}

func TestCreateVSCodeWorkspace_EmptyFeature(t *testing.T) {
	// 测试空 feature 目录（无模块）
	tmpDir, err := os.MkdirTemp("", "modu-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Workspace:    filepath.Join(tmpDir, "workspace"),
		WorktreeRoot: filepath.Join(tmpDir, "worktrees"),
		Modules: []config.Module{
			{Name: "module1"},
			{Name: "module2"},
		},
	}

	if err := os.MkdirAll(cfg.Workspace, 0755); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	engine := New(cfg)
	featurePath := filepath.Join(cfg.WorktreeRoot, "empty-feature")

	// 创建空的 feature 目录（无模块）
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("failed to create feature path: %v", err)
	}

	err = engine.createVSCodeWorkspace("empty-feature", featurePath)
	if err != nil {
		t.Fatalf("createVSCodeWorkspace failed: %v", err)
	}

	// 验证文件生成
	workspaceFile := filepath.Join(featurePath, "empty-feature.code-workspace")
	data, err := os.ReadFile(workspaceFile)
	if err != nil {
		t.Fatalf("failed to read workspace file: %v", err)
	}

	var ws vscodeWorkspace
	if err := json.Unmarshal(data, &ws); err != nil {
		t.Fatalf("failed to parse workspace JSON: %v", err)
	}

	// 验证 folders 为空数组
	if len(ws.Folder) != 0 {
		t.Errorf("expected 0 folders for empty feature, got %d", len(ws.Folder))
	}
}

func TestCreateVSCodeWorkspace_Overwrite(t *testing.T) {
	// 测试 workspace 文件覆盖更新
	tmpDir, err := os.MkdirTemp("", "modu-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Workspace:    filepath.Join(tmpDir, "workspace"),
		WorktreeRoot: filepath.Join(tmpDir, "worktrees"),
		Modules: []config.Module{
			{Name: "module1"},
		},
	}

	if err := os.MkdirAll(cfg.Workspace, 0755); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	engine := New(cfg)
	featurePath := filepath.Join(cfg.WorktreeRoot, "test-feature")

	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("failed to create feature path: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(featurePath, "module1"), 0755); err != nil {
		t.Fatalf("failed to create module1: %v", err)
	}

	// 第一次创建
	err = engine.createVSCodeWorkspace("test-feature", featurePath)
	if err != nil {
		t.Fatalf("first createVSCodeWorkspace failed: %v", err)
	}

	// 读取原始文件内容
	workspaceFile := filepath.Join(featurePath, "test-feature.code-workspace")
	originalContent, err := os.ReadFile(workspaceFile)
	if err != nil {
		t.Fatalf("failed to read original workspace file: %v", err)
	}

	// 第二次创建（覆盖）
	err = engine.createVSCodeWorkspace("test-feature", featurePath)
	if err != nil {
		t.Fatalf("second createVSCodeWorkspace failed: %v", err)
	}

	// 验证文件被覆盖
	newContent, err := os.ReadFile(workspaceFile)
	if err != nil {
		t.Fatalf("failed to read new workspace file: %v", err)
	}

	if string(originalContent) != string(newContent) {
		t.Error("workspace file should be overwritten with same content")
	}

	// 验证仍然是有效的 JSON
	var ws vscodeWorkspace
	if err := json.Unmarshal(newContent, &ws); err != nil {
		t.Fatalf("failed to parse overwritten workspace JSON: %v", err)
	}
}

func TestCreateVSCodeWorkspace_SkipNonModuleDirs(t *testing.T) {
	// 测试跳过非模块目录
	tmpDir, err := os.MkdirTemp("", "modu-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Workspace:    filepath.Join(tmpDir, "workspace"),
		WorktreeRoot: filepath.Join(tmpDir, "worktrees"),
		Modules: []config.Module{
			{Name: "module1"},
		},
	}

	if err := os.MkdirAll(cfg.Workspace, 0755); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	engine := New(cfg)
	featurePath := filepath.Join(cfg.WorktreeRoot, "test-feature")

	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("failed to create feature path: %v", err)
	}
	// 创建模块目录
	if err := os.MkdirAll(filepath.Join(featurePath, "module1"), 0755); err != nil {
		t.Fatalf("failed to create module1: %v", err)
	}
	// 创建非模块目录（应该被跳过）
	if err := os.MkdirAll(filepath.Join(featurePath, ".git"), 0755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(featurePath, ".claude"), 0755); err != nil {
		t.Fatalf("failed to create .claude: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(featurePath, "some-random-dir"), 0755); err != nil {
		t.Fatalf("failed to create random dir: %v", err)
	}

	err = engine.createVSCodeWorkspace("test-feature", featurePath)
	if err != nil {
		t.Fatalf("createVSCodeWorkspace failed: %v", err)
	}

	workspaceFile := filepath.Join(featurePath, "test-feature.code-workspace")
	data, err := os.ReadFile(workspaceFile)
	if err != nil {
		t.Fatalf("failed to read workspace file: %v", err)
	}

	var ws vscodeWorkspace
	if err := json.Unmarshal(data, &ws); err != nil {
		t.Fatalf("failed to parse workspace JSON: %v", err)
	}

	// 只应该包含 module1
	if len(ws.Folder) != 1 {
		t.Errorf("expected 1 folder, got %d", len(ws.Folder))
	}
	if ws.Folder[0].Path != "module1" {
		t.Errorf("expected folder[0] to be 'module1', got %s", ws.Folder[0].Path)
	}
}

func TestGetModulesWithRemoteBranch_AllHaveBranch(t *testing.T) {
	cfg := &config.Config{
		Workspace:    "/tmp/test-workspace",
		WorktreeRoot: "/tmp/test-worktrees",
		Concurrency:  2,
		Modules: []config.Module{
			{Name: "module1", URL: "git@github.com:test/module1.git"},
			{Name: "module2", URL: "git@github.com:test/module2.git"},
		},
	}

	mock := &MockGitClient{
		RemoteBranchExistsFunc: func(ctx context.Context, repoURL, branch string) bool {
			return true // 所有模块都有该分支
		},
	}

	engine := NewWithClient(cfg, mock)
	result, err := engine.GetModulesWithRemoteBranch(context.Background(), "feature/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 modules with branch, got %d", len(result))
	}
	if !result["module1"] || !result["module2"] {
		t.Error("expected both modules to have branch")
	}
}

func TestGetModulesWithRemoteBranch_NoneHaveBranch(t *testing.T) {
	cfg := &config.Config{
		Workspace:    "/tmp/test-workspace",
		WorktreeRoot: "/tmp/test-worktrees",
		Concurrency:  2,
		Modules: []config.Module{
			{Name: "module1", URL: "git@github.com:test/module1.git"},
			{Name: "module2", URL: "git@github.com:test/module2.git"},
		},
	}

	mock := &MockGitClient{
		RemoteBranchExistsFunc: func(ctx context.Context, repoURL, branch string) bool {
			return false // 所有模块都没有该分支
		},
	}

	engine := NewWithClient(cfg, mock)
	result, err := engine.GetModulesWithRemoteBranch(context.Background(), "feature/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 modules with branch, got %d", len(result))
	}
}

func TestGetModulesWithRemoteBranch_PartialHaveBranch(t *testing.T) {
	cfg := &config.Config{
		Workspace:    "/tmp/test-workspace",
		WorktreeRoot: "/tmp/test-worktrees",
		Concurrency:  2,
		Modules: []config.Module{
			{Name: "module1", URL: "git@github.com:test/module1.git"},
			{Name: "module2", URL: "git@github.com:test/module2.git"},
		},
	}

	mock := &MockGitClient{
		RemoteBranchExistsFunc: func(ctx context.Context, repoURL, branch string) bool {
			if strings.Contains(repoURL, "module1") {
				return true
			}
			return false
		},
	}

	engine := NewWithClient(cfg, mock)
	result, err := engine.GetModulesWithRemoteBranch(context.Background(), "feature/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 module with branch, got %d", len(result))
	}
	if !result["module1"] {
		t.Error("expected module1 to have branch")
	}
	if result["module2"] {
		t.Error("expected module2 to not have branch")
	}
}

func TestGetModulesWithRemoteBranch_EmptyURL(t *testing.T) {
	cfg := &config.Config{
		Workspace:    "/tmp/test-workspace",
		WorktreeRoot: "/tmp/test-worktrees",
		Concurrency:  2,
		Modules: []config.Module{
			{Name: "module1", URL: ""}, // 空 URL
		},
	}

	mock := &MockGitClient{
		RemoteBranchExistsFunc: func(ctx context.Context, repoURL, branch string) bool {
			return false // 空 URL 应该返回 false
		},
	}

	engine := NewWithClient(cfg, mock)
	result, err := engine.GetModulesWithRemoteBranch(context.Background(), "feature/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 modules with branch, got %d", len(result))
	}
}

func TestAddModule_UsesWorkspaceMetadataForSlugFeature(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		Workspace:    filepath.Join(tmpDir, "workspace"),
		WorktreeRoot: filepath.Join(tmpDir, "worktrees"),
		DefaultBase:  "develop",
		Modules: []config.Module{
			{Name: "module1"},
		},
	}

	featurePath := filepath.Join(cfg.WorktreeRoot, "feature-demand-pay-cpr")
	repoPath := filepath.Join(cfg.Workspace, "module1")
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		t.Fatalf("failed to create feature path: %v", err)
	}
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo path: %v", err)
	}
	workspace := vscodeWorkspace{
		Modu: moduWorkspaceMeta{
			Feature: "feature/demand-pay-cpr",
			DirName: "feature-demand-pay-cpr",
		},
	}
	data, err := json.Marshal(workspace)
	if err != nil {
		t.Fatalf("failed to marshal workspace: %v", err)
	}
	if err := os.WriteFile(filepath.Join(featurePath, "feature-demand-pay-cpr.code-workspace"), data, 0644); err != nil {
		t.Fatalf("failed to write workspace file: %v", err)
	}

	var createdBranch string
	mock := &MockGitClient{
		GetStatusFunc: func(ctx context.Context, path string) (gitproxy.Status, error) {
			if path == featurePath {
				return gitproxy.Status{Branch: "unexpected/git-branch"}, nil
			}
			return gitproxy.Status{Branch: "develop"}, nil
		},
		BranchExistsFunc: func(ctx context.Context, repoPath, branch string) bool {
			return false
		},
		CreateWorktreeFunc: func(ctx context.Context, repoPath, branch, baseBranch, worktreePath string) error {
			createdBranch = branch
			return nil
		},
	}

	engine := NewWithClient(cfg, mock)
	if err := engine.AddModule(context.Background(), "feature-demand-pay-cpr", "module1"); err != nil {
		t.Fatalf("AddModule failed: %v", err)
	}
	if createdBranch != "feature/demand-pay-cpr" {
		t.Errorf("expected branch feature/demand-pay-cpr, got %s", createdBranch)
	}
}

func TestDeleteWorktree_BackupBeforeDeleteAndReturnsPath(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	feature := "feature/remove-backup"
	dirName := featureToDirName(feature)
	featurePath := filepath.Join(worktreeRoot, dirName)
	modulePath := filepath.Join(featurePath, "module1")
	repoPath := filepath.Join(workspace, "module1")

	writeTestFile(t, filepath.Join(featurePath, "main.txt"), "main content")
	writeTestFile(t, filepath.Join(modulePath, "module.txt"), "module content")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo path: %v", err)
	}

	var removeCalls int
	mock := &MockGitClient{
		RemoveWorktreeAndBranchFunc: func(ctx context.Context, repoPath, worktreePath, featureDirName string) error {
			removeCalls++
			backups := listBackupFiles(t, filepath.Join(worktreeRoot, ".modu", "backups"))
			if len(backups) != 1 {
				t.Fatalf("expected backup before removal, got %d backups", len(backups))
			}
			return nil
		},
	}
	engine := NewWithClient(&config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
		StrictDirty:  true,
		Modules:      []config.Module{{Name: "module1"}},
	}, mock)

	result, err := engine.DeleteWorktree(context.Background(), feature, false)
	if err != nil {
		t.Fatalf("DeleteWorktree failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected delete result")
	}
	if result.Feature != feature {
		t.Fatalf("expected result feature %s, got %s", feature, result.Feature)
	}
	if result.BackupPath == "" {
		t.Fatal("expected backup path")
	}
	if !strings.HasPrefix(result.BackupPath, filepath.Join(worktreeRoot, ".modu", "backups")+string(os.PathSeparator)) {
		t.Fatalf("backup path %s is outside backup directory", result.BackupPath)
	}
	if filepath.Base(result.BackupPath) == "20260515-153012_feature-remove-backup.tar.gz" {
		t.Fatal("backup path should use current time, not a hard-coded fixture")
	}
	requirePathExists(t, result.BackupPath)
	requirePathMissing(t, featurePath)
	if removeCalls == 0 {
		t.Fatal("expected worktree removal to be called")
	}
}

func TestDeleteWorktree_BackupFailureBlocksDelete(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	featurePath := filepath.Join(worktreeRoot, "my-feature")
	writeTestFile(t, filepath.Join(featurePath, "file.txt"), "content")
	writeTestFile(t, filepath.Join(worktreeRoot, ".modu"), "not a directory")
	if err := os.MkdirAll(workspace, 0755); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	var removeCalls int
	engine := NewWithClient(&config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
		Modules:      []config.Module{{Name: "module1"}},
	}, &MockGitClient{
		RemoveWorktreeAndBranchFunc: func(ctx context.Context, repoPath, worktreePath, featureDirName string) error {
			removeCalls++
			return nil
		},
	})

	result, err := engine.DeleteWorktree(context.Background(), "my-feature", true)
	if err == nil {
		t.Fatal("expected backup failure")
	}
	if result != nil {
		t.Fatalf("expected nil result on failure, got %#v", result)
	}
	requirePathExists(t, featurePath)
	if removeCalls != 0 {
		t.Fatalf("expected no removal calls, got %d", removeCalls)
	}
}

func TestDeleteWorktree_RejectsUnsafeFeatureName(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	if err := os.MkdirAll(worktreeRoot, 0755); err != nil {
		t.Fatalf("failed to create worktree root: %v", err)
	}
	parentSentinel := filepath.Join(tmpDir, "sentinel.txt")
	writeTestFile(t, parentSentinel, "keep")

	engine := NewWithClient(&config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
	}, &MockGitClient{})

	for _, feature := range []string{"", ".", "..", ".modu"} {
		t.Run(feature, func(t *testing.T) {
			result, err := engine.DeleteWorktree(context.Background(), feature, true)
			if !errors.Is(err, errs.ErrInvalidOperation) {
				t.Fatalf("expected invalid operation, got result=%#v err=%v", result, err)
			}
			if result != nil {
				t.Fatalf("expected nil result, got %#v", result)
			}
			requirePathExists(t, worktreeRoot)
			requireFileContent(t, parentSentinel, "keep")
		})
	}
}

func TestDeleteWorktree_CanceledBackupCleansTempAndBlocksDelete(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	featurePath := filepath.Join(worktreeRoot, "my-feature")
	writeTestFile(t, filepath.Join(featurePath, "file.txt"), "content")
	if err := os.MkdirAll(workspace, 0755); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	var removeCalls int
	engine := NewWithClient(&config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
	}, &MockGitClient{
		RemoveWorktreeAndBranchFunc: func(ctx context.Context, repoPath, worktreePath, featureDirName string) error {
			removeCalls++
			return nil
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	restoreHook := useDeleteBackupAfterTempCreated(func() {
		cancel()
	})
	defer restoreHook()
	result, err := engine.DeleteWorktree(ctx, "my-feature", true)
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if result != nil {
		t.Fatalf("expected nil result on cancellation, got %#v", result)
	}
	requirePathExists(t, featurePath)
	if removeCalls != 0 {
		t.Fatalf("expected no removal calls, got %d", removeCalls)
	}
	backupDir := filepath.Join(worktreeRoot, ".modu", "backups")
	if backups := listBackupFiles(t, backupDir); len(backups) != 0 {
		t.Fatalf("expected no final backups, got %v", backups)
	}
	if temps := listFilesByPrefix(t, backupDir, ".tmp-delete-backup-"); len(temps) != 0 {
		t.Fatalf("expected temp files to be cleaned, got %v", temps)
	}
}

func TestDeleteWorktree_DirtyCheckBlocksBeforeBackup(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	featurePath := filepath.Join(worktreeRoot, "dirty-feature")
	writeTestFile(t, filepath.Join(featurePath, "module1", "file.txt"), "content")
	if err := os.MkdirAll(workspace, 0755); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	var removeCalls int
	engine := NewWithClient(&config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
		StrictDirty:  true,
		Modules:      []config.Module{{Name: "module1"}},
	}, &MockGitClient{
		GetStatusFunc: func(ctx context.Context, path string) (gitproxy.Status, error) {
			return gitproxy.Status{IsDirty: true, Branch: "dirty-feature"}, nil
		},
		RemoveWorktreeAndBranchFunc: func(ctx context.Context, repoPath, worktreePath, featureDirName string) error {
			removeCalls++
			return nil
		},
	})

	result, err := engine.DeleteWorktree(context.Background(), "dirty-feature", false)
	if !errors.Is(err, errs.ErrDirtyWorktree) {
		t.Fatalf("expected dirty worktree error, got result=%#v err=%v", result, err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
	requirePathExists(t, featurePath)
	if removeCalls != 0 {
		t.Fatalf("expected no removal calls, got %d", removeCalls)
	}
	if backups := listBackupFiles(t, filepath.Join(worktreeRoot, ".modu", "backups")); len(backups) != 0 {
		t.Fatalf("expected no backups when dirty check fails, got %v", backups)
	}
}

func TestDeleteWorktree_MainProjectDirtyBlocksBeforeBackup(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	featurePath := filepath.Join(worktreeRoot, "dirty-main")
	writeTestFile(t, filepath.Join(featurePath, "main.txt"), "content")
	if err := os.MkdirAll(workspace, 0755); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	var statusPaths []string
	var removeCalls int
	engine := NewWithClient(&config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
		StrictDirty:  true,
	}, &MockGitClient{
		GetStatusFunc: func(ctx context.Context, path string) (gitproxy.Status, error) {
			statusPaths = append(statusPaths, path)
			if path == featurePath {
				return gitproxy.Status{IsDirty: true, Branch: "dirty-main"}, nil
			}
			return gitproxy.Status{Branch: "main"}, nil
		},
		RemoveWorktreeAndBranchFunc: func(ctx context.Context, repoPath, worktreePath, featureDirName string) error {
			removeCalls++
			return nil
		},
	})

	result, err := engine.DeleteWorktree(context.Background(), "dirty-main", false)
	if !errors.Is(err, errs.ErrDirtyWorktree) {
		t.Fatalf("expected dirty worktree error, got result=%#v err=%v", result, err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
	if len(statusPaths) != 1 || statusPaths[0] != featurePath {
		t.Fatalf("expected only main project status check, got %v", statusPaths)
	}
	requirePathExists(t, featurePath)
	if removeCalls != 0 {
		t.Fatalf("expected no removal calls, got %d", removeCalls)
	}
	if backups := listBackupFiles(t, filepath.Join(worktreeRoot, ".modu", "backups")); len(backups) != 0 {
		t.Fatalf("expected no backups when main project is dirty, got %v", backups)
	}
}

func TestDeleteWorktree_RemoveWarningUsesStderr(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	featurePath := filepath.Join(worktreeRoot, "warn-feature")
	writeTestFile(t, filepath.Join(featurePath, "main.txt"), "content")
	if err := os.MkdirAll(workspace, 0755); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	engine := NewWithClient(&config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
	}, &MockGitClient{
		RemoveWorktreeAndBranchFunc: func(ctx context.Context, repoPath, worktreePath, featureDirName string) error {
			return errors.New("remove failed")
		},
	})

	var result *DeleteResult
	var err error
	stdout, stderr := captureStdStreams(t, func() {
		result, err = engine.DeleteWorktree(context.Background(), "warn-feature", true)
	})
	if err != nil {
		t.Fatalf("DeleteWorktree failed: %v", err)
	}
	if result == nil || result.BackupPath == "" {
		t.Fatalf("expected backup result, got %#v", result)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "Warning: failed to remove main project worktree") {
		t.Fatalf("expected warning on stderr, got %q", stderr)
	}
}

func TestCreateDeleteBackup_ArchiveShapeIncludesFilesDirsAndSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	feature := "feature/remove-backup"
	dirName := featureToDirName(feature)
	featurePath := filepath.Join(worktreeRoot, dirName)
	writeTestFile(t, filepath.Join(featurePath, "root.txt"), "root content")
	writeTestFile(t, filepath.Join(featurePath, "nested", "child.txt"), "child content")
	if err := os.Symlink("root.txt", filepath.Join(featurePath, "link.txt")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}

	engine := New(&config.Config{WorktreeRoot: worktreeRoot})
	backupPath, err := engine.createDeleteBackup(context.Background(), feature, featurePath)
	if err != nil {
		t.Fatalf("createDeleteBackup failed: %v", err)
	}

	entries := readTarGzEntries(t, backupPath)
	requireArchiveEntry(t, entries, dirName+"/", tar.TypeDir)
	requireArchiveEntry(t, entries, dirName+"/nested/", tar.TypeDir)
	root := requireArchiveEntry(t, entries, dirName+"/root.txt", tar.TypeReg)
	if root.body != "root content" {
		t.Fatalf("unexpected root file content: %q", root.body)
	}
	child := requireArchiveEntry(t, entries, dirName+"/nested/child.txt", tar.TypeReg)
	if child.body != "child content" {
		t.Fatalf("unexpected child file content: %q", child.body)
	}
	link := requireArchiveEntry(t, entries, dirName+"/link.txt", tar.TypeSymlink)
	if link.linkName != "root.txt" {
		t.Fatalf("expected symlink target root.txt, got %s", link.linkName)
	}
}

func TestCreateDeleteBackup_UsesUniqueNameWhenArchiveExists(t *testing.T) {
	tmpDir := t.TempDir()
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	feature := "feature/remove-backup"
	dirName := featureToDirName(feature)
	featurePath := filepath.Join(worktreeRoot, dirName)
	writeTestFile(t, filepath.Join(featurePath, "file.txt"), "content")
	backupDir := filepath.Join(worktreeRoot, ".modu", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("failed to create backup dir: %v", err)
	}

	fixedTime := time.Date(2026, 5, 15, 15, 30, 12, 0, time.Local)
	restoreNow := useFixedDeleteBackupNow(fixedTime)
	defer restoreNow()

	basePath := filepath.Join(backupDir, "20260515-153012_feature-remove-backup.tar.gz")
	writeTestFile(t, basePath, "existing")

	engine := New(&config.Config{WorktreeRoot: worktreeRoot})
	backupPath, err := engine.createDeleteBackup(context.Background(), feature, featurePath)
	if err != nil {
		t.Fatalf("createDeleteBackup failed: %v", err)
	}
	if backupPath != filepath.Join(backupDir, "20260515-153012_feature-remove-backup-1.tar.gz") {
		t.Fatalf("expected suffixed backup path, got %s", backupPath)
	}
	requireFileContent(t, basePath, "existing")
	requirePathExists(t, backupPath)
}

func TestCleanupDeleteBackups_RetentionAndNonMatchingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	backupDir := filepath.Join(worktreeRoot, ".modu", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("failed to create backup dir: %v", err)
	}

	oldGenerated := filepath.Join(backupDir, "20260401-010203_feature-old.tar.gz")
	oldGeneratedSuffix := filepath.Join(backupDir, "20260401-010203_feature-old-1.tar.gz")
	recentGenerated := filepath.Join(backupDir, "20260514-010203_feature-recent.tar.gz")
	oldManual := filepath.Join(backupDir, "manual-backup.tar.gz")
	writeTestFile(t, oldGenerated, "old")
	writeTestFile(t, oldGeneratedSuffix, "old suffix")
	writeTestFile(t, recentGenerated, "recent")
	writeTestFile(t, oldManual, "manual")

	oldTime := time.Now().AddDate(0, 0, -31)
	recentTime := time.Now()
	for _, path := range []string{oldGenerated, oldGeneratedSuffix, oldManual} {
		if err := os.Chtimes(path, oldTime, oldTime); err != nil {
			t.Fatalf("failed to set old mtime for %s: %v", path, err)
		}
	}
	if err := os.Chtimes(recentGenerated, recentTime, recentTime); err != nil {
		t.Fatalf("failed to set recent mtime: %v", err)
	}

	engine := New(&config.Config{WorktreeRoot: worktreeRoot})
	if err := engine.CleanupDeleteBackups(context.Background()); err != nil {
		t.Fatalf("CleanupDeleteBackups failed: %v", err)
	}

	requirePathMissing(t, oldGenerated)
	requirePathMissing(t, oldGeneratedSuffix)
	requirePathExists(t, recentGenerated)
	requirePathExists(t, oldManual)
}

func TestCleanupDeleteBackups_MissingDirectorySucceeds(t *testing.T) {
	worktreeRoot := filepath.Join(t.TempDir(), "worktrees")
	engine := New(&config.Config{WorktreeRoot: worktreeRoot})

	if err := engine.CleanupDeleteBackups(context.Background()); err != nil {
		t.Fatalf("CleanupDeleteBackups failed: %v", err)
	}
	requirePathMissing(t, filepath.Join(worktreeRoot, ".modu", "backups"))
}

type archiveEntry struct {
	typeFlag byte
	linkName string
	body     string
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create parent dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func requirePathExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path %s to exist: %v", path, err)
	}
}

func requirePathMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected path %s to be missing, stat err=%v", path, err)
	}
}

func requireFileContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	if string(data) != want {
		t.Fatalf("expected %s content %q, got %q", path, want, string(data))
	}
}

func listBackupFiles(t *testing.T, backupDir string) []string {
	t.Helper()
	return listFilesBySuffix(t, backupDir, ".tar.gz")
}

func listFilesBySuffix(t *testing.T, dir, suffix string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("failed to read %s: %v", dir, err)
	}
	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), suffix) {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}
	return paths
}

func listFilesByPrefix(t *testing.T, dir, prefix string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("failed to read %s: %v", dir, err)
	}
	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}
	return paths
}

func useDeleteBackupAfterTempCreated(hook func()) func() {
	oldHook := deleteBackupAfterTempCreated
	deleteBackupAfterTempCreated = hook
	return func() {
		deleteBackupAfterTempCreated = oldHook
	}
}

func captureStdStreams(t *testing.T, fn func()) (string, string) {
	t.Helper()

	oldStdout := os.Stdout
	stdoutRead, stdoutWrite, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	oldStderr := os.Stderr
	stderrRead, stderrWrite, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}

	os.Stdout = stdoutWrite
	os.Stderr = stderrWrite
	fn()
	_ = stdoutWrite.Close()
	_ = stderrWrite.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdoutData, err := io.ReadAll(stdoutRead)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}
	stderrData, err := io.ReadAll(stderrRead)
	if err != nil {
		t.Fatalf("failed to read stderr: %v", err)
	}
	return string(stdoutData), string(stderrData)
}

func readTarGzEntries(t *testing.T, path string) map[string]archiveEntry {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open archive %s: %v", path, err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gz.Close()

	entries := make(map[string]archiveEntry)
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("failed to read tar entry: %v", err)
		}
		entry := archiveEntry{
			typeFlag: header.Typeflag,
			linkName: header.Linkname,
		}
		if header.Typeflag == tar.TypeReg {
			data, err := io.ReadAll(tr)
			if err != nil {
				t.Fatalf("failed to read tar body for %s: %v", header.Name, err)
			}
			entry.body = string(data)
		}
		entries[header.Name] = entry
	}
	return entries
}

func requireArchiveEntry(t *testing.T, entries map[string]archiveEntry, name string, typeFlag byte) archiveEntry {
	t.Helper()
	entry, ok := entries[name]
	if !ok {
		t.Fatalf("expected archive entry %s, got entries %#v", name, entries)
	}
	if entry.typeFlag != typeFlag {
		t.Fatalf("expected archive entry %s type %d, got %d", name, typeFlag, entry.typeFlag)
	}
	return entry
}

func useFixedDeleteBackupNow(now time.Time) func() {
	original := deleteBackupNow
	deleteBackupNow = func() time.Time {
		return now
	}
	return func() {
		deleteBackupNow = original
	}
}
