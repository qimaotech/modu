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

	if cfg.Workspace != "." {
		t.Errorf("expected workspace ., got %s", cfg.Workspace)
	}
	if cfg.WorktreeRoot != "../worktrees" {
		t.Errorf("expected worktree-root ../worktrees, got %s", cfg.WorktreeRoot)
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

// TestLoadConfig_EnvVarDollarFormat 测试 $VAR 格式环境变量
func TestLoadConfig_EnvVarDollarFormat(t *testing.T) {
	// 设置环境变量
	t.Setenv("MY_WORKSPACE", "/test/workspace")
	t.Setenv("MY_WORKTREE_ROOT", "/test/worktrees")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".modu.yaml")
	content := `workspace: $MY_WORKSPACE
worktree-root: $MY_WORKTREE_ROOT
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

	if cfg.Workspace != "/test/workspace" {
		t.Errorf("expected workspace /test/workspace, got %s", cfg.Workspace)
	}
	if cfg.WorktreeRoot != "/test/worktrees" {
		t.Errorf("expected worktree-root /test/worktrees, got %s", cfg.WorktreeRoot)
	}
}

// TestLoadConfig_EnvVarBraceFormat 测试 ${VAR} 格式环境变量
func TestLoadConfig_EnvVarBraceFormat(t *testing.T) {
	t.Setenv("MY_WORKSPACE", "/test/workspace")
	t.Setenv("MY_WORKTREE_ROOT", "/test/worktrees")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".modu.yaml")
	content := `workspace: ${MY_WORKSPACE}
worktree-root: ${MY_WORKTREE_ROOT}
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

	if cfg.Workspace != "/test/workspace" {
		t.Errorf("expected workspace /test/workspace, got %s", cfg.Workspace)
	}
	if cfg.WorktreeRoot != "/test/worktrees" {
		t.Errorf("expected worktree-root /test/worktrees, got %s", cfg.WorktreeRoot)
	}
}

// TestLoadConfig_EnvVarUndefined 测试环境变量未定义时报错
func TestLoadConfig_EnvVarUndefined(t *testing.T) {
	// 确保环境变量不存在
	os.Unsetenv("UNDEFINED_WORKSPACE_VAR")
	os.Unsetenv("UNDEFINED_WORKTREE_VAR")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".modu.yaml")
	content := `workspace: $UNDEFINED_WORKSPACE_VAR
worktree-root: /opt/worktrees
default-base: develop
modules:
  - name: test
    url: git@example.com:test.git
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("expected error for undefined environment variable")
	}
	if !contains(err.Error(), "undefined environment variable") {
		t.Errorf("expected error about undefined environment variable, got: %v", err)
	}
	if !contains(err.Error(), "workspace") {
		t.Errorf("expected error to mention 'workspace' field, got: %v", err)
	}
}

// TestLoadConfig_EnvVarMixed 测试路径包含环境变量
func TestLoadConfig_EnvVarMixed(t *testing.T) {
	t.Setenv("USER", "testuser")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".modu.yaml")
	content := `workspace: /home/$USER/workspace
worktree-root: /home/$USER/worktrees
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

	if cfg.Workspace != "/home/testuser/workspace" {
		t.Errorf("expected workspace /home/testuser/workspace, got %s", cfg.Workspace)
	}
	if cfg.WorktreeRoot != "/home/testuser/worktrees" {
		t.Errorf("expected worktree-root /home/testuser/worktrees, got %s", cfg.WorktreeRoot)
	}
}

// TestLoadConfig_NoEnvVar 测试无环境变量的配置正常工作
func TestLoadConfig_NoEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".modu.yaml")
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

	if cfg.Workspace != "/opt/workspace" {
		t.Errorf("expected workspace /opt/workspace, got %s", cfg.Workspace)
	}
	if cfg.WorktreeRoot != "/opt/worktrees" {
		t.Errorf("expected worktree-root /opt/worktrees, got %s", cfg.WorktreeRoot)
	}
}

// TestLoadConfigForScan 测试加载配置用于scan命令
func TestLoadConfigForScan(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".modu.yaml")
	content := `workspace: /opt/workspace
worktree-root: /opt/worktrees
default-base: develop
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfigForScan(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Workspace != "/opt/workspace" {
		t.Errorf("expected workspace /opt/workspace, got %s", cfg.Workspace)
	}
	if len(cfg.Modules) != 0 {
		t.Errorf("expected 0 modules, got %d", len(cfg.Modules))
	}
}

// TestLoadConfigForScan_ValidationError 测试加载配置跳过模块但仍验证基础字段
func TestLoadConfigForScan_ValidationError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".modu.yaml")
	content := `workspace: /opt/workspace
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadConfigForScan(configPath)
	if err == nil {
		t.Error("expected error for missing worktree-root and default-base")
	}
}

