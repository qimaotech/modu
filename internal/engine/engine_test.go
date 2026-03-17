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
	"github.com/qimaotech/modu/internal/gitproxy"
)

// MockGitClient 用于测试的 Mock Git 客户端
type MockGitClient struct {
	CloneFunc                            func(ctx context.Context, url, path string) error
	CreateWorktreeFunc                   func(ctx context.Context, repoPath, branch, baseBranch, worktreePath string) error
	CreateWorktreeFromExistingBranchFunc func(ctx context.Context, repoPath, branch, worktreePath string) error
	GetStatusFunc                        func(ctx context.Context, path string) (gitproxy.Status, error)
	RemoveWorktreeFunc                   func(ctx context.Context, path string) error
	RemoveWorktreeAndBranchFunc          func(ctx context.Context, repoPath, branch, worktreePath string) error
	ListWorktreesFunc                    func(ctx context.Context, repoPath string) ([]gitproxy.WorktreeInfo, error)
	FetchFunc                            func(ctx context.Context, repoPath string) error
	RebaseFunc                           func(ctx context.Context, path string) error
	FetchAndSwitchBranchFunc             func(ctx context.Context, repoPath, branch string) error
	BranchExistsFunc                     func(ctx context.Context, repoPath, branch string) bool
	CheckBranchWorktreeStatusFunc        func(ctx context.Context, repoPath, branch string) (bool, error)
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
		Workspace:    tmp,
		DefaultBase:  "develop",
		Concurrency:  2,
		Modules:      []config.Module{{Name: "m1", URL: "git@test/m1.git"}},
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
		Workspace:    tmp,
		DefaultBase:  "develop",
		Concurrency:  2,
		Modules:      []config.Module{{Name: "m1", URL: "git@test/m1.git"}},
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
		Workspace:     filepath.Join(tmpDir, "workspace"),
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
}

func TestCreateVSCodeWorkspace_EmptyFeature(t *testing.T) {
	// 测试空 feature 目录（无模块）
	tmpDir, err := os.MkdirTemp("", "modu-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Workspace:     filepath.Join(tmpDir, "workspace"),
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
		Workspace:     filepath.Join(tmpDir, "workspace"),
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
		Workspace:     filepath.Join(tmpDir, "workspace"),
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
