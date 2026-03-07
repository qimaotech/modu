package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			content: `workspace: /opt/workspace
worktree-root: /opt/worktrees
default-base: develop
concurrency: 8
auto-fetch: true
strict-dirty-check: true
modules:
  - name: auth-svc
    url: git@github.com:example/auth-svc.git
`,
			wantErr: false,
		},
		{
			name: "missing workspace",
			content: `worktree-root: /opt/worktrees
default-base: develop
modules:
  - name: auth-svc
    url: git@github.com:example/auth-svc.git
`,
			wantErr:     true,
			errContains: "workspace is required",
		},
		{
			name: "missing worktree-root",
			content: `workspace: /opt/workspace
default-base: develop
modules:
  - name: auth-svc
    url: git@github.com:example/auth-svc.git
`,
			wantErr:     true,
			errContains: "worktree-root is required",
		},
		{
			name: "missing default-base",
			content: `workspace: /opt/workspace
worktree-root: /opt/worktrees
modules:
  - name: auth-svc
    url: git@github.com:example/auth-svc.git
`,
			wantErr:     true,
			errContains: "default-base is required",
		},
		{
			name: "missing modules",
			content: `workspace: /opt/workspace
worktree-root: /opt/worktrees
default-base: develop
`,
			wantErr:     true,
			errContains: "at least one module is required",
		},
		{
			name:        "missing both required fields",
			content:     `modules: []`,
			wantErr:     true,
			errContains: "workspace is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建临时文件
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".modu.yaml")

			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			cfg, err := LoadConfig(configPath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if cfg.Workspace != "/opt/workspace" {
				t.Errorf("expected workspace /opt/workspace, got %s", cfg.Workspace)
			}
			if cfg.WorktreeRoot != "/opt/worktrees" {
				t.Errorf("expected worktree-root /opt/worktrees, got %s", cfg.WorktreeRoot)
			}
			if cfg.DefaultBase != "develop" {
				t.Errorf("expected default-base develop, got %s", cfg.DefaultBase)
			}
			if cfg.Concurrency != 8 {
				t.Errorf("expected concurrency 8, got %d", cfg.Concurrency)
			}
			if len(cfg.Modules) != 1 {
				t.Errorf("expected 1 module, got %d", len(cfg.Modules))
			}
		})
	}
}