// TestResolveEnvVars 测试环境变量解析
func TestResolveEnvVars(t *testing.T) {
	t.Run("无环境变量", func(t *testing.T) {
		value := "/path/to/dir"
		err := resolveEnvVars(&value, "test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if value != "/path/to/dir" {
			t.Errorf("expected /path/to/dir, got %s", value)
		}
	})

	t.Run("环境变量已定义", func(t *testing.T) {
		os.Setenv("TEST_MODU_PATH", "/test/path")
		defer os.Unsetenv("TEST_MODU_PATH")

		value := "$TEST_MODU_PATH/subdir"
		err := resolveEnvVars(&value, "test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if value != "/test/path/subdir" {
			t.Errorf("expected /test/path/subdir, got %s", value)
		}
	})

	t.Run("环境变量未定义", func(t *testing.T) {
		os.Unsetenv("UNDEFINED_VAR")
		value := "$UNDEFINED_VAR/subdir"
		err := resolveEnvVars(&value, "test")
		if err == nil {
			t.Error("expected error for undefined environment variable")
		}
	})

	t.Run("环境变量使用花括号语法", func(t *testing.T) {
		os.Setenv("TEST_HOOK_PATH", "/hooks")
		defer os.Unsetenv("TEST_HOOK_PATH")

		value := "${TEST_HOOK_PATH}/scripts"
		err := resolveEnvVars(&value, "test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if value != "/hooks/scripts" {
			t.Errorf("expected /hooks/scripts, got %s", value)
		}
	})
}

// TestExtractUndefinedVars 测试提取未定义的环境变量
func TestExtractUndefinedVars(t *testing.T) {
	t.Run("提取未定义的环境变量", func(t *testing.T) {
		os.Setenv("DEFINED_VAR", "value")
		defer os.Unsetenv("DEFINED_VAR")

		undefined := extractUndefinedVars("$DEFINED_VAR/${OTHER_UNDEFINED}")
		if len(undefined) != 1 {
			t.Errorf("expected 1 undefined var, got %d", len(undefined))
		}
	})

	t.Run("无环境变量", func(t *testing.T) {
		undefined := extractUndefinedVars("/path/to/dir")
		if len(undefined) != 0 {
			t.Errorf("expected 0 undefined vars, got %d", len(undefined))
		}
	})

	t.Run("多个未定义变量只返回一次", func(t *testing.T) {
		os.Unsetenv("VAR1")
		os.Unsetenv("VAR2")
		undefined := extractUndefinedVars("$VAR1 $VAR2 $VAR1")
		if len(undefined) != 2 {
			t.Errorf("expected 2 unique undefined vars, got %d", len(undefined))
		}
	})
}

