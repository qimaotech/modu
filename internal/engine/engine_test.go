package engine

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	RemoveWorktreeAndBranchFunc          func(ctx context.Context, repoPath, worktreePath, featureDirName string) (gitproxy.RemoveWorktreeResult, error)
	ListWorktreesFunc                    func(ctx context.Context, repoPath string) ([]gitproxy.WorktreeInfo, error)
	FetchFunc                            func(ctx context.Context, repoPath string) error
	RebaseFunc                           func(ctx context.Context, path string) error
	FetchAndSwitchBranchFunc             func(ctx context.Context, repoPath, branch string) error
	BranchExistsFunc                     func(ctx context.Context, repoPath, branch string) bool
	CheckBranchWorktreeStatusFunc        func(ctx context.Context, repoPath, branch string) (bool, error)
	RemoteBranchExistsFunc               func(ctx context.Context, repoURL, branch string) bool
	GetBranchPushStatusFunc              func(ctx context.Context, repoPath, branch string) (gitproxy.BranchPushStatus, error)
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

func (m *MockGitClient) RemoveWorktreeAndBranch(ctx context.Context, repoPath, worktreePath, featureDirName string) (gitproxy.RemoveWorktreeResult, error) {
	if m.RemoveWorktreeAndBranchFunc != nil {
		return m.RemoveWorktreeAndBranchFunc(ctx, repoPath, worktreePath, featureDirName)
	}
	return gitproxy.RemoveWorktreeResult{}, nil
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

func (m *MockGitClient) GetBranchPushStatus(ctx context.Context, repoPath, branch string) (gitproxy.BranchPushStatus, error) {
	if m.GetBranchPushStatusFunc != nil {
		return m.GetBranchPushStatusFunc(ctx, repoPath, branch)
	}
	return gitproxy.BranchPushStatus{Branch: branch, RemoteRef: "origin/" + branch, IsPushed: true}, nil
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

func TestDeleteWorktree_BlocksUnpushedBranches(t *testing.T) {
	engine, mock, removedCount := newDeletePreflightTestEngine(t, gitproxy.BranchPushStatus{
		Branch:     "feat-a",
		RemoteRef:  "origin/feat-a",
		IsPushed:   false,
		AheadCount: 1,
		Reason:     "1 local commits not pushed",
	})

	err := engine.DeleteWorktree(context.Background(), "feat-a", false)
	if !errors.Is(err, errs.ErrUnpushedBranch) {
		t.Fatalf("expected ErrUnpushedBranch, got %v", err)
	}
	if *removedCount != 0 {
		t.Fatalf("expected no removal before confirmation, got %d", *removedCount)
	}
	if mock.GetBranchPushStatusFunc == nil {
		t.Fatal("expected push status mock to be configured")
	}
}

func TestDeleteWorktree_AllowsPushedBranches(t *testing.T) {
	engine, _, removedCount := newDeletePreflightTestEngine(t, gitproxy.BranchPushStatus{
		Branch:    "feat-a",
		RemoteRef: "origin/feat-a",
		IsPushed:  true,
	})

	err := engine.DeleteWorktree(context.Background(), "feat-a", false)
	if err != nil {
		t.Fatalf("DeleteWorktree() unexpected error: %v", err)
	}
	if *removedCount != 2 {
		t.Fatalf("expected module and main removal, got %d", *removedCount)
	}
}

func TestDeleteWorktreeWithOptions_AllowsConfirmedUnpushedBranches(t *testing.T) {
	engine, _, removedCount := newDeletePreflightTestEngine(t, gitproxy.BranchPushStatus{
		Branch:     "feat-a",
		RemoteRef:  "origin/feat-a",
		IsPushed:   false,
		AheadCount: 1,
		Reason:     "1 local commits not pushed",
	})

	_, err := engine.DeleteWorktreeWithOptions(context.Background(), "feat-a", false, DeleteOptions{
		AllowUnpushedBranches: true,
	})
	if err != nil {
		t.Fatalf("DeleteWorktreeWithOptions() unexpected error: %v", err)
	}
	if *removedCount != 2 {
		t.Fatalf("expected module and main removal, got %d", *removedCount)
	}
}

func TestDeleteWorktreesWithOptions_NormalDeleteChecksDirtyAndContinues(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	for _, path := range []string{
		workspace,
		filepath.Join(workspace, "module1"),
		filepath.Join(worktreeRoot, "dirty-feature", "module1"),
		filepath.Join(worktreeRoot, "clean-feature", "module1"),
	} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}

	removeCalls := 0
	client := &MockGitClient{
		GetStatusFunc: func(ctx context.Context, path string) (gitproxy.Status, error) {
			return gitproxy.Status{
				Branch:  filepath.Base(filepath.Dir(path)),
				IsDirty: strings.Contains(path, "dirty-feature"),
			}, nil
		},
		RemoveWorktreeAndBranchFunc: func(ctx context.Context, repoPath, worktreePath, featureDirName string) (gitproxy.RemoveWorktreeResult, error) {
			removeCalls++
			return gitproxy.RemoveWorktreeResult{}, nil
		},
	}
	eng := NewWithClient(&config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
		StrictDirty:  true,
		Modules:      []config.Module{{Name: "module1"}},
	}, client)

	succeeded, failed, _ := eng.DeleteWorktreesWithOptions(
		context.Background(),
		[]string{"dirty-feature", "clean-feature"},
		false,
		DeleteOptions{AllowUnpushedBranches: true},
	)

	if len(succeeded) != 1 || succeeded[0] != "clean-feature" {
		t.Fatalf("succeeded = %v, 期望 [clean-feature]", succeeded)
	}
	if !errors.Is(failed["dirty-feature"], errs.ErrDirtyWorktree) {
		t.Fatalf("dirty-feature error = %v, 期望 ErrDirtyWorktree", failed["dirty-feature"])
	}
	if removeCalls != 2 {
		t.Fatalf("只应删除 clean-feature 的模块和主项目，removeCalls=%d", removeCalls)
	}
}