func TestLoadConfig_DefaultConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".modu.yaml")
	content := `workspace: /opt/workspace
worktree-root: /opt/worktrees
default-base: develop
modules:
  - name: auth-svc
    url: git@github.com:example/auth-svc.git
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Concurrency != 5 {
		t.Errorf("expected default concurrency 5, got %d", cfg.Concurrency)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Workspace != "./workspace" {
		t.Errorf("expected workspace ./workspace, got %s", cfg.Workspace)
	}
	if cfg.WorktreeRoot != "./worktrees" {
		t.Errorf("expected worktree-root ./worktrees, got %s", cfg.WorktreeRoot)
	}
	if cfg.DefaultBase != "develop" {
		t.Errorf("expected default-base develop, got %s", cfg.DefaultBase)
	}
	if cfg.Concurrency != 5 {
		t.Errorf("expected concurrency 5, got %d", cfg.Concurrency)
	}
	if !cfg.AutoFetch {
		t.Error("expected AutoFetch to be true")
	}
	if !cfg.StrictDirty {
		t.Error("expected StrictDirty to be true")
	}
	if len(cfg.Modules) != 0 {
		t.Errorf("expected empty modules, got %d", len(cfg.Modules))
	}
}

// TestSaveConfig 测试保存配置
func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")

	cfg := &Config{
		Workspace:    "/test/workspace",
		WorktreeRoot: "/test/worktrees",
		DefaultBase:  "main",
		Concurrency: 3,
		AutoFetch:    false,
		StrictDirty:  false,
		Modules: []Module{
			{Name: "test-module", URL: "git@example.com:test/module.git"},
		},
	}

	err := SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证文件已创建
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected config file to exist")
	}

	// 验证可以重新加载
	loadedCfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if loadedCfg.Workspace != cfg.Workspace {
		t.Errorf("expected workspace %s, got %s", cfg.Workspace, loadedCfg.Workspace)
	}
	if loadedCfg.WorktreeRoot != cfg.WorktreeRoot {
		t.Errorf("expected worktree-root %s, got %s", cfg.WorktreeRoot, loadedCfg.WorktreeRoot)
	}
	if loadedCfg.DefaultBase != cfg.DefaultBase {
		t.Errorf("expected default-base %s, got %s", cfg.DefaultBase, loadedCfg.DefaultBase)
	}
	if loadedCfg.Concurrency != cfg.Concurrency {
		t.Errorf("expected concurrency %d, got %d", cfg.Concurrency, loadedCfg.Concurrency)
	}
	if len(loadedCfg.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(loadedCfg.Modules))
	}
}

// TestSaveConfig_InvalidPath 测试保存到无效路径
func TestSaveConfig_InvalidPath(t *testing.T) {
	cfg := &Config{
		Workspace:    "/test",
		WorktreeRoot: "/test",
		DefaultBase:  "main",
		Modules:     []Module{},
	}

	err := SaveConfig(cfg, "/invalid/path/that/does/not/exist/test.yaml")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

// TestIsConfigNotFoundError 测试配置不存在错误检查
func TestIsConfigNotFoundError(t *testing.T) {
	// 测试包装的错误
	wrappedErr := fmt.Errorf("wrapped: %w", ErrConfigNotFound)
	if !IsConfigNotFoundError(wrappedErr) {
		t.Error("expected true for wrapped config not found error")
	}

	// 测试直接错误
	if !IsConfigNotFoundError(ErrConfigNotFound) {
		t.Error("expected true for config not found error")
	}

	// 测试其他错误
	otherErr := errors.New("some other error")
	if IsConfigNotFoundError(otherErr) {
		t.Error("expected false for other error")
	}

	// 测试 nil
	if IsConfigNotFoundError(nil) {
		t.Error("expected false for nil")
	}
}

// TestLoadConfig_FileNotFound 测试文件不存在
func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/path/that/does/not/exist.yaml")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
	if !IsConfigNotFoundError(err) {
		t.Error("expected config not found error")
	}
}

// TestLoadConfig_InvalidYAML 测试无效 YAML
func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	if err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

// TestLoadConfig_AbsolutePath 测试绝对路径解析
func TestLoadConfig_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")

	content := `workspace: /opt/workspace
worktree-root: /opt/worktrees
default-base: develop
modules:
  - name: test
    url: git@example.com:test.git
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证配置加载成功
	if cfg.Workspace != "/opt/workspace" {
		t.Errorf("expected workspace /opt/workspace, got %s", cfg.Workspace)
	}
}

// TestScanWorkspace 测试扫描 workspace 目录
func TestScanWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)

	// 创建测试 git 仓库
	testCases := []struct {
		name     string
		url      string
		wantName string
	}{
		{name: "module-a", url: "git@github.com:example/module-a.git", wantName: "module-a"},
		{name: "module-b", url: "git@github.com:example/module-b.git", wantName: "module-b"},
	}

	for _, tc := range testCases {
		moduleDir := filepath.Join(workspaceDir, tc.name)
		os.MkdirAll(moduleDir, 0755)

		// 初始化 git 仓库
		gitDir := filepath.Join(moduleDir, ".git")
		os.MkdirAll(gitDir, 0755)

		// 创建 git config 文件
		configContent := `[core]
	repositoryformatversion = 0
[remote "origin"]
	url = ` + tc.url + `
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "main"]
	remote = origin
	merge = refs/heads/main
`
		os.WriteFile(filepath.Join(gitDir, "config"), []byte(configContent), 0644)
	}

	// 创建非 git 目录（应该被忽略）
	nonGitDir := filepath.Join(workspaceDir, "not-a-repo")
	os.MkdirAll(nonGitDir, 0755)

	// 执行扫描
	modules, err := ScanWorkspace(context.Background(), workspaceDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证结果
	if len(modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(modules))
	}

	// 验证模块名称和 URL
	found := make(map[string]string)
	for _, m := range modules {
		found[m.Name] = m.URL
	}

	for _, tc := range testCases {
		url, ok := found[tc.wantName]
		if !ok {
			t.Errorf("expected to find module %s", tc.wantName)
			continue
		}
		if url != tc.url {
			t.Errorf("expected URL %s, got %s", tc.url, url)
		}
	}
}

// TestScanWorkspace_NotExist 测试 workspace 不存在
func TestScanWorkspace_NotExist(t *testing.T) {
	_, err := ScanWorkspace(context.Background(), "/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent workspace")
	}
}