// TestValidate 测试配置验证
func TestValidate(t *testing.T) {
	t.Run("有效配置", func(t *testing.T) {
		cfg := &Config{
			Workspace:    "/workspace",
			WorktreeRoot: "/worktrees",
			DefaultBase:  "develop",
			Modules:      []Module{{Name: "m1", URL: "url"}},
		}
		err := validate(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("workspace为空", func(t *testing.T) {
		cfg := &Config{
			WorktreeRoot: "/worktrees",
			DefaultBase:  "develop",
			Modules:      []Module{{Name: "m1", URL: "url"}},
		}
		err := validate(cfg)
		if err == nil {
			t.Error("expected error for empty workspace")
		}
	})

	t.Run("worktreeRoot为空", func(t *testing.T) {
		cfg := &Config{
			Workspace:   "/workspace",
			DefaultBase: "develop",
			Modules:     []Module{{Name: "m1", URL: "url"}},
		}
		err := validate(cfg)
		if err == nil {
			t.Error("expected error for empty worktreeRoot")
		}
	})

	t.Run("defaultBase为空", func(t *testing.T) {
		cfg := &Config{
			Workspace:    "/workspace",
			WorktreeRoot: "/worktrees",
			Modules:      []Module{{Name: "m1", URL: "url"}},
		}
		err := validate(cfg)
		if err == nil {
			t.Error("expected error for empty defaultBase")
		}
	})

	t.Run("modules为空", func(t *testing.T) {
		cfg := &Config{
			Workspace:    "/workspace",
			WorktreeRoot: "/worktrees",
			DefaultBase:  "develop",
			Modules:      []Module{},
		}
		err := validate(cfg)
		if err == nil {
			t.Error("expected error for empty modules")
		}
	})
}

// TestValidateBasic 测试基础配置验证
func TestValidateBasic(t *testing.T) {
	t.Run("有效配置", func(t *testing.T) {
		cfg := &Config{
			Workspace:    "/workspace",
			WorktreeRoot: "/worktrees",
			DefaultBase:  "develop",
		}
		err := validateBasic(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("缺少workspace", func(t *testing.T) {
		cfg := &Config{
			WorktreeRoot: "/worktrees",
			DefaultBase:  "develop",
		}
		err := validateBasic(cfg)
		if err == nil {
			t.Error("expected error for missing workspace")
		}
	})

	t.Run("缺少worktreeRoot", func(t *testing.T) {
		cfg := &Config{
			Workspace:   "/workspace",
			DefaultBase: "develop",
		}
		err := validateBasic(cfg)
		if err == nil {
			t.Error("expected error for missing worktreeRoot")
		}
	})

	t.Run("缺少defaultBase", func(t *testing.T) {
		cfg := &Config{
			Workspace:    "/workspace",
			WorktreeRoot: "/worktrees",
		}
		err := validateBasic(cfg)
		if err == nil {
			t.Error("expected error for missing defaultBase")
		}
	})
}

// TestReadGitRemoteURL 测试读取git远程URL
func TestReadGitRemoteURL(t *testing.T) {
	t.Run("读取有效的git配置", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitConfig := filepath.Join(tmpDir, "config")
		content := `[core]
	repositoryformatversion = 0
[remote "origin"]
	url = https://github.com/example/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`
		err := os.WriteFile(gitConfig, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}

		url, err := readGitRemoteURL(gitConfig)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if url != "https://github.com/example/repo.git" {
			t.Errorf("expected URL https://github.com/example/repo.git, got %s", url)
		}
	})

	t.Run("无remote origin", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitConfig := filepath.Join(tmpDir, "config")
		content := `[core]
	repositoryformatversion = 0
`
		err := os.WriteFile(gitConfig, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}

		_, err = readGitRemoteURL(gitConfig)
		if err == nil {
			t.Error("expected error for missing remote origin")
		}
	})

	t.Run("文件不存在", func(t *testing.T) {
		_, err := readGitRemoteURL("/path/that/does/not/exist")
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})
}

// TestUpdateGitignore 测试更新gitignore
func TestUpdateGitignore(t *testing.T) {
	t.Run("创建新的gitignore", func(t *testing.T) {
		tmpDir := t.TempDir()

		modules := []Module{
			{Name: "module1"},
			{Name: "module2"},
		}

		err := UpdateGitignore(tmpDir, modules)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
		if err != nil {
			t.Fatalf("failed to read .gitignore: %v", err)
		}
		if !contains(string(content), "module1/") {
			t.Error("expected .gitignore to contain module1/")
		}
		if !contains(string(content), "module2/") {
			t.Error("expected .gitignore to contain module2/")
		}
	})

	t.Run("追加到已存在的gitignore", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")

		existing := "# existing\ngo/\n"
		os.WriteFile(gitignorePath, []byte(existing), 0644)

		modules := []Module{{Name: "module1"}}

		err := UpdateGitignore(tmpDir, modules)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		content, _ := os.ReadFile(gitignorePath)
		if !contains(string(content), "# existing") {
			t.Error("expected .gitignore to preserve existing content")
		}
		if !contains(string(content), "module1/") {
			t.Error("expected .gitignore to contain module1/")
		}
	})

	t.Run("去重已有模块", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")

		existing := "module1/\nmodule2/\n"
		os.WriteFile(gitignorePath, []byte(existing), 0644)

		modules := []Module{
			{Name: "module1"},
			{Name: "module3"},
		}

		err := UpdateGitignore(tmpDir, modules)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		content, _ := os.ReadFile(gitignorePath)
		// module1 应该只出现一次
		count := countOccurrences(string(content), "module1/")
		if count != 1 {
			t.Errorf("expected module1/ to appear once, got %d", count)
		}
		// module3 应该出现一次
		count = countOccurrences(string(content), "module3/")
		if count != 1 {
			t.Errorf("expected module3/ to appear once, got %d", count)
		}
	})
}

// TestSaveConfig_AutoCreateWorktreeDir 测试保存配置时自动创建worktree目录
func TestSaveConfig_AutoCreateWorktreeDir(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", ".modu.yaml")
	worktreesDir := filepath.Join(tmpDir, "subdir", "worktrees")

	os.MkdirAll(filepath.Dir(configPath), 0755)

	cfg := &Config{
		Workspace:    tmpDir,
		WorktreeRoot: "worktrees",
		DefaultBase:  "develop",
	}

	err := SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证 worktree 目录已创建
	if _, err := os.Stat(worktreesDir); os.IsNotExist(err) {
		t.Error("expected worktree directory to be created")
	}
}

// 辅助函数
func countOccurrences(s, substr string) int {
	count := 0
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			count++
		}
	}
	return count
}