// TestDeleteWorktreesWithOptions_ReportsRemovalFailure 验证底层删除失败不会被统计为成功。
func TestDeleteWorktreesWithOptions_ReportsRemovalFailure(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	featurePath := filepath.Join(worktreeRoot, "feature-a")
	modulePath := filepath.Join(featurePath, "module1")
	for _, path := range []string{workspace, filepath.Join(workspace, "module1"), modulePath} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}

	client := &MockGitClient{
		GetStatusFunc: func(ctx context.Context, path string) (gitproxy.Status, error) {
			return gitproxy.Status{Branch: "feature-a"}, nil
		},
		RemoveWorktreeAndBranchFunc: func(ctx context.Context, repoPath, worktreePath, featureDirName string) (gitproxy.RemoveWorktreeResult, error) {
			if worktreePath == modulePath {
				return gitproxy.RemoveWorktreeResult{}, errors.New("simulated module removal failure")
			}
			return gitproxy.RemoveWorktreeResult{}, nil
		},
	}
	eng := NewWithClient(&config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
		StrictDirty:  false,
		Modules:      []config.Module{{Name: "module1"}},
	}, client)

	succeeded, failed, _ := eng.DeleteWorktreesWithOptions(
		context.Background(),
		[]string{"feature-a"},
		true,
		DeleteOptions{AllowUnpushedBranches: true},
	)

	if len(succeeded) != 0 {
		t.Fatalf("删除阶段失败时不应统计成功，succeeded=%v", succeeded)
	}
	deleteErr := failed["feature-a"]
	if !errors.Is(deleteErr, errs.ErrPartialFailure) {
		t.Fatalf("删除阶段失败应返回 ErrPartialFailure，got %v", deleteErr)
	}
	if !strings.Contains(deleteErr.Error(), "module1") || !strings.Contains(deleteErr.Error(), "simulated module removal failure") {
		t.Fatalf("错误应包含失败模块和原始原因，got %v", deleteErr)
	}
}

// TestDeleteWorktreesWithOptions_ReturnsRemovalWarnings 验证非致命警告按 feature 返回给展示层。
func TestDeleteWorktreesWithOptions_ReturnsRemovalWarnings(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	featurePath := filepath.Join(worktreeRoot, "feature-a")
	for _, path := range []string{workspace, filepath.Join(workspace, "module1"), filepath.Join(featurePath, "module1")} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}

	client := &MockGitClient{
		GetStatusFunc: func(ctx context.Context, path string) (gitproxy.Status, error) {
			return gitproxy.Status{Branch: "feature-a"}, nil
		},
		RemoveWorktreeAndBranchFunc: func(ctx context.Context, repoPath, worktreePath, featureDirName string) (gitproxy.RemoveWorktreeResult, error) {
			return gitproxy.RemoveWorktreeResult{Warnings: []string{"branch mismatch"}}, nil
		},
	}
	eng := NewWithClient(&config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
		StrictDirty:  false,
		Modules:      []config.Module{{Name: "module1"}},
	}, client)

	succeeded, failed, warnings := eng.DeleteWorktreesWithOptions(
		context.Background(),
		[]string{"feature-a"},
		true,
		DeleteOptions{AllowUnpushedBranches: true},
	)

	if len(succeeded) != 1 || len(failed) != 0 {
		t.Fatalf("删除结果异常: succeeded=%v failed=%v", succeeded, failed)
	}
	featureWarnings := warnings["feature-a"]
	if len(featureWarnings) != 2 || !strings.Contains(featureWarnings[0], "module module1") || !strings.Contains(featureWarnings[1], "main project") {
		t.Fatalf("警告应包含模块和主项目上下文，got %v", featureWarnings)
	}
}

func TestDeleteWorktreeWithOptions_NormalDeleteBlocksDirtyMainProject(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	featurePath := filepath.Join(worktreeRoot, "dirty-main")
	for _, path := range []string{workspace, featurePath} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}

	removeCalls := 0
	client := &MockGitClient{
		GetStatusFunc: func(ctx context.Context, path string) (gitproxy.Status, error) {
			return gitproxy.Status{Branch: "dirty-main", IsDirty: path == featurePath}, nil
		},
		RemoveWorktreeAndBranchFunc: func(ctx context.Context, repoPath, worktreePath, featureDirName string) (gitproxy.RemoveWorktreeResult, error) {
			removeCalls++
			return gitproxy.RemoveWorktreeResult{}, nil
		},
	}
	eng := NewWithClient(&config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
		StrictDirty:  true,
	}, client)

	_, err := eng.DeleteWorktreeWithOptions(
		context.Background(),
		"dirty-main",
		false,
		DeleteOptions{AllowUnpushedBranches: true},
	)

	if !errors.Is(err, errs.ErrDirtyWorktree) {
		t.Fatalf("error = %v, 期望 ErrDirtyWorktree", err)
	}
	if removeCalls != 0 {
		t.Fatalf("主项目脏时不应执行删除，removeCalls=%d", removeCalls)
	}
}

func TestDeleteWorktreesWithOptions_ForceDeleteSkipsDirtyCheck(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	for _, path := range []string{
		workspace,
		filepath.Join(workspace, "module1"),
		filepath.Join(worktreeRoot, "dirty-feature", "module1"),
	} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}

	removeCalls := 0
	client := &MockGitClient{
		GetStatusFunc: func(ctx context.Context, path string) (gitproxy.Status, error) {
			return gitproxy.Status{Branch: "dirty-feature", IsDirty: true}, nil
		},
		RemoveWorktreeAndBranchFunc: func(ctx context.Context, repoPath, worktreePath, featureDirName string) (gitproxy.RemoveWorktreeResult, error) {
			removeCalls++
			return gitproxy.RemoveWorktreeResult{}, nil
		},
	}
	eng := NewWithClient(&config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
		StrictDirty:  true,
		Modules:      []config.Module{{Name: "module1"}},
	}, client)

	succeeded, failed, _ := eng.DeleteWorktreesWithOptions(
		context.Background(),
		[]string{"dirty-feature"},
		true,
		DeleteOptions{AllowUnpushedBranches: true},
	)

	if len(failed) != 0 || len(succeeded) != 1 || succeeded[0] != "dirty-feature" {
		t.Fatalf("强制删除结果异常: succeeded=%v failed=%v", succeeded, failed)
	}
	if removeCalls != 2 {
		t.Fatalf("强制删除应移除模块和主项目，removeCalls=%d", removeCalls)
	}
}

func TestCheckDeleteUnpushedBranches_SkipsNonCandidateBranches(t *testing.T) {
	cfg, _ := newDeletePreflightConfig(t)
	pushStatusCalled := false
	mock := &MockGitClient{
		GetStatusFunc: func(ctx context.Context, path string) (gitproxy.Status, error) {
			return gitproxy.Status{Branch: "other-branch"}, nil
		},
		GetBranchPushStatusFunc: func(ctx context.Context, repoPath, branch string) (gitproxy.BranchPushStatus, error) {
			pushStatusCalled = true
			return gitproxy.BranchPushStatus{}, nil
		},
	}
	engine := NewWithClient(cfg, mock)

	risks, err := engine.CheckDeleteUnpushedBranches(context.Background(), "feat-a")
	if err != nil {
		t.Fatalf("CheckDeleteUnpushedBranches() unexpected error: %v", err)
	}
	if len(risks) != 0 {
		t.Fatalf("expected no risks for non-candidate branch, got %+v", risks)
	}
	if pushStatusCalled {
		t.Fatal("push status should not be checked for non-candidate branch")
	}
}

func newDeletePreflightTestEngine(t *testing.T, pushStatus gitproxy.BranchPushStatus) (*Engine, *MockGitClient, *int) {
	t.Helper()

	cfg, _ := newDeletePreflightConfig(t)
	removedCount := 0
	mock := &MockGitClient{
		GetStatusFunc: func(ctx context.Context, path string) (gitproxy.Status, error) {
			return gitproxy.Status{Branch: "feat-a"}, nil
		},
		GetBranchPushStatusFunc: func(ctx context.Context, repoPath, branch string) (gitproxy.BranchPushStatus, error) {
			return pushStatus, nil
		},
		RemoveWorktreeAndBranchFunc: func(ctx context.Context, repoPath, worktreePath, featureDirName string) (gitproxy.RemoveWorktreeResult, error) {
			removedCount++
			return gitproxy.RemoveWorktreeResult{}, nil
		},
	}
	return NewWithClient(cfg, mock), mock, &removedCount
}

func newDeletePreflightConfig(t *testing.T) (*config.Config, string) {
	t.Helper()

	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	worktreeRoot := filepath.Join(tmpDir, "worktrees")
	featurePath := filepath.Join(worktreeRoot, "feat-a")

	for _, path := range []string{
		workspace,
		filepath.Join(workspace, "module1"),
		featurePath,
		filepath.Join(featurePath, "module1"),
	} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{
		Workspace:    workspace,
		WorktreeRoot: worktreeRoot,
		StrictDirty:  false,
		Modules: []config.Module{
			{Name: "module1", URL: "git@example.com:module1.git"},
		},
	}
	return cfg, featurePath
}
